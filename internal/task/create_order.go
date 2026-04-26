package task

import (
	"github.com/cathudson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const CreateOrderTaskType = "orders:create_order"

type CreateOrderTask struct {
	ID           uuid.UUID
	AccountID    uuid.UUID
	InstrumentID uuid.UUID
	OrderBy      domain.OrderBy
	Quantity     *decimal.Decimal
	Amount       *decimal.Decimal
	OrderSide    domain.OrderSide
}
