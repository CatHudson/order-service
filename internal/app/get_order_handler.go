package app

import (
	"context"
	"errors"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/mappers"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/store"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type getOrderHandler struct {
	orderStore store.OrderStore
}

func newGetOrderHandler(orderStore store.OrderStore) *getOrderHandler {
	return &getOrderHandler{orderStore: orderStore}
}

func (h *getOrderHandler) handle(ctx context.Context, request *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	err := h.validate(request)
	if err != nil {
		return nil, err
	}

	entity, err := h.orderStore.GetByID(ctx, uuid.MustParse(request.GetId().GetValue()))
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "store error: %v", err)
	}

	return &proto.GetOrderResponse{Order: mappers.OrderToProto(entity)}, nil
}

func (h *getOrderHandler) validate(request *proto.GetOrderRequest) error {
	_, err := uuid.Parse(request.GetId().GetValue())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}
	return nil
}
