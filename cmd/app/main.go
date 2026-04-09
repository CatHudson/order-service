package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	grpcApp "github.com/cathudson/order-service/internal/app"
	"github.com/cathudson/order-service/internal/config"
	"github.com/cathudson/order-service/internal/db"
	"github.com/cathudson/order-service/internal/generated"
	"github.com/cathudson/order-service/internal/interceptors"
	report "github.com/cathudson/order-service/internal/reporter"
	"github.com/cathudson/order-service/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error on startup: %v", err)
	}
}

func run() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() {
		_ = l.Sync()
	}()

	logger := l.Sugar()

	cfg, err := config.Load("/config/config.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	defer cancel()

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, cfg.GRPC.Network, fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	dbConn, err := db.NewPostgresDB(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()

	orderStore := store.NewOrderStore(dbConn)

	go report.NewReporter(orderStore, logger).Run(ctx)

	app := grpcApp.New(orderStore)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDInterceptor),
	)
	generated.RegisterOrderServiceServer(server, app)
	reflection.Register(server)
	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
