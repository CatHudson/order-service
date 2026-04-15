package store

import (
	"context"
	"time"
)

type ReplicaHealthChecker struct {
	conn *ConnContainer
}

func NewReplicaHealthChecker(conn *ConnContainer) *ReplicaHealthChecker {
	return &ReplicaHealthChecker{conn: conn}
}

func (r *ReplicaHealthChecker) Run(ctx context.Context) {
	const tickInterval = 10 * time.Second
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	r.conn.replicaHealthy.Store(r.conn.replica.PingContext(ctx) == nil)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := r.conn.replica.PingContext(ctx)
			if err != nil {
				r.conn.replicaHealthy.Store(false)
			} else {
				r.conn.replicaHealthy.Store(true)
			}
		}
	}
}
