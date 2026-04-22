package producer

import (
	"context"

	"github.com/cathudson/order-service/internal/config"
	"github.com/cathudson/order-service/internal/proto"
)

//go:generate moq -rm -out gen/order_result_producer_mock.go -pkg producergen . OrderResultProducer

type OrderResultProducer interface {
	Produce(ctx context.Context, event *proto.OrderResultEvent) error
	Close() error
}

func NewOrderResultProducer(cfg config.KafkaConfig) OrderResultProducer {
	return NewKafkaProducer[*proto.OrderResultEvent](
		cfg.Address,
		cfg.Producers.OrderResultTopic,
		func(e *proto.OrderResultEvent) []byte {
			return []byte(e.GetId().GetValue())
		},
	)
}
