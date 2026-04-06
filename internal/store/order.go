package store

import (
	"context"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/google/uuid"
)

//go:generate moq -rm -out gen/order_repository_mock.go -pkg storegen . OrderStore

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}

type orderStore struct {
	// TODO: implement
}

func NewOrderStore() OrderStore {
	return &orderStore{}
}

// nolint: nilnil, revive // Will implement later
func (s *orderStore) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	panic("not implemented")
}

// nolint: nilnil, revive // Will implement later
func (s *orderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	panic("not implemented")
}
