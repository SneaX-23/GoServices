package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

func NewClaims(userID string) *Claims {
	return &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "auth-service",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
}

func GenerateAccessToken(userID, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("JWT_ACCESS_SECRET is not set")
	}

	claims := NewClaims(userID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
