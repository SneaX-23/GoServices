package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPRepository interface {
	Store(ctx context.Context, email, otp string) error
}

type redisOTPRepository struct {
	rdb *redis.Client
}

func NewRedisOtpRepo(rdb *redis.Client) OTPRepository {
	return &redisOTPRepository{rdb: rdb}
}

func (r *redisOTPRepository) Store(ctx context.Context, email, otp string) error {
	return r.rdb.Set(ctx, "otp:"+email, otp, 10*time.Minute).Err()
}
