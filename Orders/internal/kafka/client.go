package kafka

import (
	"github.com/Shopify/sarama"
)

type Client struct {
	producer sarama.SyncProducer
}

func NewClient(brokers string) *Client {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		panic(err)
	}

	return &Client{
		producer: producer,
	}
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
	return c.producer.Close()
}
