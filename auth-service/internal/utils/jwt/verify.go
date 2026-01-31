package jwtutil

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

func VerifyAccessToken(tokenString, secret string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	if secret == "" {
		return nil, errors.New("JWT_ACCESS_SECRET is not set")
	}

	// Entry point for jwt verification
	// parses and validates jwt
	// returns *jwt.Token -> parsed token and claims
	//			error -> any failure
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
