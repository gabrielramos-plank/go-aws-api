package config

import (
	"errors"
	"net"
	"net/url"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables. DATABASE_URL wins if set;
// otherwise the URL is built from DB_* variables, which is how ECS injects them
// (DB_PASSWORD comes from SSM Parameter Store as a secret).
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL != "" {
		return cfg, nil
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		return Config{}, errors.New("either DATABASE_URL or DB_HOST must be set")
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getenv("DB_USER", "postgres"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(host, getenv("DB_PORT", "5432")),
		Path:     "/" + getenv("DB_NAME", "app"),
		RawQuery: url.Values{"sslmode": {getenv("DB_SSLMODE", "require")}}.Encode(),
	}
	cfg.DatabaseURL = u.String()
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
