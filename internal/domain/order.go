package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	IdempotencyKey uuid.UUID
	InstrumentID   uuid.UUID
	Amount         int64
	Status         OrderStatus
	UpdatedAt      time.Time
	CreatedAt      time.Time
}

type OrderStatus string

const (
	OrderStatusNew     OrderStatus = "NEW"
	OrderStatusPending OrderStatus = "PENDING"
	OrderStatusSuccess OrderStatus = "SUCCESS"
	OrderStatusFailed  OrderStatus = "FAILED"
)
