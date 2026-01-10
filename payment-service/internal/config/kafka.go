package config

import (
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

const TopicPaymentEvents = "payment-events"

const ConsumerGroupID = "payment-service-group"

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

func NewKafkaReader(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  GetBrokers(),
		GroupID:  ConsumerGroupID,
		Topic:    topic,
		MaxWait:  5 * time.Second,
		MinBytes: 10e3, //10KB
		MaxBytes: 10e6, //10MB
	})
}
