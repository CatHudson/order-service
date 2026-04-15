package app

import (
	"context"

	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/service"
	"github.com/cathudson/order-service/internal/store"
)

type App struct {
	createOrderHandler *createOrderHandler
	getOrderHandler    *getOrderHandler

	generated.UnimplementedOrderServiceServer
}

func New(orderService *service.OrderService, orderStore store.OrderStore) *App {
	return &App{
		createOrderHandler: newCreateOrderHandler(orderService),
		getOrderHandler:    newGetOrderHandler(orderStore),
	}
}

func (a *App) CreateOrder(ctx context.Context, request *generated.CreateOrderRequest) (*generated.CreateOrderResponse, error) {
	return a.createOrderHandler.handle(ctx, request)
}

func (a *App) GetOrder(ctx context.Context, request *generated.GetOrderRequest) (*generated.GetOrderResponse, error) {
	return a.getOrderHandler.handle(ctx, request)
}
