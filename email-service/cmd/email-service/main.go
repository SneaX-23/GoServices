package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/SneaX-23/GoServices/auth-service/pkg/genproto"
	"github.com/SneaX-23/GoServices/email-service/internal/config"
	"github.com/SneaX-23/GoServices/email-service/internal/consumers"
	"github.com/SneaX-23/GoServices/email-service/internal/services"
	"github.com/SneaX-23/GoServices/email-service/internal/templates"
	"github.com/SneaX-23/GoServices/email-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := utils.New(os.Getenv("APP_ENV"))
	slog.SetDefault(log)
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to gRPC server", "err", err)
	}

	defer conn.Close()

	authClient := pb.NewAuthServiceClient(conn)

	emailCfg := config.LoadEmailConfig()

	svcConfig := services.ServiceConfig{LoginURL: "http://localhost:3000/login"}

	renderer, err := templates.NewRenderer("./internal/templates/emails")
	if err != nil {
		slog.Error("Error in rendering templates", "err", err)
	}

	mailer := services.NewResendMailer(emailCfg.ApiKey, emailCfg.From)

	emailService := services.NewEmailservice(mailer, renderer, svcConfig, authClient)

	// Start the Consumers
	// run it in goroutine so it doesnt block other consumers
	go consumers.PaymentConsumer(emailService)
	go consumers.NewUserConsumer(emailService)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down...")
}
