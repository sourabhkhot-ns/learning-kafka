package service

import (
	"encoding/json"

	"github.com/learning-kafka/Orders/internal/kafka"
	"github.com/learning-kafka/Orders/internal/model"
)

type OrderService struct {
	kafkaClient *kafka.Client
	orders      []model.Order
}

func NewOrderService(kafkaClient *kafka.Client) *OrderService {
	return &OrderService{
		kafkaClient: kafkaClient,
		orders:      make([]model.Order, 0),
	}
}

func (s *OrderService) CreateOrder(order *model.Order) error {
	s.orders = append(s.orders, *order)

	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return s.kafkaClient.PublishMessage("order-events", orderJSON)
}

func (s *OrderService) GetOrders() ([]model.Order, error) {
	return s.orders, nil
}
