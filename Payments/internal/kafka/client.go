package kafka

import (
	"github.com/Shopify/sarama"
)

type Client struct {
	consumer sarama.Consumer
	producer sarama.SyncProducer
}

func NewClient(brokers string) *Client {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		panic(err)
	}

	consumer, err := sarama.NewConsumer([]string{brokers}, config)
	if err != nil {
		panic(err)
	}

	return &Client{
		consumer: consumer,
		producer: producer,
	}
}

func (c *Client) ConsumeMessages(topic string, handler func([]byte) error) error {
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

func (c *Client) PublishMessage(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := c.producer.SendMessage(msg)
	return err
}

func (c *Client) Close() error {
	if err := c.producer.Close(); err != nil {
		return err
	}
	return c.consumer.Close()
}
