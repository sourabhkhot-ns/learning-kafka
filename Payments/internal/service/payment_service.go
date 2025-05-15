package service

import (
	"encoding/json"
	"log"

	"github.com/learning-kafka/Payments/internal/kafka"
)

type PaymentService struct {
	kafkaClient *kafka.Client
}

type Order struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type PaymentEvent struct {
	OrderID string  `json:"order_id"`
	Status  string  `json:"status"`
	Amount  float64 `json:"amount"`
}

func NewPaymentService(kafkaClient *kafka.Client) *PaymentService {
	return &PaymentService{
		kafkaClient: kafkaClient,
	}
}

func (s *PaymentService) ProcessPayment(message []byte) error {
	var order Order
	if err := json.Unmarshal(message, &order); err != nil {
		return err
	}

	// Simulate payment processing
	log.Printf("Processing payment for order %s with amount %.2f", order.ID, order.Amount)

	// Create payment event
	paymentEvent := PaymentEvent{
		OrderID: order.ID,
		Status:  "completed",
		Amount:  order.Amount,
	}

	// Publish payment event
	eventJSON, err := json.Marshal(paymentEvent)
	if err != nil {
		return err
	}

	return s.kafkaClient.PublishMessage("payment-events", eventJSON)
}
