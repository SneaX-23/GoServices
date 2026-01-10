package consumers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/SneaX-23/GoServices/payment-service/internal/config"
	"github.com/segmentio/kafka-go"
)

type PaymentPayload struct {
	ReservationId string `json:"reservationId"`
	UserId        string `json:"userId"`
	Amount        int    `json:"amount"`
	TimeStamp     string `json:"timestamp"`
	Reason        string `json:"reason"`
}

// The top-level message envelope
type KafkaEnvelope struct {
	Type string         `json:"type"`
	Data PaymentPayload `json:"data"`
}

func RefundConsumer() {
	reader := config.NewKafkaReader(config.TopicPaymentEvents)
	defer reader.Close()

	workLimit := make(chan struct{}, 5)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Error reading kafka message", "err", err)
			break
		}

		workLimit <- struct{}{}

		go func(msg kafka.Message) {
			defer func() { <-workLimit }()

			var envelope KafkaEnvelope
			if err := json.Unmarshal(msg.Value, &envelope); err != nil {
				slog.Error("Error parsing message", "err", err)
				return
			}

			messagetype := envelope.Type
			slog.Info("Received message", "type", messagetype)

			// missing persistant database storage for single source of truth
			// may need to update payment-events kafka topic to only
			// send email after successful refund

			if messagetype == "REFUND_INITIATED" {
				// redis client
				rdb := config.NewRedisClient()
				defer rdb.Close()

				slog.Info("Refund succesfull for reservation id", "reservationId", envelope.Data.ReservationId)
				slog.Info("For the amoutn", "amount", envelope.Data.Amount)
			}

		}(m)
	}
}
