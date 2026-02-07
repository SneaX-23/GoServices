package messaging

import (
	"context"
	"encoding/json"

	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type KafkaHeaderCarrier []kafka.Header

func (c *KafkaHeaderCarrier) Set(key string, value string) {
	*c = append(*c, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c *KafkaHeaderCarrier) Get(key string) string {
	for _, h := range *c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = h.Key
	}
	return keys
}

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

	headers := KafkaHeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &headers)

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value:   authPayload,
		Headers: []kafka.Header(headers),
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
