package consumer

import (
	"context"

	pe "github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/task"
	"go.uber.org/zap"
)

type OrderResultConsumer struct {
	asynq  task.AsynqClient
	logger *zap.SugaredLogger
}

func NewOrderResultConsumer(asynq task.AsynqClient, logger *zap.SugaredLogger) *CreateOrderConsumer {
	return &CreateOrderConsumer{asynq: asynq, logger: logger}
}

func (c *OrderResultConsumer) Handle(ctx context.Context, event *pe.CreateOrderEvent) error {
	// poll redis here

	return nil
}
