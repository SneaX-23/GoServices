package service

import (
	"context"
	"errors"
	"testing"

	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCacheAndEmit_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		otp           string
		mockStoreErr  error
		mockPubErr    error
		expectedError bool
	}{
		{
			name:          "Success Case",
			email:         "user@test.com",
			otp:           "123456",
			mockStoreErr:  nil,
			mockPubErr:    nil,
			expectedError: false,
		},
		{
			name:          "Cache Failure",
			email:         "user@test.com",
			otp:           "123456",
			mockStoreErr:  errors.New("redis connection lost"),
			mockPubErr:    nil,
			expectedError: true,
		},
		{
			name:          "Messaging Queue Failure",
			email:         "user@test.com",
			otp:           "123456",
			mockStoreErr:  nil,
			mockPubErr:    errors.New("kafka broker unreachable"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOtp := repository.NewMockOTPRepository(t)
			mockQueue := messaging.NewMockProducer(t)
			mockDB := repository.NewMockUserRepository(t)

			scv := NewAuthService(mockDB, mockOtp, mockQueue)

			mockOtp.On("Store", mock.Anything, tt.email, tt.otp).Return(tt.mockStoreErr)

			if tt.mockStoreErr == nil {
				mockQueue.On("PublishUserEvent", mock.Anything, tt.email, tt.otp).Return(tt.mockPubErr)
			}

			err := scv.CacheAndEmit(context.Background(), tt.email, tt.otp)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
