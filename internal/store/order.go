package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/google/uuid"
)

//go:generate moq -rm -out gen/order_repository_mock.go -pkg storegen . OrderStore

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)
}

type orderStore struct {
	conn *sql.DB
}

func NewOrderStore(conn *sql.DB) OrderStore {
	return &orderStore{conn: conn}
}

// nolint: revive // Will implement later
func (s *orderStore) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return nil, errors.New("not implemented")
}

// nolint: revive // Will implement later
func (s *orderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return nil, errors.New("not implemented")
}

// nolint: revive // Will implement later
func (s *orderStore) GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	return nil, errors.New("not implemented")
}
