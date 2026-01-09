package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/SneaX-23/GoServices/payment-service/internal/config"
	"github.com/segmentio/kafka-go"
)

// struct to decode respone nody
type ResponseBody struct {
	ReservationID string `json:"reservationId"`
	Amount        int    `json:"amount"`
}

type ProducerData struct {
	ReservationID string    `json:"reservationId"`
	Amount        int       `json:"amount"`
	UserID        string    `json:"userId"`
	TimeStamp     time.Time `json:"timeStamp"`
	Reason        string    `json:"reason"`
}

type Payload struct {
	EventType string       `json:"eventType"`
	Data      ProducerData `json:"data"`
}

// handles payments to /pay route
func HandlePayment(w http.ResponseWriter, r *http.Request) {
	var rBody ResponseBody

	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"success": "processing",
		"message": "Payment is being processed. You will receive an email shortly.",
	})

	userID := r.Header.Get("x-user-id")

	slog.Info("Processing payment", "reservationId", rBody.ReservationID)

	success := rand.Intn(100)
	time.Sleep(2 * time.Second)

	eventType := "PAYMENT_FAILED"
	reason := "abc"

	if success > 20 {
		eventType = "PAYMENT_SUCCESS"
		reason = "xyz"
	}

	eventPayload := Payload{
		EventType: eventType,
		Data: ProducerData{
			ReservationID: rBody.ReservationID,
			UserID:        userID,
			Amount:        rBody.Amount,
			TimeStamp:     time.Now(),
			Reason:        reason,
		},
	}

	serializedData, err := json.Marshal(eventPayload)
	if err != nil {
		slog.Error("Could not marshal event payload", "err", err)
		return
	}

	writer := config.NewProducer("payment-events")
	defer writer.Close()

	if err := writer.WriteMessages(context.Background(), kafka.Message{
		Value: serializedData,
	}); err != nil {
		slog.Error("Failed to publish Kafka message", "err", err)
	}
}
