package app

import (
	"context"

	"github.com/cathudson/order-service/internal/generated"
)

type getHealthHandler struct{}

func newGetHealthHandler() *getHealthHandler {
	return &getHealthHandler{}
}

func (h *getHealthHandler) handle(ctx context.Context, request *generated.GetHealthRequest) (*generated.GetHealthResponse, error) {
	_ = ctx
	_ = request
	return &generated.GetHealthResponse{Status: generated.HealthStatus_HEALTH_STATUS_OK}, nil
}
