package app

import (
	"context"

	"github.com/cathudson/order-service/internal/generated"
)

type App struct {
	getHealthHandler *getHealthHandler
	// other handlers

	generated.UnimplementedOrderServiceServer
}

func New() *App {
	return &App{
		getHealthHandler: newGetHealthHandler(),
	}
}

func (a *App) GetHealth(ctx context.Context, request *generated.GetHealthRequest) (*generated.GetHealthResponse, error) {
	return a.getHealthHandler.handle(ctx, request)
}
