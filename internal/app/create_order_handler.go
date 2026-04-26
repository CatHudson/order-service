package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/mapper"
	"github.com/cathudson/order-service/internal/producer"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/store"
	"github.com/cathudson/order-service/internal/util"
	"github.com/google/uuid"
	sdecimal "github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const waitFor = 2 * time.Second

type createOrderHandler struct {
	createOrderProducer producer.CreateOrderProducer
	orderResultStore    store.OrderResultStore
	orderStore          store.OrderStore
	logger              *zap.SugaredLogger
	now                 func() time.Time
}

func newCreateOrderHandler(producer producer.CreateOrderProducer, orderResultStore store.OrderResultStore, orderStore store.OrderStore, logger *zap.SugaredLogger) *createOrderHandler {
	return &createOrderHandler{createOrderProducer: producer, orderResultStore: orderResultStore, orderStore: orderStore, logger: logger, now: time.Now}
}

func (h *createOrderHandler) handle(ctx context.Context, request *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	err := h.validate(request)
	if err != nil {
		return nil, err
	}

	const timeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	entity := mapper.CreateOrderEventFromGrpc(request)
	err = h.createOrderProducer.Produce(ctx, entity)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("create order producer failed: %v", err))
	}

	responseEntity := responseTemplate(entity)
	orderResult, err := h.orderResultStore.Get(ctx, entity.GetId().GetValue(), waitFor)
	if err != nil || orderResult == nil {
		if !errors.Is(err, store.ErrOrderResultNotFound) {
			h.logger.Warnw("redis error, falling back to Postgres",
				"orderID", entity.GetId().GetValue(), "error", err)
		}
		return h.orderResultFromStore(ctx, responseEntity, entity), nil
	}

	responseEntity = enrichOrderResponseFromCache(responseEntity, orderResult)

	return &proto.CreateOrderResponse{Order: responseEntity}, nil
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

func enrichOrderResponseFromCache(responseEntity *proto.Order, event *proto.OrderResultEvent) *proto.Order {
	responseEntity.Price = event.GetPrice()
	responseEntity.Status = event.GetStatus()
	responseEntity.ErrorMessage = event.ErrorMessage
	responseEntity.Quantity = event.GetQuantity()
	responseEntity.Amount = event.GetAmount()

	return responseEntity
}

func (h *createOrderHandler) orderResultFromStore(ctx context.Context, responseEntity *proto.Order, event *proto.CreateOrderEvent) *proto.CreateOrderResponse {
	pOrder, err := h.orderStore.GetByID(ctx, uuid.MustParse(event.GetId().GetValue()))
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			h.logger.Infow("order not yet in Postgres, returning NEW", "orderID", event.GetId().GetValue())
		} else {
			h.logger.Warnw("Postgres fallback failed", "orderID", event.GetId().GetValue(), "error", err)
		}
		return &proto.CreateOrderResponse{Order: responseEntity}
	}
	responseEntity.Price = util.DecimalToMoney(pOrder.Price)
	responseEntity.Status = mapper.OrderStatusToProto(pOrder.Status)
	responseEntity.ErrorMessage = pOrder.ErrorMessage
	responseEntity.Quantity = util.DecimalToProto(pOrder.Quantity)
	responseEntity.Amount = util.DecimalToMoney(pOrder.Amount)

	return &proto.CreateOrderResponse{Order: responseEntity}
}

func responseTemplate(entity *proto.CreateOrderEvent) *proto.Order {
	return &proto.Order{
		Id:           entity.GetId(),
		AccountId:    entity.GetAccountId(),
		InstrumentId: entity.GetInstrumentId(),
		OrderBy:      entity.GetOrderBy(),
		Quantity:     entity.GetQuantity(),
		Amount:       entity.GetAmount(),
		Price:        nil,
		Side:         entity.GetSide(),
		Status:       proto.OrderStatus_ORDER_STATUS_NEW,
		ErrorMessage: nil,
	}
}
