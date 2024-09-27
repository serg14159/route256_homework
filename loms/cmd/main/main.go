package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	loms "route256/loms/internal/app/loms"
	repo_order "route256/loms/internal/repository/orders"
	repo_stocks "route256/loms/internal/repository/stocks"
	loms_usecase "route256/loms/internal/service/loms"

	config "route256/loms/internal/config"
	"route256/loms/internal/server"

	"log"

	db "route256/loms/internal/pkg/db"

	"github.com/joho/godotenv"
)

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
	fmt.Printf("cfg: %v", cfg)

	// DB connect
	ctx := context.Background()
	conn, err := db.NewConnect(ctx, &cfg.Database)
	if err != nil {
		log.Printf("Failed connect to database, err:%s", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// TxManager
	txManager := db.NewTransactionManager(conn)

	// Repository order
	repoOrder := repo_order.NewOrderRepository(conn)

	// Repository stocks
	repoStocks := repo_stocks.NewStockRepository(conn)

	// Loms usecase
	lomsUsecaseService := loms_usecase.NewService(repoOrder, repoStocks, txManager)

	// Loms
	controller := loms.NewService(lomsUsecaseService)

	// GRPC server
	if err := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, controller).Start(); err != nil {
		log.Printf("Failed creating gRPC server, err:%s", err)
		return
	}
}
