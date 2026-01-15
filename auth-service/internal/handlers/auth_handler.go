package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"github.com/SneaX-23/GoServices/auth-service/internal/utils"
)

type UserHandler struct {
	repo repository.UserRepository
	rdb  repository.OTPRepository
}

func NewUserHandler(repo repository.UserRepository, rdb repository.OTPRepository) *UserHandler {
	return &UserHandler{repo: repo, rdb: rdb}
}

type NewUserData struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var rBody EmailRequest

	// Decode json
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if email already registered
	user, err := h.repo.GetByEmail(r.Context(), rBody.Email)
	if err != nil {
		slog.Error("Database error", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Email already registered
	if user != nil {
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  false,
			"messsage": "Email already registered",
		})
		return
	}
	// Success response, new email registration
	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Otp has been sent to your email",
	})
	otp, err := utils.SecureRandom6Digit()
	if err != nil {
		slog.Error("Errr generating otp", "err", err)
	}
	strOpt := strconv.Itoa(int(otp))
	// context for redis
	ctx := context.Background()

	err = h.rdb.Store(ctx, rBody.Email, strOpt)
	if err != nil {
		slog.Error("Error storing email otp in redis", "err", err)
		return
	}
}
