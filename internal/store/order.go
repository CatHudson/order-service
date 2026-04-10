package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (s *orderStore) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	now := time.Now()
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
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
	RETURNING id, account_id, idempotency_key, instrument_id,
            order_by, side, amount, quantity, price,
            status, error_message, created_at, updated_at`

	row := s.conn.QueryRowContext(ctx, query,
		order.ID, order.AccountID, order.IdempotencyKey, order.InstrumentID,
		order.OrderBy, order.Side, order.Amount, order.Quantity, order.Price,
		order.Status, order.ErrorMessage, now, now)

	return scanOrder(row)
}

func (s *orderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	const query = `SELECT
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
        updated_at  
	FROM orders WHERE id = $1`

	row := s.conn.QueryRowContext(ctx, query, id)
	order, err := scanOrder(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return order, nil
}

func (s *orderStore) GetAllByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	const query = `SELECT
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
        updated_at
	FROM orders WHERE status = $1`

	rows, err := s.conn.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("get all orders by status: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	var order *domain.Order
	for rows.Next() {
		order, err = scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders by status: %w", err)
	}

	return orders, nil
}
