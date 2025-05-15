package app

import (
	"github.com/learning-kafka/Payments/internal/kafka"
	"github.com/learning-kafka/Payments/internal/service"
)

type App struct {
	kafkaClient *kafka.Client
	service     *service.PaymentService
}

func NewApp(kafkaBrokers string) *App {
	kafkaClient := kafka.NewClient(kafkaBrokers)
	paymentService := service.NewPaymentService(kafkaClient)

	return &App{
		kafkaClient: kafkaClient,
		service:     paymentService,
	}
}

func (a *App) Run() error {
	return a.kafkaClient.ConsumeMessages("order-events", a.service.ProcessPayment)
}
