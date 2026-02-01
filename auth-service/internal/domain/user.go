package domain

import "time"

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type EmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type Payload struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type VerifEmailPayload struct {
	Type string `json:"type"`
	Data Payload
}

type VerifyOTP struct {
	Email string `json:"email"`
	OTP   int    `json:"otp"`
}

type UsernameRequest struct {
	Username string `json:"username"`
}

type ExistingRefreshToken struct {
	ID          string `json:"id"`
	HashedToken string `json:"hashedtoken"`
	UserID      string `json:"userID"`
	ReplacedBy  string `json:"replacedBy"`
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Response struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}
