package messaging

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type User struct {
	Email string `json:"email"`
}

func (p *KafkaProducer) PublishUserEvent(ctx context.Context, user User) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
