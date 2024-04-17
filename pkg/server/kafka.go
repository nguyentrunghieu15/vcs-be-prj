package server

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type ConsumerClientInterface interface {
	ReadMessage(context.Context) (kafka.Message, error)
	CloseReader() error
}

type ConsumerKafka struct {
	k *kafka.Reader
}

func (c *ConsumerKafka) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.k.ReadMessage(ctx)
}

func (c *ConsumerKafka) CloseReader() error {
	return c.k.Close()
}

type ConfigConsumer struct {
	Broker     []string
	Topic      string
	GroupID    string
	Logger     kafka.Logger
	ErrorLoger kafka.Logger
}

func NewComsumer(c ConfigConsumer) *ConsumerKafka {
	return &ConsumerKafka{
		k: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers: c.Broker,
				GroupID: c.GroupID,
				Topic:   c.Topic,
			},
		),
	}
}

type ProducerClientInterface interface {
	WriteMessage(context.Context, ...kafka.Message) error
}

type ProducerKafka struct {
	k *kafka.Writer
}

func (c *ProducerKafka) WriteMessage(ctx context.Context, msgs ...kafka.Message) error {
	return c.k.WriteMessages(ctx, msgs...)
}

type ConfigProducer struct {
	Broker     []string
	Topic      string
	Balancer   kafka.Balancer
	Logger     kafka.Logger
	ErrorLoger kafka.Logger
}

func NewProducer(c ConfigProducer) *ProducerKafka {
	return &ProducerKafka{
		k: kafka.NewWriter(
			kafka.WriterConfig{
				Brokers:  c.Broker,
				Topic:    c.Topic,
				Balancer: c.Balancer,
			},
		),
	}
}
