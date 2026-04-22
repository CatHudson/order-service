package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

//go:generate moq -rm -out gen/order_repository_mock.go -pkg storegen . OrderStore

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	UpdateProcessingResult(ctx context.Context, id uuid.UUID, price, amount, quantity *decimal.Decimal, status domain.OrderStatus) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)
}

type orderStore struct {
	conn DBGetter
}

func NewOrderStore(db DBGetter) OrderStore {
	return &orderStore{conn: db}
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

	_, err := sqlx.NamedExecContext(ctx, s.conn.Primary(ctx), query, order)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %w", domain.ErrOrderAlreadyExists, err)
		}
		return err
	}

	return nil
}

func (s *orderStore) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	const query = `
		UPDATE orders 
		SET status = $1, updated_at = $2 
		WHERE id = $3 
		AND updated_at <= $2 
		AND status NOT IN ('SUCCESS', 'FAILED', 'CANCELED')`

	_, err := s.conn.Primary(ctx).ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

func (s *orderStore) UpdateProcessingResult(ctx context.Context, id uuid.UUID, price, amount, quantity *decimal.Decimal, status domain.OrderStatus) error {
	const query = `
		UPDATE orders 
		SET price = $1, amount = $2, quantity = $3, status $4, updated_at = $5 
		WHERE id = $6
			AND updated_at <= $5 
			AND status NOT IN ('SUCCESS', 'FAILED', 'CANCELED')`

	_, err := s.conn.Primary(ctx).ExecContext(ctx, query, price, amount, quantity, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update order data: %w", err)
	}

	return nil
}

func (s *orderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	const query = `SELECT * FROM orders WHERE id = $1`

	var order domain.Order
	err := sqlx.GetContext(ctx, s.conn.Replica(), &order, query, id)
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
	err := sqlx.SelectContext(ctx, s.conn.Replica(), &orders, query, status)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
