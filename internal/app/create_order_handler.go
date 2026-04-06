package app

import (
	"context"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/mappers"
	"github.com/cathudson/order-service/internal/store"
	"github.com/google/uuid"
)

type createOrderHandler struct {
	orderStore store.OrderStore
	now        func() time.Time
}

func newCreateOrderHandler(orderStore store.OrderStore) *createOrderHandler {
	return &createOrderHandler{orderStore: orderStore, now: time.Now}
}

func (h *createOrderHandler) handle(ctx context.Context, request *generated.CreateOrderRequest) (*generated.CreateOrderResponse, error) {
	err := h.validate(request)
	if err != nil {
		return nil, err
	}

	entity := mappers.OrderFromCreateOrderRequest(request, h.now())
	entity, err = h.orderStore.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	return &generated.CreateOrderResponse{
		Id:     &generated.UUID{Value: entity.ID.String()},
		Status: mappers.OrderStatusToProto(entity.Status),
	}, nil
}

func (h *createOrderHandler) validate(request *generated.CreateOrderRequest) error {
	if request.GetAmount() == 0 {
		return fmt.Errorf("amount must not be zero")
	}
	if _, err := uuid.Parse(request.GetAccountId().GetValue()); err != nil {
		return fmt.Errorf("invalid account id: %w", err)
	}
	if _, err := uuid.Parse(request.GetIdempotencyKey().GetValue()); err != nil {
		return fmt.Errorf("invalid idempotency key: %w", err)
	}
	if _, err := uuid.Parse(request.GetInstrumentId().GetValue()); err != nil {
		return fmt.Errorf("invalid instrument id: %w", err)
	}
	return nil
}
