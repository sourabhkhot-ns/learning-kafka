package kafka

import (
	"github.com/Shopify/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
}

func NewConsumer(brokers string) *Consumer {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer([]string{brokers}, config)
	if err != nil {
		panic(err)
	}

	return &Consumer{
		consumer: consumer,
	}
}

func (c *Consumer) ConsumeMessages(topic string, handler func([]byte) error) error {
	partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		if err := handler(msg.Value); err != nil {
			return err
		}
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
