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
	"github.com/cathudson/order-service/internal/consumer"
	"github.com/cathudson/order-service/internal/db"
	"github.com/cathudson/order-service/internal/health"
	"github.com/cathudson/order-service/internal/interceptor"
	"github.com/cathudson/order-service/internal/producer"
	"github.com/cathudson/order-service/internal/proto"
	report "github.com/cathudson/order-service/internal/reporter"
	"github.com/cathudson/order-service/internal/service"
	"github.com/cathudson/order-service/internal/store"
	"github.com/cathudson/order-service/internal/task"
	"github.com/cathudson/order-service/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
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

//nolint:funlen
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
		logger.Warnw("failed to connect to Replica:", "error", err)
	} else {
		defer replica.Close()
	}
	connContainer := store.NewConnContainer(primary, replica)

	tx := store.NewTransactor(primary)

	orderStore := store.NewOrderStore(connContainer)
	ordersAuditLogStore := store.NewOrdersAuditLogStore(connContainer)

	// services
	orderService := service.NewOrderService(tx, orderStore, ordersAuditLogStore)

	// asynq
	mux := asynq.NewServeMux()
	createOrderProcessor := worker.NewCreateOrderProcessor(orderService)
	createOrderProcessor.Register(mux)

	asynqClient := task.NewAsynqClient(asynq.RedisClientOpt{Addr: cfg.Redis.Addr})
	defer asynqClient.Close()

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: cfg.Asynq.Concurrency,
		},
	)
	err = asynqServer.Start(mux)
	if err != nil {
		return fmt.Errorf("failed to start asynq server: %w", err)
	}
	defer asynqServer.Stop()

	// kafka
	createOrderProducer := producer.NewCreateOrderProducer(cfg.Kafka)
	orderResultProducer := producer.NewOrderResultProducer(cfg.Kafka)
	createOrderConsumer := consumer.NewCreateOrderConsumer(asynqClient, logger)
	createOrderReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.Address},
		Topic:   cfg.Kafka.Consumers.CreateOrderTopic,
		GroupID: cfg.Kafka.Consumers.GroupID,
	})
	runner := consumer.NewRunner(createOrderReader, createOrderConsumer, func() *proto.CreateOrderEvent { return &proto.CreateOrderEvent{} }, logger)
	go runner.Run(ctx)

	// goroutine reporter
	go report.NewReporter(orderStore, logger).Run(ctx)

	// app
	app := grpcApp.New(createOrderProducer, orderStore)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.RequestIDInterceptor),
	)

	// health check
	healthServer := gh.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	go health.NewWatcher(healthServer, primary, replica, connContainer.SetReplicaHealthy).Run(ctx)

	proto.RegisterOrderServiceServer(server, app)
	reflection.Register(server)
	go func() {
		<-ctx.Done()
		server.GracefulStop()
		healthServer.Shutdown()
		asynqServer.Shutdown()
		if err = createOrderReader.Close(); err != nil {
			logger.Errorf("failed to close consumer: %v", err)
		}
		if err = createOrderProducer.Close(); err != nil {
			logger.Errorf("failed to close producer: %v", err)
		}
		if err = orderResultProducer.Close(); err != nil {
			logger.Errorf("failed to close producer: %v", err)
		}
	}()
	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
