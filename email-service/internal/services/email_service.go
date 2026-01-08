package services

import (
	"fmt"
	"log/slog"

	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/templates"
)

type ServiceConfig struct {
	LoginURL string
}

type EmailService struct {
	mailer   Mailer
	renderer *templates.Renderer
	config   ServiceConfig
}

func NewEmailservice(mailer Mailer, renderer *templates.Renderer, config ServiceConfig) *EmailService {
	return &EmailService{
		mailer:   mailer,
		renderer: renderer,
		config:   config,
	}
}

func (s *EmailService) SendEmail(to, emailType string, data any) error {
	var templateName string
	var subject string

	// Map the messageType(emailType) to a template file
	switch emailType {
	case "PAYMENT_CONFIRMED":
		templateName = "payment_success.html"
		subject = "Payment Successful"
	case "PAYMENT_FAILED":
		templateName = "payment_failed.html"
		subject = "Payment Failed"
	case "new_user":
		templateName = "new_user.html"
		subject = "Welcome!"
	case "REFUND_INITIATED":
		templateName = "payment_refund.html"
		subject = "Refund Initiated"
	default:
		slog.Error("Unknown email type", "type", emailType)
		return fmt.Errorf("unknown email type: %s", emailType)
	}

	// Render the body
	body, err := s.renderer.Render(templateName, data)
	if err != nil {
		return err
	}

	// Send via Mailer
	return s.mailer.Send(to, subject, body)
}
