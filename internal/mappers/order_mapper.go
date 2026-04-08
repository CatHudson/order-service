package mappers

import (
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/google/uuid"
)

func OrderFromCreateOrderRequest(request *generated.CreateOrderRequest, time time.Time) *domain.Order {
	entity := &domain.Order{
		ID:             uuid.New(),
		AccountID:      uuid.MustParse(request.GetAccountId().GetValue()),
		IdempotencyKey: uuid.MustParse(request.GetIdempotencyKey().GetValue()),
		InstrumentID:   uuid.MustParse(request.GetInstrumentId().GetValue()),
		Side:           orderSideFromProto(request.GetSide()),
		Amount:         request.GetAmount(),
		Status:         domain.OrderStatusNew,
		UpdatedAt:      time,
		CreatedAt:      time,
	}
	return entity
}

// nolint: unused // will be used later
func orderStatusFromProto(status generated.OrderStatus) domain.OrderStatus {
	switch status {
	case generated.OrderStatus_ORDER_STATUS_NEW:
		return domain.OrderStatusNew
	case generated.OrderStatus_ORDER_STATUS_PENDING:
		return domain.OrderStatusPending
	case generated.OrderStatus_ORDER_STATUS_SUCCESSFUL:
		return domain.OrderStatusSuccess
	case generated.OrderStatus_ORDER_STATUS_FAILED:
		return domain.OrderStatusFailed
	case generated.OrderStatus_ORDER_STATUS_UNSPECIFIED:
		return domain.OrderStatusNew
	default:
		return ""
	}
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

func OrderToProto(entity *domain.Order) *generated.Order {
	return &generated.Order{
		Id:             uuidToProto(entity.ID),
		AccountId:      uuidToProto(entity.AccountID),
		IdempotencyKey: uuidToProto(entity.IdempotencyKey),
		InstrumentId:   uuidToProto(entity.InstrumentID),
		Amount:         entity.Amount,
		Status:         OrderStatusToProto(entity.Status),
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
	default:
		return generated.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
