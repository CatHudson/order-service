package service

import (
	"context"
	"fmt"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/mapper"
	"github.com/cathudson/order-service/internal/store"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderService struct {
	orderStore          store.OrderStore
	ordersAuditLogStore store.OrdersAuditLogStore
	tx                  store.DBTransactor
}

func NewOrderService(tx store.DBTransactor, orderStore store.OrderStore, ordersAuditLogStore store.OrdersAuditLogStore) *OrderService {
	return &OrderService{
		orderStore:          orderStore,
		ordersAuditLogStore: ordersAuditLogStore,
		tx:                  tx,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	return s.tx.Exec(ctx, func(txCtx context.Context) error {
		if err := s.orderStore.Create(txCtx, order); err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		if err := s.ordersAuditLogStore.Create(txCtx, mapper.OrderCreatedAuditLog(order)); err != nil {
			return fmt.Errorf("create audit log: %w", err)
		}

		return nil
	})
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderStore.GetByID(ctx, id)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus, errorMessage *string) error {
	return s.orderStore.UpdateStatus(ctx, id, status, errorMessage)
}

func (s *OrderService) UpdateProcessingResult(ctx context.Context, id uuid.UUID, price, amount, quantity *decimal.Decimal, status domain.OrderStatus) error {
	return s.orderStore.UpdateProcessingResult(ctx, id, price, amount, quantity, status)
}
