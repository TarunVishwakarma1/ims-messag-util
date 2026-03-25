package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env", "Error", err)
	}

	port, err := strconv.Atoi(getEnv("SMTP_PORT", "465"))

	if err != nil {
		slog.Warn("Error in coversion for port, falling back to default", "Error", err)
		port = 465
	}

	return &Config{
		SMTPHost: getEnv("SMTP_HOST", "smtp.hostinger.com"),
		SMTPPort: port,
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
