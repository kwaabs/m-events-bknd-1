package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/kwaabs/m-events/internal/config"
)

// Claims is the JWT payload we issue and validate.
type Claims struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	jwt.RegisteredClaims
}

// JWTService issues and validates signed JWTs.
type JWTService struct {
	cfg config.JWTConfig
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

// Issue creates a signed JWT for the given LDAP user.
// Each token gets a unique JTI (JWT ID) so it can be individually revoked.
func (s *JWTService) Issue(user *LDAPUser) (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(s.cfg.Expiry)
	jti := uuid.New().String()

	claims := Claims{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Groups:      user.Groups,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   user.Username,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	return signed, expiry, nil
}

// Validate parses and verifies a JWT string, returning the embedded claims.
func (s *JWTService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(s.cfg.Secret), nil
		},
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

var (
	ErrTokenExpired = fmt.Errorf("token has expired")
	ErrTokenInvalid = fmt.Errorf("token is invalid")
)
