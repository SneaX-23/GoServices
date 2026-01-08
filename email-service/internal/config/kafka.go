package config

import (
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// Topics for consumer
const (
	TopicPaymentEvents = "payment-events"
	TopicUserCreated   = "user-created"
)

const ConsumerGroupID = "email-service-consumer-group"

// GetBrokers handles the environment variable logic
func GetBrokers() []string {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "kafka:9092"
	}
	return []string{broker}
}

// NewKafkaReader returns an initialized kafka.Reader
func NewKafkaReader(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  GetBrokers(),
		GroupID:  ConsumerGroupID,
		Topic:    topic,
		MaxWait:  5 * time.Second,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
