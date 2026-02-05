package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	jwtutil "github.com/SneaX-23/GoServices/auth-service/internal/utils/jwt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthService struct {
	db     repository.UserRepository
	cache  repository.OTPRepository
	queue  messaging.Producer
	tracer trace.Tracer
	secret string
}

func NewAuthService(
	db repository.UserRepository,
	cache repository.OTPRepository,
	queue messaging.Producer,
	tracer trace.Tracer,
	secret string,
) *AuthService {
	return &AuthService{
		db:     db,
		cache:  cache,
		queue:  queue,
		tracer: tracer,
		secret: secret,
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

func (s *AuthService) GetEmailByID(ctx context.Context, userID string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetEmailById")
	defer span.End()

	email, err := s.db.GetEmailByID(ctx, userID)
	if err != nil {
		return "", err
	}

	return email, nil
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

func (s *AuthService) CreateUser(ctx context.Context, user domain.User) error {
	ctx, span := s.tracer.Start(ctx, "service.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.username", user.Username),
		attribute.String("user.email", user.Email),
	)

	if err := s.db.Create(ctx, &user); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GetAccessAndRefreshT(ctx context.Context, userID string) (string, string, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetAccessAndRefreshT")
	defer span.End()

	span.SetAttributes(attribute.String("user.userID", userID))

	accessToken, err := jwtutil.GenerateAccessToken(userID, s.secret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "accessToken generation failed")
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}
	span.SetStatus(codes.Ok, "accessToken generated")

	refreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "refreshToken generation failed")
		return "", "", fmt.Errorf("failed to generate refreshToken: %w")
	}
	span.SetStatus(codes.Ok, "refreshToken generated")

	hashedToken := jwtutil.HashToken(refreshToken)
	if err := s.db.StoreRefreshToken(ctx, userID, hashedToken); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store hashedToken")
		return "", "", fmt.Errorf("failed to store hashedToken: %w", err)
	}
	span.SetStatus(codes.Ok, "hashedToken stored")

	return accessToken, refreshToken, nil
}

func (s *AuthService) GetExistingRefreshToken(ctx context.Context, refreshToken string) (*domain.ExistingRefreshToken, error) {
	ctx, span := s.tracer.Start(ctx, "service.RefreshToken")
	defer span.End()

	hashedToken := jwtutil.HashToken(refreshToken)

	existingToken, err := s.db.FindTokenByHash(ctx, hashedToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error finding existingToken")
		return nil, fmt.Errorf("failed to find existingToken: %w", err)
	}
	return existingToken, nil
}

func (s *AuthService) DeleteToken(ctx context.Context, tokenID string) error {
	ctx, span := s.tracer.Start(ctx, "service.DeleteToken")
	defer span.End()

	err := s.db.DeleteToken(ctx, tokenID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error deleting token")
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

func (s *AuthService) RotateToken(ctx context.Context, tokenID, userID string) (string, string, error) {
	ctx, span := s.tracer.Start(ctx, "service.RotateToken")
	defer span.End()

	span.SetAttributes(attribute.String("user.userID", userID))

	accessToken, err := jwtutil.GenerateAccessToken(userID, s.secret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "accessToken generation failed")
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}
	span.SetStatus(codes.Ok, "accessToken generated")

	refreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "refreshToken generation failed")
		return "", "", fmt.Errorf("failed to generate refreshToken: %w")
	}
	span.SetStatus(codes.Ok, "refreshToken generated")

	hashedToken := jwtutil.HashToken(refreshToken)

	if err := s.db.RotateRefreshToken(ctx, userID, tokenID, hashedToken); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to rotate refreshToken")
		return "", "", fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RevokeAllTokens(ctx context.Context, userID string) error {
	ctx, span := s.tracer.Start(ctx, "service.RevokeAllTokens")
	defer span.End()
	span.SetAttributes(attribute.String("user.ID", userID))

	if err := s.db.RevokeAllTokens(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to revoke")
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

func (s *AuthService) VerifyAccessToken(ctx context.Context, token string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "service.VerifyTokenEndpoint")
	defer span.End()

	claims, err := jwtutil.VerifyAccessToken(token, s.secret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid token")
		return "", err
	}

	return claims.UserID, nil
}
