package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/SneaX-23/GoServices/auth-service/internal/service"
	"github.com/SneaX-23/GoServices/auth-service/internal/utils"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewUserHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var rBody domain.EmailRequest

	// Decode json
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if email already registered
	user, err := h.service.GetUserByEmail(r.Context(), rBody.Email)
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

	// convert otp to string to store in redis cache
	strOpt := strconv.Itoa(int(otp))

	// context for redis
	ctx := context.Background()

	err = h.service.CacheOtp(ctx, rBody.Email, strOpt)
	if err != nil {
		slog.Error("Error storing email otp in redis", "err", err)
		return
	}

	err = h.service.QueueEvent(ctx, rBody.Email, strOpt)
	if err != nil {
		slog.Error("Error emiting kafka event", "err", err)
		return
	}
}
