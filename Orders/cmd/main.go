package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
)

type MenuItem struct {
	ItemID       int     `json:"item_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	RestaurantID int     `json:"restaurant_id"`
}

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

type CreateOrderRequest struct {
	CustomerID   int         `json:"customer_id" binding:"required"`
	RestaurantID int         `json:"restaurant_id" binding:"required"`
	Items        []OrderItem `json:"items" binding:"required"`
}

type OrderService struct {
	producer sarama.SyncProducer
	orderID  int // Simple counter for demo purposes
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

func (s *OrderService) createOrder(c *gin.Context) {
	var createRequest CreateOrderRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Increment order ID (in a real system, this would come from the database)
	s.orderID++

	// Calculate total amount
	var totalAmount float64
	for _, item := range createRequest.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	order := Order{
		OrderID:      s.orderID,
		CustomerID:   createRequest.CustomerID,
		RestaurantID: createRequest.RestaurantID,
		OrderDate:    time.Now(),
		TotalAmount:  totalAmount,
		Status:       "PENDING",
		Items:        createRequest.Items,
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal order"})
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: "order-events",
		Value: sarama.StringEncoder(orderJSON),
		Key:   sarama.StringEncoder(string(order.OrderID)),
	}

	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish order event"})
		return
	}

	log.Printf("Order event published to partition %d at offset %d\n", partition, offset)
	c.JSON(http.StatusCreated, order)
}

func main() {
	producer, err := newKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer producer.Close()

	orderService := &OrderService{
		producer: producer,
		orderID:  0,
	}

	r := gin.Default()
	r.POST("/orders", orderService.createOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order service starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
