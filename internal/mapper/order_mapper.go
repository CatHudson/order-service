package mapper

import (
	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/util"
	"github.com/google/uuid"
)

/*
------------------------From proto
*/

func CreateOrderEventFromGrpc(request *proto.CreateOrderRequest) *proto.CreateOrderEvent {
	entity := &proto.CreateOrderEvent{
		Id:             uuidToProto(util.UUID()),
		AccountId:      request.GetAccountId(),
		IdempotencyKey: request.GetIdempotencyKey(),
		InstrumentId:   request.GetInstrumentId(),
		OrderBy:        proto.OrderBy_ORDER_BY_UNSPECIFIED,
		Quantity:       nil,
		Amount:         nil,
		Side:           request.GetSide(),
	}

	switch request.GetAmount().(type) {
	case *proto.CreateOrderRequest_MonetaryValue:
		entity.Amount = request.GetMonetaryValue()
		entity.OrderBy = proto.OrderBy_ORDER_BY_AMOUNT
	case *proto.CreateOrderRequest_Quantity:
		entity.Quantity = request.GetQuantity()
		entity.OrderBy = proto.OrderBy_ORDER_BY_QUANTITY
	}

	return entity
}

func OrderByFromProto(side proto.OrderBy) domain.OrderBy {
	switch side {
	case proto.OrderBy_ORDER_BY_AMOUNT:
		return domain.OrderByAmount
	case proto.OrderBy_ORDER_BY_QUANTITY:
		return domain.OrderByQuantity
	case proto.OrderBy_ORDER_BY_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}

func OrderSideFromProto(side proto.OrderSide) domain.OrderSide {
	switch side {
	case proto.OrderSide_ORDER_SIDE_BUY:
		return domain.OrderSideBuy
	case proto.OrderSide_ORDER_SIDE_SELL:
		return domain.OrderSideSell
	case proto.OrderSide_ORDER_SIDE_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}

/*
------------------------To proto
*/

func OrderToProto(entity *domain.Order) *proto.Order {
	proto := &proto.Order{
		Id:             uuidToProto(entity.ID),
		AccountId:      uuidToProto(entity.AccountID),
		IdempotencyKey: uuidToProto(entity.IdempotencyKey),
		InstrumentId:   uuidToProto(entity.InstrumentID),
		OrderBy:        orderByToProto(entity.OrderBy),
		Quantity:       nil,
		Amount:         nil,
		Price:          nil,
		Side:           orderSideToProto(entity.Side),
		Status:         OrderStatusToProto(entity.Status),
		ErrorMessage:   entity.ErrorMessage,
	}

	if entity.Quantity != nil {
		proto.Quantity = util.DecimalToProto(*entity.Quantity)
	}
	if entity.Amount != nil {
		proto.Amount = util.DecimalToMoney(*entity.Amount)
	}
	if entity.Price != nil {
		proto.Price = util.DecimalToMoney(*entity.Price)
	}

	return proto
}

func orderByToProto(orderBy domain.OrderBy) proto.OrderBy {
	switch orderBy {
	case domain.OrderByQuantity:
		return proto.OrderBy_ORDER_BY_QUANTITY
	case domain.OrderByAmount:
		return proto.OrderBy_ORDER_BY_AMOUNT
	default:
		return proto.OrderBy_ORDER_BY_UNSPECIFIED
	}
}

func orderSideToProto(entity domain.OrderSide) proto.OrderSide {
	switch entity {
	case domain.OrderSideBuy:
		return proto.OrderSide_ORDER_SIDE_BUY
	case domain.OrderSideSell:
		return proto.OrderSide_ORDER_SIDE_SELL
	default:
		return proto.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func uuidToProto(id uuid.UUID) *proto.UUID {
	return &proto.UUID{Value: id.String()}
}

func OrderStatusToProto(status domain.OrderStatus) proto.OrderStatus {
	switch status {
	case domain.OrderStatusNew:
		return proto.OrderStatus_ORDER_STATUS_NEW
	case domain.OrderStatusPending:
		return proto.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusSuccess:
		return proto.OrderStatus_ORDER_STATUS_SUCCESSFUL
	case domain.OrderStatusFailed:
		return proto.OrderStatus_ORDER_STATUS_FAILED
	case domain.OrderStatusCanceled:
		return proto.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return proto.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
