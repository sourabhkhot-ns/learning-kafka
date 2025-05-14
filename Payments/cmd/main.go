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

type OrderItem struct {
	OrderItemID int     `json:"order_item_id"`
	ItemID      int     `json:"item_id"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	OrderID      int         `json:"order_id"`
	CustomerID   int         `json:"customer_id"`
	RestaurantID int         `json:"restaurant_id"`
	OrderDate    time.Time   `json:"order_date"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status"`
	Items        []OrderItem `json:"items"`
}

type PaymentResult struct {
	OrderID       int       `json:"order_id"`
	CustomerID    int       `json:"customer_id"`
	RestaurantID  int       `json:"restaurant_id"`
	TotalAmount   float64   `json:"total_amount"`
	PaymentStatus string    `json:"payment_status"`
	ProcessedAt   time.Time `json:"processed_at"`
	TransactionID string    `json:"transaction_id"`
}

type PaymentService struct {
	producer sarama.SyncProducer
}

func newKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	brokers := []string{"localhost:9092"}
	if brokersEnv := os.Getenv("KAFKA_BROKERS"); brokersEnv != "" {
		brokers = []string{brokersEnv}
	}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func (s *PaymentService) processPayment(order Order) error {
	// Simulate payment processing
	time.Sleep(2 * time.Second)

	// Generate a simple transaction ID (in a real system, this would come from the payment provider)
	transactionID := "TXN-" + time.Now().Format("20060102150405")

	result := PaymentResult{
		OrderID:       order.OrderID,
		CustomerID:    order.CustomerID,
		RestaurantID:  order.RestaurantID,
		TotalAmount:   order.TotalAmount,
		PaymentStatus: "COMPLETED",
		ProcessedAt:   time.Now(),
		TransactionID: transactionID,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "payment-events",
		Value: sarama.StringEncoder(resultJSON),
		Key:   sarama.StringEncoder(string(order.OrderID)),
	}

	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Payment event published to partition %d at offset %d\n", partition, offset)
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

	group, err := sarama.NewConsumerGroup(brokers, "payment-service", config)
	if err != nil {
		return nil, err
	}

	return group, nil
}

type ConsumerGroupHandler struct {
	paymentService *PaymentService
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var order Order
		if err := json.Unmarshal(message.Value, &order); err != nil {
			log.Printf("Error unmarshaling order: %v", err)
			continue
		}

		log.Printf("Processing payment for order: %d, Amount: %.2f", order.OrderID, order.TotalAmount)
		if err := h.paymentService.processPayment(order); err != nil {
			log.Printf("Error processing payment: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func main() {
	producer, err := newKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer producer.Close()

	paymentService := &PaymentService{
		producer: producer,
	}

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
			topics := []string{"order-events"}
			handler := &ConsumerGroupHandler{paymentService: paymentService}

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
	log.Println("Shutting down payment service...")
	cancel()
	wg.Wait()
}
