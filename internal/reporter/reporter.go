package reporter

import (
	"context"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/store"
	"go.uber.org/zap"
)

type Reporter struct {
	log        *zap.SugaredLogger
	orderStore store.OrderStore
}

func NewReporter(orderStore store.OrderStore, log *zap.SugaredLogger) *Reporter {
	return &Reporter{
		orderStore: orderStore,
		log:        log,
	}
}

func (r *Reporter) Run(ctx context.Context) {
	const queryTimeout = 4 * time.Second
	const tickInterval = 5 * time.Second
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
			orders, err := r.orderStore.GetAllByStatus(queryCtx, domain.OrderStatusPending)
			if err != nil {
				r.log.Errorf(
					"error in store: %v", err)
			} else {
				r.log.Infof("found %d orders in pending", len(orders))
			}
			cancel()
		}
	}
}
