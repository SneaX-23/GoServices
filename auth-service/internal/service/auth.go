package service

import (
	"context"
	"strconv"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthService struct {
	db     repository.UserRepository
	cache  repository.OTPRepository
	queue  messaging.Producer
	tracer trace.Tracer
}

func NewAuthService(
	db repository.UserRepository,
	cache repository.OTPRepository,
	queue messaging.Producer,
	tracer trace.Tracer,
) *AuthService {
	return &AuthService{
		db:     db,
		cache:  cache,
		queue:  queue,
		tracer: tracer,
	}
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := s.db.GetByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return nil, err
	}

	span.SetStatus(codes.Ok, "user retrieved")
	return user, nil
}

func (s *AuthService) CacheAndEmit(ctx context.Context, email, otp string) error {
	ctx, span := s.tracer.Start(ctx, "service.CacheAndEmit")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	// cache otp
	err := s.cache.Store(ctx, email, otp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cache OTP")
		return err
	}

	// emit kafka event
	err = s.queue.PublishUserEvent(ctx, email, otp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish event")
		return err
	}

	span.SetStatus(codes.Ok, "OTP cached and event published")
	return nil
}

func (s *AuthService) GetAndVerifyOTP(ctx context.Context, email string, otp int) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetAndVerifyOTP")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	// convert string to otp
	strOTP := strconv.Itoa(int(otp))

	// get otp from redis cache
	valOTP, err := s.cache.GetOtp(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get OTP")
		return false, err
	}

	isValid := strOTP == valOTP
	span.SetAttributes(attribute.Bool("otp.valid", isValid))

	if isValid {
		span.SetStatus(codes.Ok, "OTP verified")
	} else {
		span.SetStatus(codes.Error, "OTP invalid")
	}

	return isValid, nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetUserByUsername")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	user, err := s.db.GetByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return nil, err
	}

	span.SetStatus(codes.Ok, "user retrieved")
	return user, nil
}

