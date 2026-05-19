package auth

import "context"

type contextKey string

const claimsKey contextKey = "jwt_claims"

// WithClaims stores validated JWT claims in the request context.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext retrieves JWT claims stored by the auth middleware.
// Returns nil if the context carries no claims (unauthenticated request).
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}
