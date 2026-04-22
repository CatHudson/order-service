package consumer

import (
	"context"
	"fmt"

	pe "github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/store"
)

type OrderResultConsumer struct {
	orderResultStore store.OrderResultStore
}

func NewOrderResultConsumer(orderResultStore store.OrderResultStore) *OrderResultConsumer {
	return &OrderResultConsumer{orderResultStore: orderResultStore}
}

func (c *OrderResultConsumer) Handle(ctx context.Context, event *pe.OrderResultEvent) error {
	err := c.orderResultStore.Save(ctx, event.GetId().GetValue(), event)
	if err != nil {
		return fmt.Errorf("order-result consumer: %w", err)
	}
	return nil
}
