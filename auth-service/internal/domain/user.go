package domain

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
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
