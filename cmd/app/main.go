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
	"github.com/cathudson/order-service/internal/service"
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
	// logger
	l, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() {
		_ = l.Sync()
	}()

	logger := l.Sugar()

	// config
	cfg, err := config.Load("config/config.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	defer cancel()

	// listener
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, cfg.GRPC.Network, fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	// DB stuff
	dbConn, err := db.NewPostgresDB(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()
	tx := store.NewTransactor(dbConn)

	orderStore := store.NewOrderStore(dbConn)
	ordersAuditLogStore := store.NewOrdersAuditLogStore(dbConn)

	orderService := service.NewOrderService(tx, orderStore, ordersAuditLogStore)

	// goroutine reporter
	go report.NewReporter(orderStore, logger).Run(ctx)

	// app
	app := grpcApp.New(orderService, orderStore)
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
