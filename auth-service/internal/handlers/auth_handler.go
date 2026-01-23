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
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
}

func NewUserHandler(service *service.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validate,
	}
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var rBody domain.EmailRequest

	// Decode json
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// validate email address
	if err := h.validate.Struct(rBody); err != nil {
		slog.Warn("Validation failed")
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input:please provide a valid email",
		})
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

	err = h.service.CacheAndEmit(ctx, rBody.Email, strOpt)
	if err != nil {
		slog.Error("Error during caching or emiting", "err", err)
		return
	}
}

// VerifyOTP: sends otp if email not registered
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req domain.VerifyOTP

	// Decode user entered otp
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// GetAndVerifyOTP takes otp from redis and compares with userentered opt
	verifyOTP, err := h.service.GetAndVerifyOTP(r.Context(), req.Email, req.OTP)
	if err != nil {
		slog.Error("Cache error", "err", err)
		http.Error(w, "Internal cach error", http.StatusInternalServerError)
		return
	}

	if !verifyOTP {
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Wrong otp",
		})
	}

	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"message": "Otp verified successfuly",
	})
}

// Check if a username exists
func (h *AuthHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	var username domain.UsernameRequest

	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Tries to get user by username
	user, err := h.service.GetUserByUsername(r.Context(), username.Username)
	if err != nil {
		slog.Error("Server error", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If username exists user will not be empty
	if user != nil {
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Username already exists",
		})
		return
	}

	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": username.Username + " is available",
	})
}
