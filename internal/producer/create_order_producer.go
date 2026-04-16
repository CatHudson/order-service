package producer

import (
	"context"

	"github.com/cathudson/order-service/internal/config"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/protoxjson"
	"github.com/segmentio/kafka-go"
)

//go:generate moq -rm -out gen/create_order_producer_mock.go -pkg producergen . CreateOrderProducer

type CreateOrderProducer interface {
	Produce(ctx context.Context, event *proto.CreateOrderEvent) error
	Close() error
}

type createOrderProducer struct {
	writer *kafka.Writer
}

func NewCreateOrderProducer(cfg config.KafkaConfig) CreateOrderProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Address),
		Topic:    cfg.Producers.CreateOrderTopic,
		Balancer: &kafka.Hash{},
	}
	return &createOrderProducer{writer: writer}
}

func (c *createOrderProducer) Produce(ctx context.Context, event *proto.CreateOrderEvent) error {
	value, err := protoxjson.Marshal(event)
	if err != nil {
		return err
	}
	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.GetAccountId().GetValue()),
		Value: value,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *createOrderProducer) Close() error {
	return c.writer.Close()
}
