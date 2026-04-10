package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	IdempotencyKey uuid.UUID
	InstrumentID   uuid.UUID
	Side           OrderSide
	OrderBy        OrderBy
	Amount         *decimal.Decimal
	Quantity       *decimal.Decimal
	Price          *decimal.Decimal
	Status         OrderStatus
	ErrorMessage   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
