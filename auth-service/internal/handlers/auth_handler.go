package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/SneaX-23/GoServices/auth-service/internal/service"
	"github.com/SneaX-23/GoServices/auth-service/internal/telemetry"
	"github.com/SneaX-23/GoServices/auth-service/internal/utils"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
	tracer   trace.Tracer
}

func NewUserHandler(service *service.AuthService, validate *validator.Validate, tracer trace.Tracer, port int) *http.Server {
	handler := &AuthHandler{
		service:  service,
		validate: validate,
		tracer:   tracer,
	}
	wrappedHandler := telemetry.HTTPMiddleware(handler.RegisteredRoutes())
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      wrappedHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return server
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
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Wrong otp",
		})
		return
	}

	span.SetStatus(codes.Ok, "OTP verified successfully")
	w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Username already exists",
		})
		return
	}

	span.SetStatus(codes.Ok, "username available")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": username.Username + " is available",
	})
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.Sighup")
	defer span.End()

	var user domain.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "body error")
		slog.Error("error decoding json body", "err", err)
		http.Error(w, "Invalid json body", http.StatusBadRequest)
		return
	}
	span.SetAttributes(
		attribute.String("user.Username", user.Username),
		attribute.String("user.Email", user.Email),
	)

	userExits, err := h.service.GetUserByUsername(ctx, user.Username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "server error")
		slog.Error("Server error", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if userExits != nil {
		span.SetAttributes(attribute.Bool("user.exists", true))
		span.SetStatus(codes.Ok, "user already exists")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "User already exists",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "hashing error")
		slog.Error("Error while hashing password", "err", err)
		http.Error(w, "Error while hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hash)
	span.SetAttributes(attribute.Bool("password.hash", true))

	if err := h.service.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error creating user")
		slog.Error("Error while creating user", "err", err)
		http.Error(w, "Internal error while creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Account created successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.Login")
	defer span.End()

	var login domain.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invaid json body")
		slog.Error("Error decoding json body", "err", err)
		http.Error(w, "Invalid json body", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.String("login.email", login.Email))
	span.SetStatus(codes.Ok, "request body decoded")

	user, err := h.service.GetUserByEmail(ctx, login.Email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error finding user")
		slog.Error("Error getting user", "err", err)
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		span.SetStatus(codes.Error, "user Not found")
		slog.Info("User not found")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "User by email: " + login.Email + " does not exists.",
		})
		return
	}
	span.SetStatus(codes.Ok, "user found")
	span.SetAttributes(attribute.String("user.UserID", user.ID))

	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(user.Password))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "wrong password")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Wrong password",
		})
		return
	}

	accessToken, refreshToken, err := h.service.GetAccessAndRefreshT(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error generating tokens")
		slog.Error("Error generating tokens", "err", err)
		http.Error(w, "Error while login", http.StatusInternalServerError)
		return
	}

	isProd := os.Getenv("APP_ENV") == "production"

	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	}

	if !isProd {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)

	response := domain.Response{
		AccessToken: accessToken,
		User: domain.UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.Refresh")
	defer span.End()

	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cookie not found")
		http.Error(w, "Cookie not found", http.StatusBadRequest)
		return
	}

	existingToken, err := h.service.GetExistingRefreshToken(ctx, cookie.Value)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invaid token")
		slog.Error("Invalid refresh token", "err", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if existingToken.ReplacedBy != "" {
		span.SetStatus(codes.Error, "reuse detected")
		slog.Warn("Reuse detected")

		if err := h.service.RevokeAllTokens(ctx, existingToken.UserID); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to revoke")
			http.Error(w, "failed to revoke user tokens", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(
			map[string]any{
				"success": false,
				"message": "Refresh token reuse detected. Please login again",
			},
		)
		return
	}

	if time.Now().After(existingToken.ExpiresAt) {
		if err := h.service.DeleteToken(ctx, existingToken.ID); err != nil {
			span.RecordError(err)
			http.Error(w, "failed to cleanup expired token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Token expired. Please login again.",
		})

		return
	}

	accessToken, refreshToken, err := h.service.RotateToken(ctx, existingToken.ID, existingToken.UserID)
	if err != nil {
		span.RecordError(err)
		slog.Error("Error while rotating token", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	isProd := os.Getenv("APP_ENV") == "production"

	cookie = &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	}

	if !isProd {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accessToken": accessToken,
	})
}

func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "handler.VerifyToken")
	defer span.End()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		span.SetStatus(codes.Error, "no authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		span.SetStatus(codes.Error, "invalid authorization format")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := h.service.VerifyAccessToken(ctx, parts[1])
	if err != nil {
		span.RecordError(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("X-User-ID", userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"verified": true,
	})
}
