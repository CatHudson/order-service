package app

import (
	"context"

	"github.com/cathudson/order-service/internal/producer"
	"github.com/cathudson/order-service/internal/proto"
	"github.com/cathudson/order-service/internal/store"
	"go.uber.org/zap"
)

type App struct {
	createOrderHandler *createOrderHandler
	getOrderHandler    *getOrderHandler

	proto.UnimplementedOrderServiceServer
}

func New(createOrderProducer producer.CreateOrderProducer, orderStore store.OrderStore, orderResultStore store.OrderResultStore, logger *zap.SugaredLogger) *App {
	return &App{
		createOrderHandler: newCreateOrderHandler(createOrderProducer, orderResultStore, orderStore, logger),
		getOrderHandler:    newGetOrderHandler(orderStore),
	}
}

func (a *App) CreateOrder(ctx context.Context, request *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	return a.createOrderHandler.handle(ctx, request)
}

func (a *App) GetOrder(ctx context.Context, request *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	return a.getOrderHandler.handle(ctx, request)
}
