package config

import (
	"os"

	"github.com/segmentio/kafka-go"
)

func GetBrokers() []string {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return []string{broker}
}

func NewProducer(topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  GetBrokers(),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}
