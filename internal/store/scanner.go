package store

import "github.com/cathudson/order-service/internal/domain"

type scanner interface {
	Scan(dest ...any) error
}

func scanOrder(s scanner) (*domain.Order, error) {
	var o domain.Order
	err := s.Scan(
		&o.ID, &o.AccountID, &o.IdempotencyKey, &o.InstrumentID,
		&o.OrderBy, &o.Side, &o.Amount, &o.Quantity, &o.Price,
		&o.Status, &o.ErrorMessage, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
