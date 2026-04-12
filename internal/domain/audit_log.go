package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderAuditLog struct {
	ID        uuid.UUID `db:"id"`
	OrderID   uuid.UUID `db:"order_id"`
	Action    Action    `db:"action"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
}

type Action string

const (
	OrderCreated  Action = "create"
	OrderUpdated  Action = "update"
	OrderFinished Action = "finished"
	OrderCanceled Action = "cancel"
	OrderFailed   Action = "failed"
)
