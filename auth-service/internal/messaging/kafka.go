package messaging

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	PublishUserEvent(ctx context.Context, email, otp string) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(broker, topic string) Producer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}
