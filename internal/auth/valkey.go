package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

// ValkeyService wraps the Valkey client with auth-specific operations.
type ValkeyService struct {
	client valkey.Client
}

func NewValkeyService(url string) (*ValkeyService, error) {
	client, err := valkey.NewClient(valkey.MustParseURL(url))
	if err != nil {
		return nil, fmt.Errorf("valkey connect: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		return nil, fmt.Errorf("valkey ping: %w", err)
	}

	return &ValkeyService{client: client}, nil
}

func (v *ValkeyService) Close() {
	v.client.Close()
}

// RevokeToken adds a token JTI to the blacklist. TTL matches token expiry
// so Valkey cleans it up automatically.
func (v *ValkeyService) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return v.client.Do(ctx,
		v.client.B().Set().Key(blacklistKey(jti)).Value("1").Exat(expiresAt).Build(),
	).Error()
}

// IsRevoked returns true if the token JTI is on the blacklist.
func (v *ValkeyService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	res := v.client.Do(ctx,
		v.client.B().Exists().Key(blacklistKey(jti)).Build(),
	)
	if err := res.Error(); err != nil {
		return false, fmt.Errorf("check revoked: %w", err)
	}
	n, err := res.AsInt64()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CacheSet stores a value with a TTL.
func (v *ValkeyService) CacheSet(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return v.client.Do(ctx,
		v.client.B().Set().Key(cacheKey(key)).Value(string(value)).Ex(ttl).Build(),
	).Error()
}

// CacheGet retrieves a cached value. Returns nil, nil on a cache miss.
func (v *ValkeyService) CacheGet(ctx context.Context, key string) ([]byte, error) {
	res := v.client.Do(ctx,
		v.client.B().Get().Key(cacheKey(key)).Build(),
	)
	if res.Error() != nil {
		if valkey.IsValkeyNil(res.Error()) {
			return nil, nil
		}
		return nil, fmt.Errorf("cache get: %w", res.Error())
	}
	b, err := res.AsBytes()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// CacheDel invalidates a cache key.
func (v *ValkeyService) CacheDel(ctx context.Context, key string) error {
	return v.client.Do(ctx,
		v.client.B().Del().Key(cacheKey(key)).Build(),
	).Error()
}

// TrackFailedLogin increments the failed login counter for a username.
// Returns the new count. Uses two separate commands instead of a pipeline
// since valkey-go pipelines work differently per client implementation.
func (v *ValkeyService) TrackFailedLogin(ctx context.Context, username string, window time.Duration) (int64, error) {
	key := failedLoginKey(username)

	// Increment
	incrRes := v.client.Do(ctx, v.client.B().Incr().Key(key).Build())
	if err := incrRes.Error(); err != nil {
		return 0, fmt.Errorf("track failed login incr: %w", err)
	}
	count, err := incrRes.AsInt64()
	if err != nil {
		return 0, err
	}

	// Set expiry only on the first increment so the window doesn't reset
	// on every failed attempt
	if count == 1 {
		_ = v.client.Do(ctx,
			v.client.B().Expire().Key(key).Seconds(int64(window.Seconds())).Build(),
		).Error()
	}

	return count, nil
}

// ClearFailedLogins resets the failed login counter on successful login.
func (v *ValkeyService) ClearFailedLogins(ctx context.Context, username string) error {
	return v.client.Do(ctx,
		v.client.B().Del().Key(failedLoginKey(username)).Build(),
	).Error()
}

// ── Key helpers ───────────────────────────────────────────────────────────────

func blacklistKey(jti string) string { return "bl:" + jti }
func cacheKey(key string) string     { return "cache:" + key }
func failedLoginKey(u string) string { return "fail:" + u }
