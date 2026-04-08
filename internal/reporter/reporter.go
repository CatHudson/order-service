package reporter

import (
	"context"
	"sync"
	"time"

	"github.com/cathudson/order-service/internal/domain"
	"github.com/cathudson/order-service/internal/store"
	"go.uber.org/zap"
)

type Reporter struct {
	ctx        context.Context
	log        *zap.SugaredLogger
	orderStore store.OrderStore
	sync.Once
}

func NewReporter(ctx context.Context, orderStore store.OrderStore, log *zap.SugaredLogger) *Reporter {
	return &Reporter{
		ctx:        ctx,
		orderStore: orderStore,
		log:        log,
	}
}

func (r *Reporter) Run() {
	const queryTimeout = 4 * time.Second
	const tickInterval = 5 * time.Second
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			queryCtx, cancel := context.WithTimeout(r.ctx, queryTimeout)
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
