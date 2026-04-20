package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
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
	if err != nil && !errors.Is(err, domain.ErrOrderAlreadyExists) {
		return fmt.Errorf("create order: %w", err)
	}

	status, delay := simulateProcessing()
	time.Sleep(delay)
	err = p.orderService.UpdateStatus(ctx, order.ID, status)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

//nolint:gosec,mnd // enough here
func simulateProcessing() (domain.OrderStatus, time.Duration) {
	const minDelay = 100 * time.Millisecond
	const maxDelay = 3 * time.Second
	delay := minDelay + time.Duration(rand.Int64N(int64(maxDelay-minDelay)))

	r := rand.IntN(10)
	status := domain.OrderStatusSuccess
	switch {
	case r < 2:
		status = domain.OrderStatusFailed
	case r == 3:
		status = domain.OrderStatusCanceled
	}

	return status, delay
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
