package main

import (
	"log"
	"os"

	"github.com/learning-kafka/Notifications/internal/app"
)

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	application := app.NewApp(kafkaBrokers)
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
