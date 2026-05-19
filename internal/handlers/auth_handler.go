package handlers

import (
	"errors"
	"net/http"

	"github.com/kwaabs/m-events/internal/auth"
	"github.com/kwaabs/m-events/internal/config"
)

type AuthHandler struct {
	ldap   *auth.LDAPService
	dev    *auth.DevAuthService // nil in production
	jwt    *auth.JWTService
	valkey *auth.ValkeyService
	vCfg   config.ValkeyConfig
	isDev  bool
}

func NewAuthHandler(
	ldap *auth.LDAPService,
	dev *auth.DevAuthService,
	jwt *auth.JWTService,
	valkey *auth.ValkeyService,
	vCfg config.ValkeyConfig,
	isDev bool,
) *AuthHandler {
	return &AuthHandler{
		ldap:   ldap,
		dev:    dev,
		jwt:    jwt,
		valkey: valkey,
		vCfg:   vCfg,
		isDev:  isDev,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AuthMethod  string `json:"auth_method"` // "ldap" or "dev"
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required", nil)
		return
	}

	// ── Rate limiting ─────────────────────────────────────────────────────────
	if h.vCfg.MaxFailedLogins > 0 {
		count, err := h.valkey.TrackFailedLogin(r.Context(), req.Username, h.vCfg.FailedLoginWindow)
		if err == nil && count > int64(h.vCfg.MaxFailedLogins) {
			writeError(w, http.StatusTooManyRequests,
				"account temporarily locked — too many failed attempts", nil)
			return
		}
	}

	// ── Authenticate ──────────────────────────────────────────────────────────
	var user *auth.LDAPUser
	authMethod := "ldap"

	// Try dev accounts first if in development and the service is available
	if h.isDev && h.dev != nil {
		u, err := h.dev.Authenticate(req.Username, req.Password)
		if err == nil {
			user = u
			authMethod = "dev"
		}
	}

	// Fall through to LDAP if dev auth didn't match
	if user == nil {
		u, err := h.ldap.Authenticate(req.Username, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "invalid username or password", nil)
				return
			}
			writeError(w, http.StatusInternalServerError, "authentication error", err)
			return
		}
		user = u
	}

	// Clear failed login counter on success
	if h.vCfg.MaxFailedLogins > 0 {
		_ = h.valkey.ClearFailedLogins(r.Context(), req.Username)
	}

	// ── Issue token ───────────────────────────────────────────────────────────
	token, expiry, err := h.jwt.Issue(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token", err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token:       token,
		ExpiresAt:   expiry.UTC().Format("2006-01-02T15:04:05Z"),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AuthMethod:  authMethod,
	})
}

// POST /api/v1/auth/logout
// Revokes the current token by adding its JTI to the Valkey blacklist.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated", nil)
		return
	}

	jti := claims.ID
	if jti == "" {
		// Old token with no JTI — nothing to revoke, just tell client to discard
		writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	if err := h.valkey.RevokeToken(r.Context(), jti, claims.ExpiresAt.Time); err != nil {
		writeError(w, http.StatusInternalServerError, "could not revoke token", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"username":     claims.Username,
		"display_name": claims.DisplayName,
		"email":        claims.Email,
		"groups":       claims.Groups,
		"expires_at":   claims.ExpiresAt.Time.UTC().Format("2006-01-02T15:04:05Z"),
	})
}
