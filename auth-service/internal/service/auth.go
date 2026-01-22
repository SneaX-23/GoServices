package service

import (
	"context"
	"strconv"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
)

type AuthService struct {
	db    repository.UserRepository
	cache repository.OTPRepository
	queue messaging.Producer
}

func NewAuthService(db repository.UserRepository, cache repository.OTPRepository, queue messaging.Producer) *AuthService {
	return &AuthService{
		db:    db,
		cache: cache,
		queue: queue,
	}
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.db.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) CacheAndEmit(ctx context.Context, email, otp string) error {
	// cache otp
	err := s.cache.Store(ctx, email, otp)
	if err != nil {
		return err
	}
	// emit kafka event
	return s.queue.PublishUserEvent(ctx, email, otp)
}

func (s *AuthService) GetAndVerifyOTP(ctx context.Context, email string, otp int) (bool, error) {
	// convert string to otp
	strOTP := strconv.Itoa(int(otp))

	// get otp from redis cache
	valOTP, err := s.cache.GetOtp(ctx, email)
	if err != nil {
		return false, err
	}

	return strOTP == valOTP, nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.db.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}
