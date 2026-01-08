package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/config"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/consumers"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/services"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/templates"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/utils"
)

func main() {
	log := utils.New(os.Getenv("APP_ENV"))
	slog.SetDefault(log)

	emailCfg := config.LoadEmailConfig()

	svcConfig := services.ServiceConfig{LoginURL: "http://localhost:3000/login"}

	renderer, err := templates.NewRenderer("./internal/templates/emails")
	if err != nil {
		slog.Error("Error in rendering templates", "err", err)
	}

	mailer := services.NewResendMailer(emailCfg.ApiKey, emailCfg.From)

	emailService := services.NewEmailservice(mailer, renderer, svcConfig)

	// Start the Consumers
	// run it in goroutine so it doesnt block other consumers
	go consumers.PaymentConsumer(emailService)
	go consumers.NewUserConsumer(emailService)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down...")

}
