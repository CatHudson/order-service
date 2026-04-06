package app

import (
	"context"
	"fmt"

	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/mappers"
	"github.com/cathudson/order-service/internal/store"
	"github.com/google/uuid"
)

type getOrderHandler struct {
	orderStore store.OrderStore
}

func newGetOrderHandler(orderStore store.OrderStore) *getOrderHandler {
	return &getOrderHandler{orderStore: orderStore}
}

func (h *getOrderHandler) handle(ctx context.Context, request *generated.GetOrderRequest) (*generated.GetOrderResponse, error) {
	err := h.validate(request)
	if err != nil {
		return nil, err
	}

	entity, err := h.orderStore.GetByID(ctx, uuid.MustParse(request.GetId().GetValue()))
	if err != nil {
		return nil, err
	}

	return &generated.GetOrderResponse{Order: mappers.OrderToProto(entity)}, nil
}

func (h *getOrderHandler) validate(request *generated.GetOrderRequest) error {
	_, err := uuid.Parse(request.GetId().GetValue())
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	return nil
}
