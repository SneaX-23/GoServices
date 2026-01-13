package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:              getEnv("DB_HOST", "localhost"),
		Port:              mustInt("DB_PORT"),
		User:              mustEnv("DB_USER"),
		Password:          mustEnv("DB_PASSWORD"),
		Name:              mustEnv("DB_NAME"),
		SSLMode:           getEnv("DB_SSLMODE", "disable"),
		MaxConns:          int32(getEnvInt("DB_MAX_CONNS", 20)),
		MinConns:          int32(getEnvInt("DB_MIN_CONNS", 5)),
		MaxConnLifetime:   getEnvDuration("DB_MAX_CONN_LIFETIME", 1*time.Hour),
		MaxConnIdleTime:   getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		HealthCheckPeriod: getEnvDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
	}
}

// Helpers for strict loading
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return v
}

func mustInt(key string) int {
	v, err := strconv.Atoi(mustEnv(key))
	if err != nil {
		log.Fatalf("Invalid integer for %s: %v", key, err)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
