package mappers

import (
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/google/uuid"
)

func OrderFromProto(proto *generated.Order) *domain.Order {
	return &domain.Order{
		ID:             uuidFromProto(proto.GetId()),
		AccountID:      uuidFromProto(proto.GetAccountId()),
		IdempotencyKey: uuidFromProto(proto.GetIdempotencyKey()),
		InstrumentID:   uuidFromProto(proto.GetInstrumentId()),
		Amount:         proto.GetAmount(),
		Status:         orderStatusFromProto(proto.GetStatus()),
	}
}

func OrderFromCreateOrderRequest(request *generated.CreateOrderRequest, time time.Time) *domain.Order {
	entity := &domain.Order{
		ID:             uuid.New(),
		AccountID:      uuid.MustParse(request.GetAccountId().GetValue()),
		IdempotencyKey: uuid.MustParse(request.GetIdempotencyKey().GetValue()),
		InstrumentID:   uuid.MustParse(request.GetInstrumentId().GetValue()),
		Amount:         request.GetAmount(),
		Status:         domain.OrderStatusNew,
		UpdatedAt:      time,
		CreatedAt:      time,
	}
	return entity
}

func uuidFromProto(proto *generated.UUID) uuid.UUID {
	if proto == nil {
		return uuid.Nil
	}
	return uuid.MustParse(proto.GetValue())
}

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
		return domain.OrderStatusNew
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

func uuidToProto(uuid uuid.UUID) *generated.UUID {
	return &generated.UUID{
		Value: uuid.String(),
	}
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
