package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/SneaX-23/GoServices/payment-service/internal/handlers"
	"github.com/SneaX-23/GoServices/payment-service/internal/utils"
)

func main() {
	logger := utils.New(os.Getenv("APP_ENV"))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /pay", handlers.HandlePayment)

	srv := &http.Server{
		Addr:    ":4200",
		Handler: mux,
	}

	slog.Info("Server listening on port", "port", 4200)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
	}
}
