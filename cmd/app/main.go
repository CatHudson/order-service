package main

import (
	"context"
	"net"

	grpcApp "github.com/cathudson/order-service/internal/app"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", ":8081")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	orderStore := store.NewOrderStore()
	app := grpcApp.New(orderStore)
	server := grpc.NewServer()
	generated.RegisterOrderServiceServer(server, app)
	reflection.Register(server)
	err = server.Serve(listener)
	if err != nil {
		panic(err)
	}
}
