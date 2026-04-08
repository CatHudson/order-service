package app

import (
	"context"
	"time"

	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/mappers"
	"github.com/cathudson/order-service/internal/store"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	entity, err := h.orderStore.Create(ctx, mappers.OrderFromCreateOrderRequest(request, h.now()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in store: %v", err)
	}

	return &generated.CreateOrderResponse{
		Id:     &generated.UUID{Value: entity.ID.String()},
		Status: mappers.OrderStatusToProto(entity.Status),
	}, nil
}

func (h *createOrderHandler) validate(request *generated.CreateOrderRequest) error {
	if request.GetAmount() <= 0 {
		return status.Errorf(codes.InvalidArgument, "invalid amount: %v", request.GetAmount())
	}
	if request.GetSide() == generated.OrderSide_ORDER_SIDE_UNSPECIFIED {
		return status.Errorf(codes.InvalidArgument, "invalid side: %v", request.GetSide())
	}
	if _, err := uuid.Parse(request.GetAccountId().GetValue()); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid account id: %v", err)
	}
	if _, err := uuid.Parse(request.GetIdempotencyKey().GetValue()); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid idempotency key: %v", err)
	}
	if _, err := uuid.Parse(request.GetInstrumentId().GetValue()); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid instrument id: %v", err)
	}
	return nil
}
