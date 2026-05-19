package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	LDAP        LDAPConfig
	JWT         JWTConfig
	Valkey      ValkeyConfig
	DevAccounts DevAccountsConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

func (s ServerConfig) IsDevelopment() bool {
	return s.Env == "development"
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string
}

type LDAPConfig struct {
	URL        string
	UserDomain string
	BaseDN     string
	UserFilter string
	UseSSL     bool
	StartTLS   bool
	SkipVerify bool
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
	Issuer string
}

type ValkeyConfig struct {
	URL               string
	MaxFailedLogins   int           // lock after N failures (0 = disabled)
	FailedLoginWindow time.Duration // window for counting failures
	DashboardCacheTTL time.Duration // how long to cache dashboard responses
}

// DevAccount holds a local credential pair for development logins.
// Mirrors auth.DevAccount — kept separate to avoid an import cycle.
type DevAccount struct {
	Username     string
	DisplayName  string
	Email        string
	PasswordHash string
}

// DevAccountsConfig holds local accounts used in development.
// Only consulted when APP_ENV=development.
type DevAccountsConfig struct {
	Accounts []DevAccount
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "9432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	dbCfg := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "supabase_admin"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "ecg"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
	dbCfg.DSN = fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, dbCfg.Password,
		dbCfg.Host, dbCfg.Port,
		dbCfg.DBName, dbCfg.SSLMode,
	)

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "8h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}
	if getEnv("JWT_SECRET", "") == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	failedLoginWindow, _ := time.ParseDuration(getEnv("VALKEY_FAILED_LOGIN_WINDOW", "15m"))
	maxFailed, _ := strconv.Atoi(getEnv("VALKEY_MAX_FAILED_LOGINS", "5"))
	dashCacheTTL, _ := time.ParseDuration(getEnv("VALKEY_DASHBOARD_CACHE_TTL", "60s"))

	// Dev accounts — loaded from env, only used in development
	devAccounts := loadDevAccounts()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "9400"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: dbCfg,
		LDAP: LDAPConfig{
			URL:        getEnv("LDAP_URL", "ldap://localhost:389"),
			UserDomain: getEnv("LDAP_USER_DOMAIN", ""),
			BaseDN:     getEnv("LDAP_BASE_DN", ""),
			UserFilter: getEnv("LDAP_USER_FILTER", "(sAMAccountName=%s)"),
			UseSSL:     getEnv("LDAP_USE_SSL", "false") == "true",
			StartTLS:   getEnv("LDAP_STARTTLS", "false") == "true",
			SkipVerify: getEnv("LDAP_SKIP_VERIFY", "false") == "true",
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: jwtExpiry,
			Issuer: getEnv("JWT_ISSUER", "github.com/kwaabs/m-events"),
		},
		Valkey: ValkeyConfig{
			URL:               getEnv("VALKEY_URL", "redis://localhost:9479"),
			MaxFailedLogins:   maxFailed,
			FailedLoginWindow: failedLoginWindow,
			DashboardCacheTTL: dashCacheTTL,
		},
		DevAccounts: devAccounts,
	}, nil
}

// loadDevAccounts reads dev account credentials from env vars.
//
// Format in .env:
//
//	DEV_ACCOUNT_1_USERNAME=devuser
//	DEV_ACCOUNT_1_DISPLAY_NAME=Dev User
//	DEV_ACCOUNT_1_EMAIL=dev@example.com
//	DEV_ACCOUNT_1_PASSWORD_HASH=$2a$10$...   (bcrypt hash)
//
// Add more accounts with _2_, _3_, etc.
func loadDevAccounts() DevAccountsConfig {
	var accounts []DevAccount
	for i := 1; i <= 10; i++ {
		prefix := fmt.Sprintf("DEV_ACCOUNT_%d_", i)
		username := getEnv(prefix+"USERNAME", "")
		if username == "" {
			continue
		}
		accounts = append(accounts, DevAccount{
			Username:     username,
			DisplayName:  getEnv(prefix+"DISPLAY_NAME", username),
			Email:        getEnv(prefix+"EMAIL", ""),
			PasswordHash: getEnv(prefix+"PASSWORD_HASH", ""),
		})
	}
	return DevAccountsConfig{Accounts: accounts}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
