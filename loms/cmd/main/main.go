package main

import (
	"flag"
	"fmt"
	loms "route256/loms/internal/app/loms"
	repo_order "route256/loms/internal/repository/orders"
	repo_stocks "route256/loms/internal/repository/stocks"
	loms_usecase "route256/loms/internal/service/loms"

	config "route256/loms/internal/config"
	"route256/loms/internal/server"

	"log"

	"github.com/joho/godotenv"
)

func main() {
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

	// Repository order
	repoOrder := repo_order.NewOrderRepository()

	// Repository stocks
	repoStocks := repo_stocks.NewStockRepository()
	repoStocks.LoadStocks(cfg.Data.GetStockFilePath())

	// Loms usecase
	lomsUsecaseService := loms_usecase.NewService(repoOrder, repoStocks)

	// Loms
	controller := loms.NewService(lomsUsecaseService)

	// GRPC server
	if err := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, controller).Start(); err != nil {
		log.Printf("Failed creating gRPC server, err:%s", err)
		return
	}
}
