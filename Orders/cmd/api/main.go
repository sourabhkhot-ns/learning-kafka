package main

import (
	"log"
	"os"
	"strings"

	"github.com/learning-kafka/Orders/internal/app"
)

func main() {
	kafkaBrokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	application := app.NewApp(kafkaBrokers)
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
