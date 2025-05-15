package service

import (
	"encoding/json"
	"log"
)

type NotificationService struct{}

type PaymentEvent struct {
	OrderID string  `json:"order_id"`
	Status  string  `json:"status"`
	Amount  float64 `json:"amount"`
}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) SendNotification(message []byte) error {
	var event PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return err
	}

	// Simulate sending notification
	log.Printf("Sending notification for order %s: Payment %s for amount %.2f",
		event.OrderID, event.Status, event.Amount)
	return nil
}
