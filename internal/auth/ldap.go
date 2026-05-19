package auth

import (
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"

	"github.com/kwaabs/m-events/internal/config"
)

// LDAPUser holds the attributes pulled from the directory after a successful bind.
type LDAPUser struct {
	Username    string
	DisplayName string
	Email       string
	Groups      []string
}

// LDAPService authenticates users directly against an LDAP / Active Directory
// server with no service/admin account required.
type LDAPService struct {
	cfg config.LDAPConfig
}

func NewLDAPService(cfg config.LDAPConfig) *LDAPService {
	return &LDAPService{cfg: cfg}
}

// Authenticate binds directly as the user (UPN: username@domain), which both
// proves the password and opens an authenticated connection. It then does a
// self-search to retrieve display name, email, and group membership.
func (s *LDAPService) Authenticate(username, password string) (*LDAPUser, error) {
	conn, err := s.dial()
	if err != nil {
		return nil, fmt.Errorf("ldap dial: %w", err)
	}
	defer conn.Close()

	bindIdentity := username
	if s.cfg.UserDomain != "" {
		bindIdentity = username + "@" + s.cfg.UserDomain
	}

	if err := conn.Bind(bindIdentity, password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("ldap bind: %w", err)
	}

	// Self-search for display attributes — works because the user can read
	// their own entry. Failure here is non-fatal; we already know auth passed.
	if s.cfg.BaseDN != "" {
		filter := fmt.Sprintf(s.cfg.UserFilter, ldap.EscapeFilter(username))
		searchReq := ldap.NewSearchRequest(
			s.cfg.BaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0, 0, false,
			filter,
			[]string{"cn", "displayName", "mail", "memberOf", "sAMAccountName"},
			nil,
		)
		if result, err := conn.Search(searchReq); err == nil && len(result.Entries) > 0 {
			entry := result.Entries[0]
			user := &LDAPUser{
				Username:    entry.GetAttributeValue("sAMAccountName"),
				DisplayName: entry.GetAttributeValue("displayName"),
				Email:       entry.GetAttributeValue("mail"),
				Groups:      entry.GetAttributeValues("memberOf"),
			}
			if user.DisplayName == "" {
				user.DisplayName = entry.GetAttributeValue("cn")
			}
			if user.Username == "" {
				user.Username = username
			}
			return user, nil
		}
	}

	// Bind succeeded but attribute fetch skipped or failed — return minimal user
	return &LDAPUser{Username: username}, nil
}

// dial opens a connection to the LDAP server.
//
// Behaviour by config:
//   - LDAP_USE_SSL=true  → ldaps:// (TLS from the start, port 636)
//   - LDAP_STARTTLS=true → plain dial then upgrade via StartTLS (port 389)
//   - neither            → plain ldap:// with no encryption (internal networks)
func (s *LDAPService) dial() (*ldap.Conn, error) {
	if s.cfg.UseSSL {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: s.cfg.SkipVerify, //nolint:gosec
			ServerName:         ldapHost(s.cfg.URL),
		}
		return ldap.DialURL(s.cfg.URL, ldap.DialWithTLSConfig(tlsCfg))
	}

	// Plain dial first (works for ldap:// on port 389)
	conn, err := ldap.DialURL(s.cfg.URL)
	if err != nil {
		return nil, err
	}

	// Only attempt StartTLS when explicitly requested
	if s.cfg.StartTLS {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: s.cfg.SkipVerify, //nolint:gosec
			ServerName:         ldapHost(s.cfg.URL),
		}
		if err := conn.StartTLS(tlsCfg); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ldap StartTLS: %w", err)
		}
	}

	return conn, nil
}

// ldapHost extracts the bare hostname from an ldap:// or ldaps:// URL.
func ldapHost(url string) string {
	host := url
	for _, prefix := range []string{"ldaps://", "ldap://"} {
		if len(host) > len(prefix) {
			host = host[len(prefix):]
		}
	}
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

var ErrInvalidCredentials = fmt.Errorf("invalid username or password")
