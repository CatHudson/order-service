package consumer

import (
	"context"
	"errors"
	"fmt"

	pe "github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/task"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var errSkipMessage = errors.New("skip message")

type CreateOrderConsumer struct {
	asynq  task.AsynqClient
	logger *zap.SugaredLogger
}

func NewCreateOrderConsumer(asynq task.AsynqClient, logger *zap.SugaredLogger) *CreateOrderConsumer {
	return &CreateOrderConsumer{asynq: asynq, logger: logger}
}

func (c *CreateOrderConsumer) Handle(ctx context.Context, event *pe.CreateOrderEvent) error {
	err := c.validate(event)
	if err != nil {
		c.logger.Errorw("validation error", "error", err)
		return errSkipMessage
	}

	t := taskFromEvent(event)
	err = c.asynq.Enqueue(ctx, task.CreateOrderTaskType, t, asynq.TaskID(event.GetId().GetValue()))
	if err != nil {
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			c.logger.Infow("skipped duplicate task", "error", err)
			return nil
		}
		return err
	}

	return nil
}

func (c *CreateOrderConsumer) validate(event *pe.CreateOrderEvent) error {
	if _, err := uuid.Parse(event.GetId().GetValue()); err != nil {
		return fmt.Errorf("create-order consumer: invalid ID: %w", err)
	}
	if _, err := uuid.Parse(event.GetAccountId().GetValue()); err != nil {
		return fmt.Errorf("create-order consumer: invalid accountID: %w", err)
	}
	if _, err := uuid.Parse(event.GetInstrumentId().GetValue()); err != nil {
		return fmt.Errorf("create-order consumer: invalid instrumentID: %w", err)
	}
	if event.GetSide() == pe.OrderSide_ORDER_SIDE_UNSPECIFIED {
		return fmt.Errorf("create-order consumer: UNSPECIFIED side")
	}
	if event.GetOrderBy() == pe.OrderBy_ORDER_BY_UNSPECIFIED {
		return fmt.Errorf("create-order consumer: UNSPECIFIED order_by")
	}
	if event.GetQuantity() == nil && event.GetAmount() == nil {
		return fmt.Errorf("create-order consumer: qty and amount are nil")
	}
	return nil
}
