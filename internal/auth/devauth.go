package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DevAccount is a local credential pair used in development.
// Defined here to avoid an import cycle — config imports auth for LDAPUser,
// so auth cannot import config.
type DevAccount struct {
	Username     string
	DisplayName  string
	Email        string
	PasswordHash string // bcrypt hash — never plaintext
}

// DevAuthService authenticates against a set of local accounts.
// Only enabled when APP_ENV=development.
type DevAuthService struct {
	accounts map[string]DevAccount
}

// NewDevAuthService accepts a plain slice so neither auth nor config
// needs to import the other just for this constructor.
func NewDevAuthService(accounts []DevAccount) *DevAuthService {
	m := make(map[string]DevAccount, len(accounts))
	for _, a := range accounts {
		m[a.Username] = a
	}
	return &DevAuthService{accounts: m}
}

// Authenticate checks username + password against the local dev accounts.
func (s *DevAuthService) Authenticate(username, password string) (*LDAPUser, error) {
	account, ok := s.accounts[username]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return &LDAPUser{
		Username:    account.Username,
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Groups:      []string{"dev"},
	}, nil
}

// HashPassword generates a bcrypt hash. Used by tools/hashpw.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}
