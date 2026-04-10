package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//go:generate moq -rm -out gen/order_repository_mock.go -pkg storegen . OrderStore

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)
}

type orderStore struct {
	conn *sqlx.DB
}

func NewOrderStore(conn *sqlx.DB) OrderStore {
	return &orderStore{conn: conn}
}

func (s *orderStore) Create(ctx context.Context, order *domain.Order) error {
	const query = `INSERT INTO orders (
                    id, 
                    account_id, 
                    idempotency_key, 
                    instrument_id,
                    order_by, 
                    side, 
                    amount, 
                    quantity, 
                    price, 
                    status,
                    error_message,
                    created_at, 
                    updated_at) 
	VALUES (:id, 
	        :account_id, 
	        :idempotency_key, 
	        :instrument_id,
            :order_by, 
	        :side, 
	        :amount, 
	        :quantity, 
	        :price,
            :status, 
	        :error_message, 
	        :created_at, 
	        :updated_at)`

	_, err := s.conn.NamedExecContext(ctx, query, order)
	if err != nil {
		return err
	}

	return nil
}

func (s *orderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	const query = `SELECT * FROM orders WHERE id = $1`

	var order domain.Order
	err := s.conn.GetContext(ctx, &order, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (s *orderStore) GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	const query = `SELECT * FROM orders WHERE status = $1`

	var orders []*domain.Order
	err := s.conn.SelectContext(ctx, &orders, query, status)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
