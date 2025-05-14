package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

type PaymentResult struct {
	OrderID       int       `json:"order_id"`
	CustomerID    int       `json:"customer_id"`
	RestaurantID  int       `json:"restaurant_id"`
	TotalAmount   float64   `json:"total_amount"`
	PaymentStatus string    `json:"payment_status"`
	ProcessedAt   time.Time `json:"processed_at"`
	TransactionID string    `json:"transaction_id"`
}

type NotificationService struct{}

func (s *NotificationService) sendNotification(payment PaymentResult) error {
	// Simulate sending notification (e.g., email, SMS)
	log.Printf("Sending notification for Order #%d:\n"+
		"Customer ID: %d\n"+
		"Restaurant ID: %d\n"+
		"Amount: $%.2f\n"+
		"Status: %s\n"+
		"Transaction ID: %s\n"+
		"Processed At: %v\n",
		payment.OrderID,
		payment.CustomerID,
		payment.RestaurantID,
		payment.TotalAmount,
		payment.PaymentStatus,
		payment.TransactionID,
		payment.ProcessedAt)
	return nil
}

func setupConsumerGroup() (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{"localhost:9092"}
	if brokersEnv := os.Getenv("KAFKA_BROKERS"); brokersEnv != "" {
		brokers = []string{brokersEnv}
	}

	group, err := sarama.NewConsumerGroup(brokers, "notification-service", config)
	if err != nil {
		return nil, err
	}

	return group, nil
}

type ConsumerGroupHandler struct {
	notificationService *NotificationService
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var payment PaymentResult
		if err := json.Unmarshal(message.Value, &payment); err != nil {
			log.Printf("Error unmarshaling payment result: %v", err)
			continue
		}

		log.Printf("Received payment event for order: %d", payment.OrderID)
		if err := h.notificationService.sendNotification(payment); err != nil {
			log.Printf("Error sending notification: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func main() {
	notificationService := &NotificationService{}

	group, err := setupConsumerGroup()
	if err != nil {
		log.Fatalf("Failed to initialize consumer group: %v", err)
	}
	defer group.Close()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			topics := []string{"payment-events"}
			handler := &ConsumerGroupHandler{notificationService: notificationService}

			if err := group.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm
	log.Println("Shutting down notification service...")
	cancel()
	wg.Wait()
}
