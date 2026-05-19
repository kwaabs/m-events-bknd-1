package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/kwaabs/m-events/internal/auth"
)

// Authenticate validates the Bearer JWT and checks it against the Valkey
// token blacklist. Stores parsed claims in the request context.
func Authenticate(jwtSvc *auth.JWTService, valkey *auth.ValkeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearer(r)
			if tokenStr == "" {
				writeUnauthorized(w, "missing or malformed Authorization header")
				return
			}

			claims, err := jwtSvc.Validate(tokenStr)
			if err != nil {
				if errors.Is(err, auth.ErrTokenExpired) {
					writeUnauthorized(w, "token has expired")
					return
				}
				writeUnauthorized(w, "invalid token")
				return
			}

			// Check blacklist — catches logged-out tokens that haven't expired yet
			if claims.ID != "" {
				revoked, err := valkey.IsRevoked(r.Context(), claims.ID)
				if err != nil {
					// Valkey unavailable — fail open with a warning rather than
					// locking all users out. Swap to fail-closed if preferred.
					// log warning here if needed
					_ = err
				} else if revoked {
					writeUnauthorized(w, "token has been revoked")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}

func extractBearer(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="github.com/kwaabs/m-events"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
