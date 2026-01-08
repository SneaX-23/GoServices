package consumers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/config"
	"github.com/SneaX-23/Microservices-project/microservices-project/email-service/internal/services"
	"github.com/segmentio/kafka-go"
)

type NewUserPayload struct {
	Email     string `json:"email"`
	UserName  string `json:"username"`
	TimeStamp string `json:"timeStamp"`
}

type NewUserEnvelope struct {
	Type string         `json:"type"`
	Data NewUserPayload `json:"data"`
}

func NewUserConsumer(emailservice *services.EmailService) {
	reader := config.NewKafkaReader(config.TopicUserCreated)
	defer reader.Close()

	worklimit := make(chan struct{}, 5)

	for {
		m, err := reader.ReadMessage(context.Background())

		if err != nil {
			slog.Error("Error reading user created message", "err", err)
			break
		}

		worklimit <- struct{}{}

		go func(msg kafka.Message) {
			defer func() { <-worklimit }()

			var envelope NewUserEnvelope
			if err := json.Unmarshal(msg.Value, &envelope); err != nil {
				slog.Error("Error parsing message", "err", err)
				return
			}

			messageType := envelope.Type
			slog.Info("Received message", "type", messageType)
			// send email with retries
			err = SendEmailWithRetry(emailservice, envelope.Data.Email, messageType, envelope.Data)
			if err != nil {
				slog.Info("Failed to sent email to %s after retries", "email", envelope.Data.Email)
				return
			}
			// commit message
			reader.CommitMessages(context.Background(), msg)
			slog.Info("Email sent to %s", "email", envelope.Data.Email)
		}(m)
	}
}
