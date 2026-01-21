package service

import (
	"context"
	"errors"
	"testing"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
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

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockReturn    *domain.User
		mockErr       error
		expectedError bool
	}{
		{
			name:  "Success - User Found",
			email: "found@you.com",
			mockReturn: &domain.User{
				Email: "found@you.com",
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name:          "Error - Database Failure",
			email:         "error@exa.com",
			mockReturn:    nil,
			mockErr:       errors.New("db connection lost"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := repository.NewMockUserRepository(t)

			svc := NewAuthService(mockDB, nil, nil)

			mockDB.On("GetByEmail", mock.Anything, tt.email).Return(tt.mockReturn, tt.mockErr)

			user, err := svc.GetUserByEmail(context.Background(), tt.email)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.email, user.Email)
			}
		})
	}
}
