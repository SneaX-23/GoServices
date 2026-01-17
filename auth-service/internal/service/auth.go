package service

import (
	"context"

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

func (s *AuthService) CacheOtp(ctx context.Context, email, otp string) error {
	return s.cache.Store(ctx, email, otp)
}

func (s *AuthService) QueueEvent(ctx context.Context, email, otp string) error {
	return s.queue.PublishUserEvent(ctx, email, otp)
}
