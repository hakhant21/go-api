package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	JWTExpiry      time.Duration
	RefreshTTL     time.Duration
	AppBaseURL     string
	AllowedOrigins []string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	BootstrapAdmin string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	expiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	refresh, err := strconv.Atoi(getEnv("REFRESH_TTL_HOURS", "720"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TTL_HOURS: %w", err)
	}

	cfg := &Config{
		Env:            getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiry:      time.Duration(expiry) * time.Hour,
		RefreshTTL:     time.Duration(refresh) * time.Hour,
		AppBaseURL:     getEnv("APP_BASE_URL", "http://localhost:3000"),
		AllowedOrigins: splitCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUser:       getEnv("SMTP_USER", ""),
		SMTPPass:       getEnv("SMTP_PASS", ""),
		SMTPFrom:       getEnv("SMTP_FROM", "no-reply@example.com"),
		BootstrapAdmin: getEnv("BOOTSTRAP_ADMIN_EMAIL", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
