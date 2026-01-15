package config

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func GetRedisHost() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return "redis:6379"
	}
	return host
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     GetRedisHost(),
		Password: "",
		DB:       0,
	})
}
