package config

import (
	"ims-message-util/internal/utils"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Application Application
	Email       Email
}

type Email struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

type Application struct {
	Port int
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env", "Error", err)
	}

	smtpPort, err := utils.ConvertStringToInteger(getEnv("SMTP_PORT", "465"))

	if err != nil {
		slog.Warn("Error in coversion for smtpPort, falling back to default", "Error", err)
		smtpPort = 465
	}

	port, err := utils.ConvertStringToInteger(getEnv("PORT", "8081"))

	if err != nil {
		slog.Warn("Error in coversion for port, falling back to default", "Error", err)
		port = 8081
	}

	return &Config{
		Application: Application{
			Port: port,
		},
		Email: Email{
			SMTPHost: getEnv("SMTP_HOST", "smtp.hostinger.com"),
			SMTPPort: smtpPort,
			Username: getEnv("EMAIL_USERNAME", "<username>"),
			Password: getEnv("EMAIL_PASSWORD", "<password>"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
