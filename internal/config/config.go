package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		slog.Info("Error loading .env", "Error", err)
	}

	return &Config{
		SMTPHost: getEnv("SMTP_HOST", "smtp.hostinger.com"),
		SMTPPort: getEnv("SMTP_PORT", "465"),
		Username: getEnv("EMAIL_USERNAME", "<username>"),
		Password: getEnv("EMAIL_PASSWORD", "<password>"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
