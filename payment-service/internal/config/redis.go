package config

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func GetRedishost() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis:6379"
	}
	return host
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     GetRedishost(),
		Password: "",
		DB:       0,
	})
}
