package consumer

import (
	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/mapper"
	pe "github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/task"
	"github.com/cathudson/order-service/internal/util"
	"github.com/google/uuid"
)

func taskFromEvent(event *pe.CreateOrderEvent) *task.CreateOrderTask {
	entity := &task.CreateOrderTask{
		ID:             uuid.MustParse(event.GetId().GetValue()),
		AccountID:      uuid.MustParse(event.GetAccountId().GetValue()),
		IdempotencyKey: uuid.MustParse(event.GetIdempotencyKey().GetValue()),
		InstrumentID:   uuid.MustParse(event.GetInstrumentId().GetValue()),
		OrderBy:        mapper.OrderByFromProto(event.GetOrderBy()),
		Quantity:       nil,
		Amount:         nil,
		OrderSide:      mapper.OrderSideFromProto(event.GetSide()),
	}

	if entity.OrderBy == domain.OrderByQuantity {
		entity.Quantity = new(util.DecimalFromProto(event.GetQuantity()))
	} else {
		entity.Amount = new(util.MoneyToDecimal(event.GetAmount()))
	}

	return entity
}
