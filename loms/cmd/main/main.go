package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	loms "route256/loms/internal/app/loms"
	kafkaProducer "route256/loms/internal/pkg/kafka"
	repo_order "route256/loms/internal/repository/orders"
	repo_stocks "route256/loms/internal/repository/stocks"
	loms_usecase "route256/loms/internal/service/loms"
	"syscall"

	config "route256/loms/internal/config"
	"route256/loms/internal/server"

	"log"

	db "route256/loms/internal/pkg/db"

	"github.com/joho/godotenv"
)

const quitChannelBufferSize = 1
const errorChannelBufferSize = 1

func main() {
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
		log.Printf("Failed init configuration, err:%s", err)
	}

	log.Printf("Starting service: %s | version=%s | commitHash=%s | debug=%t | environment=%s",
		cfg.Project.GetName(), cfg.Project.GetVersion(), cfg.Project.GetCommitHash(), cfg.Project.GetDebug(), cfg.Project.GetEnvironment())

	// Cfg
	fmt.Printf("cfg: %v \n", cfg)

	// DB connect
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewConnect(ctx, &cfg.Database)
	if err != nil {
		log.Printf("Failed connect to database, err:%s", err)
		os.Exit(1)
	}
	defer pool.Close()

	// TxManager
	txManager := db.NewTransactionManager(pool)

	// Repository order
	repoOrder := repo_order.NewOrderRepository(pool)

	// Repository stocks
	repoStocks := repo_stocks.NewStockRepository(pool)

	// Kafka producer
	kafkaProd, err := kafkaProducer.NewKafkaProducer(&cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer func() {
		if err := kafkaProd.Close(); err != nil {
			log.Printf("Failed to close Kafka producer: %v", err)
		}
	}()

	// Loms usecase
	lomsUsecaseService := loms_usecase.NewService(repoOrder, repoStocks, txManager, kafkaProd)

	// Loms
	controller := loms.NewService(lomsUsecaseService)

	// Create channel to listen interrupt or terminate signals
	quitChan := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// GRPC server
	serverErrChan := make(chan error, errorChannelBufferSize)
	go func() {
		if err := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, controller).Start(); err != nil {
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
		}
	}

	cancel()

	log.Printf("Shutdown ...")
}
