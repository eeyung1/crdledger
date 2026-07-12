package config

import (
	"os"
	"strings"
)

// Config holds all environment-derived settings. Read once, inside main(),
// never as package-level vars initialized from os.Getenv() — that captures
// an empty value if .env hasn't loaded yet.
type Config struct {
	Port           string
	DBPath         string
	SessionSecret  string
	SecureCookies  bool // true when the app is served over HTTPS
	Environment    string
	AdminUsernames map[string]bool // usernames allowed to reset other users' passwords
}

func Load() (Config, error) {
	secret, ok := os.LookupEnv("SESSION_SECRET")
	if !ok || secret == "" {
		return Config{}, errMissingSecret
	}

	cfg := Config{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "./crdledger.db"),
		SessionSecret:  secret,
		SecureCookies:  getEnv("SECURE_COOKIES", "false") == "true",
		Environment:    getEnv("ENVIRONMENT", "development"),
		AdminUsernames: parseAdminUsernames(getEnv("ADMIN_USERNAMES", "")),
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

var errMissingSecret = configError("SESSION_SECRET environment variable is required but not set")

type configError string

func (e configError) Error() string { return string(e) }

// parseAdminUsernames turns a comma-separated env var like
// "alice,bob, carol" into a lookup set. Empty/blank entries are ignored.
func parseAdminUsernames(raw string) map[string]bool {
	set := make(map[string]bool)
	for _, name := range strings.Split(raw, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			set[name] = true
		}
	}
	return set
}
