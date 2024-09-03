package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"route256/cart/internal/app/server"
	"route256/cart/internal/config"
	"route256/cart/internal/pkg/cart/repository"
	"route256/cart/internal/pkg/cart/service"
	"route256/cart/internal/pkg/clients/product_service"
	"time"

	"log"

	"github.com/joho/godotenv"
)

const quitChannelBufferSize = 1
const shutdownTimeout = 5 * time.Second

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

	// Repository
	cartRepository := repository.NewCartRepository()

	// Product service client
	productService := product_service.NewClient(&cfg.ProductService)

	// Service
	cartService := service.NewService(cartRepository, productService)

	// Server
	s := server.NewServer(&cfg.Server, cartService)

	err := s.Run()
	if err != nil {
		log.Printf("Failed to start server, err:%s", err)
	}

	// Wait os interrupt
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Printf("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Failed server shutdown, err:%s", err)
	}
	log.Printf("Server exiting")
}
