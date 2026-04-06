package app

import (
	"context"

	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/store"
)

type App struct {
	getHealthHandler   *getHealthHandler
	createOrderHandler *createOrderHandler
	getOrderHandler    *getOrderHandler

	generated.UnimplementedOrderServiceServer
}

func New(orderStore store.OrderStore) *App {
	return &App{
		getHealthHandler:   newGetHealthHandler(),
		createOrderHandler: newCreateOrderHandler(orderStore),
		getOrderHandler:    newGetOrderHandler(orderStore),
	}
}

func (a *App) GetHealth(ctx context.Context, request *generated.GetHealthRequest) (*generated.GetHealthResponse, error) {
	return a.getHealthHandler.handle(ctx, request)
}

func (a *App) CreateOrder(ctx context.Context, request *generated.CreateOrderRequest) (*generated.CreateOrderResponse, error) {
	return a.createOrderHandler.handle(ctx, request)
}

func (a *App) GetOrder(ctx context.Context, request *generated.GetOrderRequest) (*generated.GetOrderResponse, error) {
	return a.getOrderHandler.handle(ctx, request)
}
