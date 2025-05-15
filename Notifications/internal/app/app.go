package app

import (
	"github.com/learning-kafka/Notifications/internal/kafka"
	"github.com/learning-kafka/Notifications/internal/service"
)

type App struct {
	kafkaConsumer *kafka.Consumer
	service       *service.NotificationService
}

func NewApp(kafkaBrokers string) *App {
	kafkaConsumer := kafka.NewConsumer(kafkaBrokers)
	notificationService := service.NewNotificationService()

	return &App{
		kafkaConsumer: kafkaConsumer,
		service:       notificationService,
	}
}

func (a *App) Run() error {
	return a.kafkaConsumer.ConsumeMessages("payment-events", a.service.SendNotification)
}
