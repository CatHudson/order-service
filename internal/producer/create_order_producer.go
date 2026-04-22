package producer

import (
	"context"

	"github.com/cathudson/order-service/internal/config"
	"github.com/cathudson/order-service/internal/proto"
)

//go:generate moq -rm -out gen/create_order_producer_mock.go -pkg producergen . CreateOrderProducer

type CreateOrderProducer interface {
	Produce(ctx context.Context, event *proto.CreateOrderEvent) error
	Close() error
}

func NewCreateOrderProducer(cfg config.KafkaConfig) CreateOrderProducer {
	return NewKafkaProducer[*proto.CreateOrderEvent](
		cfg.Address,
		cfg.Producers.CreateOrderTopic,
		func(e *proto.CreateOrderEvent) []byte {
			return []byte(e.GetAccountId().GetValue())
		},
	)
}
