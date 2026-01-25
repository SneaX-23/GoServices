package messaging

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

type Producer interface {
	PublishUserEvent(ctx context.Context, email, otp string) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
	tracer trace.Tracer
}

func NewKafkaProducer(broker, topic string, tracer trace.Tracer) Producer {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		tracer: tracer,
	}
}

