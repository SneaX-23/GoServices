package handlers

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (h *AuthHandler) RegisteredRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("api/v1/auth", func(r chi.Router) {
		r.Post("/verify-email", h.VerifyEmail)
		r.Post("verify-otp", h.VerifyOTP)
		r.Post("check-username", h.CheckUsername)
		r.Post("signup", h.Signup)
		r.Post("login", h.Login)
		r.Get("verify", h.VerifyToken)
		r.Get("refresh", h.Refresh)
	})

	return r
}
