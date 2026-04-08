package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	grpcApp "github.com/cathudson/order-service/internal/app"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/interceptors"
	report "github.com/cathudson/order-service/internal/reporter"
	"github.com/cathudson/order-service/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	l, _ := zap.NewDevelopment()
	log := l.Sugar()
	defer func() { _ = l.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	defer cancel()
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err) //nolint:gocritic // it is OK to forget log flush at this point
	}
	defer listener.Close()

	orderStore := store.NewOrderStore()

	go report.NewReporter(ctx, orderStore, log).Run()

	app := grpcApp.New(orderStore)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDInterceptor),
	)
	generated.RegisterOrderServiceServer(server, app)
	reflection.Register(server)
	err = server.Serve(listener)
	if err != nil {
		log.Infof("failed to serve: %v", err)
		return
	}
}
