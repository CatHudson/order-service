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
	"github.com/cathudson/order-service/internal/health"
	"github.com/cathudson/order-service/internal/interceptors"
	report "github.com/cathudson/order-service/internal/reporter"
	"github.com/cathudson/order-service/internal/service"
	"github.com/cathudson/order-service/internal/store"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	gh "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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
	var primary, replica *sqlx.DB
	primary, err = db.NewPostgresDB(ctx, cfg.Postgres.Primary)
	if err != nil {
		return fmt.Errorf("failed to connect to Primary: %w", err)
	}
	defer primary.Close()
	replica, err = db.NewPostgresDB(ctx, cfg.Postgres.Replica)
	if err != nil {
		logger.Warnf("failed to connect to Replica: %v", err)
	} else {
		defer replica.Close()
	}
	connContainer := store.NewConnContainer(primary, replica)

	tx := store.NewTransactor(primary)

	orderStore := store.NewOrderStore(connContainer)
	ordersAuditLogStore := store.NewOrdersAuditLogStore(connContainer)

	orderService := service.NewOrderService(tx, orderStore, ordersAuditLogStore)

	// goroutine reporter
	go report.NewReporter(orderStore, logger).Run(ctx)

	// app
	app := grpcApp.New(orderService, orderStore)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDInterceptor),
	)

	// health check
	healthServer := gh.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	go health.NewChecker(healthServer, primary, replica, connContainer.SetReplicaHealthy).Run(ctx)

	generated.RegisterOrderServiceServer(server, app)
	reflection.Register(server)
	go func() {
		<-ctx.Done()
		server.GracefulStop()
		healthServer.Shutdown()
	}()
	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
