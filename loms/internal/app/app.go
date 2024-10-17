package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	loms "route256/loms/internal/app/loms"
	"route256/loms/internal/config"
	db "route256/loms/internal/pkg/db"
	kafkaProducer "route256/loms/internal/pkg/kafka"
	repo_order "route256/loms/internal/repository/orders"
	repo_outbox "route256/loms/internal/repository/outbox"
	repo_stocks "route256/loms/internal/repository/stocks"
	"route256/loms/internal/server"
	loms_usecase "route256/loms/internal/service/loms"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const quitChannelBufferSize = 1
const errorChannelBufferSize = 1
const processOutboxInterval time.Duration = 2 * time.Second

type App struct {
	config     *config.Config
	ctx        context.Context
	cancel     context.CancelFunc
	pool       *pgxpool.Pool
	usecase    *loms_usecase.LomsService
	controller *loms.Service
	server     *server.GrpcServer
	producer   *kafkaProducer.KafkaProducer
}

// NewApp
func NewApp() (*App, error) {
	// Load environment
	_ = godotenv.Load()

	// Read flag
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = "config.yml"
	}

	// Read config
	cfg := config.NewConfig()
	if err := cfg.ReadConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to init configuration: %w", err)
	}

	log.Printf("Starting service: %s | version=%s | commitHash=%s | debug=%t | environment=%s",
		cfg.Project.GetName(), cfg.Project.GetVersion(), cfg.Project.GetCommitHash(), cfg.Project.GetDebug(), cfg.Project.GetEnvironment())

	// Cfg
	fmt.Printf("cfg: %v \n", cfg)

	// Context
	ctx, cancel := context.WithCancel(context.Background())

	// DB connect
	pool, err := db.NewConnect(ctx, &cfg.Database)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// TxManager
	txManager := db.NewTransactionManager(pool)

	// Repository order/stocks/outbox
	repoOrder := repo_order.NewOrderRepository(pool)
	repoStocks := repo_stocks.NewStockRepository(pool)
	repoOutbox := repo_outbox.NewOutboxRepository(pool)

	// Kafka producer
	kafkaProd, err := kafkaProducer.NewKafkaProducer(&cfg.Kafka)
	if err != nil {
		pool.Close()
		cancel()
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Loms usecase
	lomsUsecaseService := loms_usecase.NewService(repoOrder, repoStocks, repoOutbox, txManager, kafkaProd)

	// Loms controller
	controller := loms.NewService(lomsUsecaseService)

	// GRPC server
	grpcServer := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, controller)

	return &App{
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
		pool:       pool,
		usecase:    lomsUsecaseService,
		controller: controller,
		server:     grpcServer,
		producer:   kafkaProd,
	}, nil
}

// Run
func (a *App) Run() error {
	// Run outbox
	go a.startOutboxProcessor()

	// Create channel to listen interrupt or terminate signals
	quitChan := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// GRPC server
	serverErrChan := make(chan error, errorChannelBufferSize)
	go func() {
		if err := a.server.Start(); err != nil {
			log.Printf("Failed creating gRPC server, err:%s", err)
			serverErrChan <- err
		}
	}()

	// Wait
	select {
	case <-quitChan:
		log.Println("Shutdown signal received")
	case err := <-serverErrChan:
		if err != nil {
			log.Printf("Server error: %v", err)
			return err
		}
	}

	// Shutdown
	a.Shutdown()
	return nil
}

// startOutboxProcessor
func (a *App) startOutboxProcessor() {
	ticker := time.NewTicker(processOutboxInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.usecase.ProcessOutbox(a.ctx); err != nil {
				log.Printf("Error processing outbox: %v", err)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// Shutdown
func (a *App) Shutdown() {
	a.cancel()

	a.pool.Close()

	if err := a.producer.Close(); err != nil {
		log.Printf("Failed to close Kafka producer: %v", err)
	}

	log.Printf("Shutdown complete")
}
