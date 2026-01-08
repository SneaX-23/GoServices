package config

import "os"

type EmailConfig struct {
	ApiKey string
	From   string
}

func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		ApiKey: os.Getenv("RESEND_API_KEY"),
		From:   getEnv("EMAIL_FROM", "Email service <no-reply@sneax.quest>"),
	}
}

// Look up if EMAIL_FROM exist
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
