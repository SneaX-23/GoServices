package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type NewUserData struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Email struct {
	Email string `json:"email"`
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var rBody Email

	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		slog.Error("Error in json body", "err", err)
		return
	}

	w.Header().Set("Content/type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"success": "true",
		"message": "Otp has been sent to your email",
	})
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
}
