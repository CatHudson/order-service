package mappers

import (
	"encoding/json"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/utils"
)

func OrderCreatedAuditLog(order *domain.Order) *domain.OrderAuditLog {
	entity := &domain.OrderAuditLog{
		ID:        utils.UUID(),
		OrderID:   order.ID,
		Action:    domain.OrderCreated,
		Payload:   nil,
		CreatedAt: time.Now(),
	}
	payload, err := json.Marshal(order)
	if err == nil {
		entity.Payload = payload
	}
	// log in case of marshall error?
	return entity
}
