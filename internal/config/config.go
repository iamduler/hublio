package config

import (
	"fmt"
	"shopping-cart/internal/utils"
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
}

func NewConfig() *Config {
	serverAddress := ":" + utils.GetEnv("SERVER_PORT", "")
	fmt.Println("Server address:", serverAddress)

	mailProviderConfig := make(map[string]any)

	mailProviderType := utils.GetEnv("MAIL_PROVIDER_TYPE", "mailtrap")

	if mailProviderType == "mailtrap" {
		mailTrapConfig := map[string]any{
			"mail_sender": utils.GetEnv("MAILTRAP_MAIL_SENDER", ""),
			"name_sender": utils.GetEnv("MAILTRAP_NAME_SENDER", ""),
			"api_url":     utils.GetEnv("MAILTRAP_API_URL", ""),
			"api_key":     utils.GetEnv("MAILTRAP_API_KEY", ""),
		}

		mailProviderConfig["mailtrap"] = mailTrapConfig
	}

	return &Config{
		ServerAddress: serverAddress,

		DB: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5432"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			DBName:   utils.GetEnv("DB_NAME", "myapp"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},

		MailProviderType:   mailProviderType,
		MailProviderConfig: mailProviderConfig,
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
