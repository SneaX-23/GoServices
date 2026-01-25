package messaging

import (
	"context"
	"encoding/json"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (p *kafkaProducer) PublishUserEvent(ctx context.Context, email, otp string) error {
	ctx, span := p.tracer.Start(ctx, "messaging.PublishUserEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination", p.writer.Topic),
		attribute.String("messaging.operation", "publish"),
	)

	var payload domain.VerifEmailPayload
	payload.Type = "verify-email"
	payload.Data.Email = email
	payload.Data.Otp = otp

	authPayload, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal payload")
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: authPayload,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish message")
		return err
	}

	span.SetStatus(codes.Ok, "message published successfully")
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
