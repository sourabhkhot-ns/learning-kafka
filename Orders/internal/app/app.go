package app

import (
	"github.com/gin-gonic/gin"
	"github.com/learning-kafka/Orders/internal/handler"
	"github.com/learning-kafka/Orders/internal/kafka"
	"github.com/learning-kafka/Orders/internal/service"
)

type App struct {
	router      *gin.Engine
	kafkaClient *kafka.Client
	handler     *handler.OrderHandler
	service     *service.OrderService
}

func NewApp(kafkaBrokers string) *App {
	kafkaClient := kafka.NewClient(kafkaBrokers)
	orderService := service.NewOrderService(kafkaClient)
	orderHandler := handler.NewOrderHandler(orderService)

	router := gin.Default()

	return &App{
		router:      router,
		kafkaClient: kafkaClient,
		handler:     orderHandler,
		service:     orderService,
	}
}

func (a *App) Run() error {
	a.setupRoutes()
	return a.router.Run(":8080")
}

func (a *App) setupRoutes() {
	v1 := a.router.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.POST("", a.handler.CreateOrder)
			orders.GET("", a.handler.GetOrders)
		}
	}
}
