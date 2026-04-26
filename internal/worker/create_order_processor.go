package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/mapper"
	"github.com/cathudson/order-service/internal/producer"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/service"
	"github.com/cathudson/order-service/internal/task"
	"github.com/cathudson/order-service/internal/util"
	"github.com/hibiken/asynq"
	"github.com/shopspring/decimal"
)

type CreateOrderProcessor struct {
	orderService        *service.OrderService
	orderResultProducer producer.OrderResultProducer
}

func NewCreateOrderProcessor(orderService *service.OrderService, orderResultProducer producer.OrderResultProducer) *CreateOrderProcessor {
	return &CreateOrderProcessor{orderService: orderService, orderResultProducer: orderResultProducer}
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
	switch {
	case err == nil:
		// fresh order - continue processing
	case errors.Is(err, domain.ErrOrderAlreadyExists):
		dbOrder, dbErr := p.orderService.GetByID(ctx, order.ID)
		if dbErr != nil {
			return fmt.Errorf("failed to fetch existing order from Postgres: %w", dbErr)
		}
		order = dbOrder
	default:
		return fmt.Errorf("create order: %w", err)
	}

	if !order.IsTerminal() {
		err = p.process(ctx, order)
		if err != nil {
			return fmt.Errorf("process order: %w", err)
		}
	}

	orderResult := &proto.OrderResultEvent{
		Id:           &proto.UUID{Value: order.ID.String()},
		Side:         mapper.OrderSideToProto(order.Side),
		OrderBy:      mapper.OrderByToProto(order.OrderBy),
		Status:       mapper.OrderStatusToProto(order.Status),
		Price:        util.DecimalToMoney(order.Price),
		Amount:       util.DecimalToMoney(order.Amount),
		Quantity:     util.DecimalToProto(order.Quantity),
		ErrorMessage: order.ErrorMessage,
	}
	if err = p.orderResultProducer.Produce(ctx, orderResult); err != nil {
		return fmt.Errorf("produce order result: %w", err)
	}

	return nil
}

//nolint:gosec,mnd // enough here
func (p *CreateOrderProcessor) process(ctx context.Context, order *domain.Order) error {
	if err := p.orderService.UpdateStatus(ctx, order.ID, domain.OrderStatusPending, nil); err != nil {
		return fmt.Errorf("process order switch to pending: %w", err)
	}

	priceInt := rand.IntN(1000) + 1
	priceF := float64(priceInt) + rand.Float64()
	price := new(decimal.RequireFromString(fmt.Sprintf("%.5f", priceF)))
	order.Price = price

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
	order.Status = status

	time.Sleep(delay)

	if status != domain.OrderStatusSuccess {
		var errorMessage *string
		if status == domain.OrderStatusFailed {
			errorMessage = new("order failed")
		}
		if err := p.orderService.UpdateStatus(ctx, order.ID, order.Status, errorMessage); err != nil {
			return fmt.Errorf("process order terminal status: %w", err)
		}
		return nil
	}

	switch order.OrderBy {
	case domain.OrderByAmount:
		qty := new(order.Amount.Div(*price))
		order.Quantity = qty
		if err := p.orderService.UpdateProcessingResult(ctx, order.ID, price, order.Amount, order.Quantity, order.Status); err != nil {
			return fmt.Errorf("process order update processing result: %w", err)
		}
	case domain.OrderByQuantity:
		amount := new(order.Quantity.Mul(*price))
		order.Amount = amount
		if err := p.orderService.UpdateProcessingResult(ctx, order.ID, price, order.Amount, order.Quantity, order.Status); err != nil {
			return fmt.Errorf("process order update processing result: %w", err)
		}
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
