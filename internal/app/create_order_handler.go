package app

import (
	"context"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/mapper"
	"github.com/cathudson/order-service/internal/producer"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/google/uuid"
	sdecimal "github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type createOrderHandler struct {
	createOrderProducer producer.CreateOrderProducer
	now                 func() time.Time
}

func newCreateOrderHandler(producer producer.CreateOrderProducer) *createOrderHandler {
	return &createOrderHandler{createOrderProducer: producer, now: time.Now}
}

func (h *createOrderHandler) handle(ctx context.Context, request *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	err := h.validate(request)
	if err != nil {
		return nil, err
	}

	entity := mapper.CreateOrderEventFromGrpc(request)
	err = h.createOrderProducer.Produce(ctx, entity)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("create order producer failed: %v", err))
	}

	return &proto.CreateOrderResponse{
		Id:     entity.Id,
		Status: proto.OrderStatus_ORDER_STATUS_NEW,
	}, nil
}

func (h *createOrderHandler) validate(request *proto.CreateOrderRequest) error {
	if request.GetMonetaryValue().GetUnits() < 0 || request.GetMonetaryValue().GetNanos() < 0 {
		return status.Errorf(codes.InvalidArgument, "invalid monetary value: %v", request.GetMonetaryValue())
	}
	if request.GetQuantity() != nil {
		decimal, err := sdecimal.NewFromString(request.GetQuantity().GetValue())
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid quantity: %v", err)
		}
		if decimal.LessThan(sdecimal.Zero) {
			return status.Errorf(codes.InvalidArgument, "negative quantity not allowed: %v", decimal)
		}
	}
	if request.GetSide() == proto.OrderSide_ORDER_SIDE_UNSPECIFIED {
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
