package consumers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/config"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/services"
	"github.com/segmentio/kafka-go"
)

// The nested datae
type PaymentPayload struct {
	ReservationId string `json:"reservationId"`
	UserId        string `json:"userId"`
	Amount        int    `json:"amount"`
	TimeStamp     string `json:"timestamp"`
	Reason        string `json:"reason"`
}

// The top-level message envelope
type KafkaEnvelope struct {
	Type string         `json:"type"`
	Data PaymentPayload `json:"data"`
}

func PaymentConsumer(emailservice *services.EmailService) {
	reader := config.NewKafkaReader(config.TopicPaymentEvents)
	defer reader.Close()

	worklimit := make(chan struct{}, 10)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Error reading kafka messages", "err", err)
			break
		}

		worklimit <- struct{}{}

		go func(msg kafka.Message) {
			defer func() { <-worklimit }()

			ctx, canlce := context.WithTimeout(context.Background(), 5*time.Second)
			defer canlce()

			var envelope KafkaEnvelope
			if err := json.Unmarshal(msg.Value, &envelope); err != nil {
				slog.Error("Error parsing message", "err", err)
				return
			}

			messageType := envelope.Type
			slog.Info("Received message", "type", messageType)

			// get user email from user service
			email, err := config.GetUserEmail(ctx, envelope.Data.UserId)
			if err != nil {
				slog.Error("Error getting user email from user service", "err", err)
				return
			}

			// send email with retries
			err = SendEmailWithRetry(emailservice, email, messageType, envelope.Data)
			if err != nil {
				slog.Info("Failed to sent email to %s after retries", "email", email)
				return
			}

			// commit message
			reader.CommitMessages(context.Background(), msg)
			slog.Info("Email sent to %s", "email", email)

		}(m)
	}
}

func SendEmailWithRetry(emailService *services.EmailService, email string, messageType string, data any) error {
	var err error
	for i := range 3 {
		err = emailService.SendEmail(email, messageType, data)
		if err == nil {
			return nil
		}
		slog.Info("Retrying eamil %s .. Attempt: %d", email, i+1)
		time.Sleep(2 * time.Second)
	}
	return err
}
