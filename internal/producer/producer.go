package producer

import (
	"context"
	"fmt"

	"github.com/cathudson/order-service/internal/protoxjson"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type KafkaProducer[M proto.Message] struct {
	writer *kafka.Writer
	key    func(M) []byte
}

func NewKafkaProducer[M proto.Message](addr, topic string, key func(M) []byte) *KafkaProducer[M] {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}
	return &KafkaProducer[M]{writer: writer, key: key}
}

func (p *KafkaProducer[M]) Produce(ctx context.Context, event M) error {
	value, err := protoxjson.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   p.key(event),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func (p *KafkaProducer[M]) Close() error {
	return p.writer.Close()
}
