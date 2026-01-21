package service

import (
	"context"
	"testing"

	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestCacheAndEmit__Success(t *testing.T) {
	mockOTP := repository.NewMockOTPRepository(t)
	mockQueue := messaging.NewMockProducer(t)
	mockDB := repository.NewMockUserRepository(t)

	svc := NewAuthService(mockDB, mockOTP, mockQueue)
	ctx := context.Background()

	email := "dev@something.com"
	otp := "123456"

	mockOTP.On("Store", ctx, email, otp).Return(nil)
	mockQueue.On("PublishUserEvent", ctx, email, otp).Return(nil)

	err := svc.CacheAndEmit(ctx, email, otp)

	assert.NoError(t, err)
}
