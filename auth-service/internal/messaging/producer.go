package messaging

import (
	"context"
	"encoding/json"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

func (p *kafkaProducer) PublishUserEvent(ctx context.Context, email, otp string) error {
	var payload domain.VerifEmailPayload
	payload.Type = "verify-email"
	payload.Data.Email = email
	payload.Data.Otp = otp

	authPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: authPayload,
	})
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
