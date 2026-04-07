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
	done       chan struct{}
}

func NewReporter(orderStore store.OrderStore, log *zap.SugaredLogger) *Reporter {
	return &Reporter{
		orderStore: orderStore,
		done:       make(chan struct{}),
		log:        log,
	}
}

func (r *Reporter) Run() {
	const interval = 5 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			orders, err := r.orderStore.GetAllByStatus(context.TODO(), domain.OrderStatusPending)
			if err != nil {
				r.log.Errorf(
					"error in store: %v", err)
			} else {
				r.log.Infof("found %d orders in pending", len(orders))
			}
		}
	}
}

func (r *Reporter) Stop() {
	close(r.done)
}
