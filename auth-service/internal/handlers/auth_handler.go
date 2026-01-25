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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
	tracer   trace.Tracer
}

func NewUserHandler(service *service.AuthService, validate *validator.Validate, tracer trace.Tracer) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validate,
		tracer:   tracer,
	}
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.VerifyEmail")
	defer span.End()

	var rBody domain.EmailRequest

	// Decode json
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("user.email", rBody.Email))

	// validate email address
	if err := h.validate.Struct(rBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		slog.Warn("Validation failed")
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input:please provide a valid email",
		})
		return
	}

	// Check if email already registered
	user, err := h.service.GetUserByEmail(ctx, rBody.Email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "database error")
		slog.Error("Database error", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Email already registered
	if user != nil {
		span.SetAttributes(attribute.Bool("user.exists", true))
		span.SetStatus(codes.Ok, "email already registered")
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  false,
			"messsage": "Email already registered",
		})
		return
	}

	otp, err := utils.SecureRandom6Digit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate OTP")
		slog.Error("Error generating otp", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// convert otp to string to store in redis cache
	strOpt := strconv.Itoa(int(otp))

	// context for redis
	cacheCtx := context.Background()

	err = h.service.CacheAndEmit(cacheCtx, rBody.Email, strOpt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to cache and emit")
		slog.Error("Error during caching or emiting", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Success response, new email registration
	span.SetStatus(codes.Ok, "OTP sent successfully")
	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Otp has been sent to your email",
	})
}

// VerifyOTP: sends otp if email not registered
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.VerifyOTP")
	defer span.End()

	var req domain.VerifyOTP

	// Decode user entered otp
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("user.email", req.Email),
		attribute.Int("otp.value", req.OTP),
	)

	// GetAndVerifyOTP takes otp from redis and compares with userentered opt
	verifyOTP, err := h.service.GetAndVerifyOTP(ctx, req.Email, req.OTP)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache error")
		slog.Error("Cache error", "err", err)
		http.Error(w, "Internal cache error", http.StatusInternalServerError)
		return
	}

	if !verifyOTP {
		span.SetStatus(codes.Error, "invalid OTP")
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Wrong otp",
		})
		return
	}

	span.SetStatus(codes.Ok, "OTP verified successfully")
	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Otp verified successfully",
	})
}

// Check if a username exists
func (h *AuthHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.CheckUsername")
	defer span.End()

	var username domain.UsernameRequest

	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.Error("JSON decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("user.username", username.Username))

	// Tries to get user by username
	user, err := h.service.GetUserByUsername(ctx, username.Username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "server error")
		slog.Error("Server error", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If username exists user will not be empty
	if user != nil {
		span.SetAttributes(attribute.Bool("username.exists", true))
		span.SetStatus(codes.Ok, "username already exists")
		w.Header().Set("Content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Username already exists",
		})
		return
	}

	span.SetStatus(codes.Ok, "username available")
	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": username.Username + " is available",
	})
}
