package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cathudson/order-service/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib" // import pgx driver
)

func NewPostgresDB(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}
