package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/service"
	"github.com/cathudson/order-service/internal/task"
	"github.com/hibiken/asynq"
)

type CreateOrderProcessor struct {
	orderService *service.OrderService
}

func NewCreateOrderProcessor(orderService *service.OrderService) *CreateOrderProcessor {
	return &CreateOrderProcessor{orderService: orderService}
}

func (p *CreateOrderProcessor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(task.CreateOrderTaskType, p.ProcessTask)
}

func (p *CreateOrderProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	entity := task.CreateOrderTask{}
	if err := json.Unmarshal(t.Payload(), &entity); err != nil {
		return fmt.Errorf("json unmarshal failed: %w: %w", err, asynq.SkipRetry)
	}

	order := orderFromTask(&entity)
	err := p.orderService.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, domain.ErrOrderAlreadyExists) {
			return nil
		}
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func orderFromTask(t *task.CreateOrderTask) *domain.Order {
	now := time.Now()
	return &domain.Order{
		ID:             t.ID,
		AccountID:      t.AccountID,
		IdempotencyKey: t.IdempotencyKey,
		InstrumentID:   t.InstrumentID,
		Side:           t.OrderSide,
		OrderBy:        t.OrderBy,
		Amount:         t.Amount,
		Quantity:       t.Quantity,
		Price:          nil,
		Status:         domain.OrderStatusNew,
		ErrorMessage:   nil,
		UpdatedAt:      now,
		CreatedAt:      now,
	}
}
