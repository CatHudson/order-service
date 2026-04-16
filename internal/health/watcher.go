package health

import (
	"context"
	"time"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	ServiceOverall = ""
	ServicePrimary = "postgres-primary"
	ServiceReplica = "postgres-replica"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type Watcher struct {
	server          *health.Server
	primary         Pinger
	replica         Pinger
	replicaCallback func(bool)
}

func NewWatcher(server *health.Server, primary Pinger, replica Pinger, replicaCallback func(bool)) *Watcher {
	return &Watcher{server: server, primary: primary, replica: replica, replicaCallback: replicaCallback}
}

func (c *Watcher) Run(ctx context.Context) {
	const tickInterval = time.Second * 10
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	c.check(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx)
		}
	}
}

func (c *Watcher) check(ctx context.Context) {
	primaryHealthy := c.primary.PingContext(ctx) == nil
	if primaryHealthy {
		c.server.SetServingStatus(ServiceOverall, grpc_health_v1.HealthCheckResponse_SERVING)
		c.server.SetServingStatus(ServicePrimary, grpc_health_v1.HealthCheckResponse_SERVING)
	} else {
		c.server.SetServingStatus(ServiceOverall, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		c.server.SetServingStatus(ServicePrimary, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	if c.replica != nil {
		replicaHealthy := c.replica.PingContext(ctx) == nil
		if replicaHealthy {
			c.server.SetServingStatus(ServiceReplica, grpc_health_v1.HealthCheckResponse_SERVING)
		} else {
			c.server.SetServingStatus(ServiceReplica, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		}
		c.replicaCallback(replicaHealthy)
	} else {
		c.server.SetServingStatus(ServiceReplica, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		c.replicaCallback(false)
	}
}
