package config

import (
	"fmt"
	"hublio/internal/platform/env"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	ServerAddress      string
	DB                 DatabaseConfig
	MailProviderType   string
	MailProviderConfig map[string]any
	WebAppURL          string
	// MFAEncryptionKey protects TOTP secrets at rest (32 bytes: raw, 64 hex chars, or base64).
	// Falls back to CREDENTIAL_ENCRYPTION_KEY so a single key can cover both in development.
	MFAEncryptionKey string
}

func NewConfig() *Config {
	serverAddress := ":" + env.GetEnv("SERVER_PORT", "8080")

	mailProviderConfig := make(map[string]any)

	mailProviderType := env.GetEnv("MAIL_PROVIDER_TYPE", "mailtrap")

	if mailProviderType == "mailtrap" {
		mailTrapConfig := map[string]any{
			"mail_sender": env.GetEnv("MAILTRAP_MAIL_SENDER", ""),
			"name_sender": env.GetEnv("MAILTRAP_NAME_SENDER", ""),
			"api_url":     env.GetEnv("MAILTRAP_API_URL", ""),
			"api_key":     env.GetEnv("MAILTRAP_API_KEY", ""),
		}

		mailProviderConfig["mailtrap"] = mailTrapConfig
	}

	return &Config{
		ServerAddress: serverAddress,

		DB: DatabaseConfig{
			Host:     env.GetEnv("DB_HOST", "localhost"),
			Port:     env.GetEnv("DB_PORT", "5432"),
			User:     env.GetEnv("DB_USER", "postgres"),
			Password: env.GetEnv("DB_PASSWORD", "postgres"),
			DBName:   env.GetEnv("DB_NAME", "myapp"),
			SSLMode:  env.GetEnv("DB_SSLMODE", "disable"),
		},

		MailProviderType:   mailProviderType,
		MailProviderConfig: mailProviderConfig,
		WebAppURL:          env.GetEnv("APP_WEB_URL", "http://localhost:3000"),
		MFAEncryptionKey:   env.GetEnv("MFA_ENCRYPTION_KEY", env.GetEnv("CREDENTIAL_ENCRYPTION_KEY", "")),
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
