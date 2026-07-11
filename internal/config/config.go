package config

import "os"

// Config holds all environment-derived settings. Read once, inside main(),
// never as package-level vars initialized from os.Getenv() — that captures
// an empty value if .env hasn't loaded yet.
type Config struct {
	Port          string
	DBPath        string
	SessionSecret string
	SecureCookies bool // true when the app is served over HTTPS
	Environment   string
}

func Load() (Config, error) {
	secret, ok := os.LookupEnv("SESSION_SECRET")
	if !ok || secret == "" {
		return Config{}, errMissingSecret
	}

	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./crdledger.db"),
		SessionSecret: secret,
		SecureCookies: getEnv("SECURE_COOKIES", "false") == "true",
		Environment:   getEnv("ENVIRONMENT", "development"),
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
