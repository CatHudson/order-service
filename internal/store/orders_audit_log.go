package store

import (
	"context"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type OrdersAuditLogStore interface {
	Create(ctx context.Context, auditLog *domain.OrderAuditLog) error
}

type ordersAuditLogStore struct {
	conn DBGetter
}

func NewOrdersAuditLogStore(db DBGetter) OrdersAuditLogStore {
	return &ordersAuditLogStore{conn: db}
}

func (s *ordersAuditLogStore) Create(ctx context.Context, auditLog *domain.OrderAuditLog) error {
	const query = `INSERT INTO orders_audit_log (
                              id,
                              order_id,
                              action,
                              payload,
                              created_at)
    				VALUES (
    				        :id, 
    				        :order_id, 
    				        :action, 
    				        :payload, 
    				        :created_at)`

	_, err := sqlx.NamedExecContext(ctx, s.conn.Primary(ctx), query, auditLog)
	if err != nil {
		return err
	}

	return nil
}
