package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID             uuid.UUID        `db:"id"`
	AccountID      uuid.UUID        `db:"account_id"`
	IdempotencyKey uuid.UUID        `db:"idempotency_key"`
	InstrumentID   uuid.UUID        `db:"instrument_id"`
	Side           OrderSide        `db:"side"`
	OrderBy        OrderBy          `db:"order_by"`
	Amount         *decimal.Decimal `db:"amount"`
	Quantity       *decimal.Decimal `db:"quantity"`
	Price          *decimal.Decimal `db:"price"`
	Status         OrderStatus      `db:"status"`
	ErrorMessage   *string          `db:"error_message"`
	UpdatedAt      time.Time        `db:"updated_at"`
	CreatedAt      time.Time        `db:"created_at"`
}

type OrderBy string

const (
	OrderByAmount   OrderBy = "AMOUNT"
	OrderByQuantity OrderBy = "QUANTITY"
)

type OrderStatus string

const (
	OrderStatusNew      OrderStatus = "NEW"
	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusSuccess  OrderStatus = "SUCCESS"
	OrderStatusFailed   OrderStatus = "FAILED"
	OrderStatusCanceled OrderStatus = "CANCELED"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)
