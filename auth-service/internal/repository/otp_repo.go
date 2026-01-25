package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OTPRepository interface {
	Store(ctx context.Context, email, otp string) error
	GetOtp(ctx context.Context, email string) (string, error)
}

type redisOTPRepository struct {
	rdb    *redis.Client
	tracer trace.Tracer
}

func NewRedisOtpRepo(rdb *redis.Client, tracer trace.Tracer) OTPRepository {
	return &redisOTPRepository{
		rdb:    rdb,
		tracer: tracer,
	}
}

func (r *redisOTPRepository) Store(ctx context.Context, email, otp string) error {
	ctx, span := r.tracer.Start(ctx, "repository.StoreOTP")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.operation", "set"),
		attribute.String("cache.key", "otp:"+email),
	)

	err := r.rdb.Set(ctx, "otp:"+email, otp, 10*time.Minute).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store OTP")
		return err
	}

	span.SetStatus(codes.Ok, "OTP stored successfully")
	return nil
}

func (r *redisOTPRepository) GetOtp(ctx context.Context, email string) (string, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetOTP")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.operation", "get"),
		attribute.String("cache.key", "otp:"+email),
	)

	val, err := r.rdb.Get(ctx, "otp:"+email).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get OTP")
		return "", err
	}

	span.SetStatus(codes.Ok, "OTP retrieved successfully")
	return val, nil
}

