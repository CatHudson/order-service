package store

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

type DB interface {
	sqlx.ExtContext
}

type DBGetter interface {
	Primary(ctx context.Context) DB
	Replica() DB
}

type txKey struct{}

type ConnContainer struct {
	primary        *sqlx.DB
	replica        *sqlx.DB
	replicaHealthy atomic.Bool
}

func NewConnContainer(primary, replica *sqlx.DB) *ConnContainer {
	return &ConnContainer{primary: primary, replica: replica, replicaHealthy: atomic.Bool{}}
}

func (c *ConnContainer) Primary(ctx context.Context) DB {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return c.primary
}

func (c *ConnContainer) Replica() DB {
	if c.replica != nil && c.replicaHealthy.Load() {
		return c.replica
	}
	return c.primary
}

func (c *ConnContainer) SetReplicaHealthy(isHealthy bool) {
	c.replicaHealthy.Store(isHealthy)
}

type DBTransactor interface {
	Exec(ctx context.Context, fn func(txCtx context.Context) error) error
}

type Transactor struct {
	primary *sqlx.DB
}

func NewTransactor(primary *sqlx.DB) *Transactor {
	return &Transactor{primary: primary}
}

func (t *Transactor) Exec(ctx context.Context, fn func(txCtx context.Context) error) error {
	if getTx(ctx) != nil {
		return fn(ctx)
	}

	transaction, err := t.primary.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	txCtx := context.WithValue(ctx, txKey{}, transaction)
	if err = fn(txCtx); err != nil {
		return fmt.Errorf("execute transaction: %w", err)
	}

	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func getTx(ctx context.Context) *sqlx.Tx {
	val := ctx.Value(txKey{})
	if val == nil {
		return nil
	}
	if tx, ok := val.(*sqlx.Tx); ok {
		return tx
	}
	return nil
}
