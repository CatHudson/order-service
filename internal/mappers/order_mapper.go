package mappers

import (
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/utils"
	"github.com/google/uuid"
)

/*
------------------------From proto
*/

func OrderFromCreateOrderRequest(request *generated.CreateOrderRequest) *domain.Order {
	now := time.Now()
	entity := &domain.Order{
		ID:             utils.UUID(),
		AccountID:      uuid.MustParse(request.GetAccountId().GetValue()),
		IdempotencyKey: uuid.MustParse(request.GetIdempotencyKey().GetValue()),
		InstrumentID:   uuid.MustParse(request.GetInstrumentId().GetValue()),
		Side:           orderSideFromProto(request.GetSide()),
		OrderBy:        "",
		Amount:         nil,
		Price:          nil,
		Quantity:       nil,
		Status:         domain.OrderStatusNew,
		ErrorMessage:   nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	switch request.GetAmount().(type) {
	case *generated.CreateOrderRequest_Quantity:
		entity.Quantity = new(utils.DecimalFromProto(request.GetQuantity()))
		entity.OrderBy = domain.OrderByQuantity
	case *generated.CreateOrderRequest_MonetaryValue:
		entity.Amount = new(utils.MoneyToDecimal(request.GetMonetaryValue()))
		entity.OrderBy = domain.OrderByAmount
	}

	return entity
}

func orderSideFromProto(side generated.OrderSide) domain.OrderSide {
	switch side {
	case generated.OrderSide_ORDER_SIDE_BUY:
		return domain.OrderSideBuy
	case generated.OrderSide_ORDER_SIDE_SELL:
		return domain.OrderSideSell
	case generated.OrderSide_ORDER_SIDE_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}

/*
------------------------To proto
*/

func OrderToProto(entity *domain.Order) *generated.Order {
	proto := &generated.Order{
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
		proto.Quantity = utils.DecimalToProto(*entity.Quantity)
	}
	if entity.Amount != nil {
		proto.Amount = utils.DecimalToMoney(*entity.Amount)
	}
	if entity.Price != nil {
		proto.Price = utils.DecimalToMoney(*entity.Price)
	}

	return proto
}

func orderByToProto(orderBy domain.OrderBy) generated.OrderBy {
	switch orderBy {
	case domain.OrderByQuantity:
		return generated.OrderBy_ORDER_BY_QUANTITY
	case domain.OrderByAmount:
		return generated.OrderBy_ORDER_BY_AMOUNT
	default:
		return generated.OrderBy_ORDER_BY_UNSPECIFIED
	}
}

func orderSideToProto(entity domain.OrderSide) generated.OrderSide {
	switch entity {
	case domain.OrderSideBuy:
		return generated.OrderSide_ORDER_SIDE_BUY
	case domain.OrderSideSell:
		return generated.OrderSide_ORDER_SIDE_SELL
	default:
		return generated.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func uuidToProto(id uuid.UUID) *generated.UUID {
	return &generated.UUID{Value: id.String()}
}

func OrderStatusToProto(status domain.OrderStatus) generated.OrderStatus {
	switch status {
	case domain.OrderStatusNew:
		return generated.OrderStatus_ORDER_STATUS_NEW
	case domain.OrderStatusPending:
		return generated.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusSuccess:
		return generated.OrderStatus_ORDER_STATUS_SUCCESSFUL
	case domain.OrderStatusFailed:
		return generated.OrderStatus_ORDER_STATUS_FAILED
	case domain.OrderStatusCanceled:
		return generated.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return generated.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
