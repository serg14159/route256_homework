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
	"time"

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

	// Repository
	cartRepository := repository.NewCartRepository()

	// Service
	cartService := service.NewService(cartRepository)

	// Server
	s := server.NewServer(&cfg.Server, cartService)

	err := s.Run()
	if err != nil {
		log.Printf("Failed to start server, err:%s", err)
	}

	// Wait os interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Printf("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Failed server shutdown, err:%s", err)
	}
	log.Printf("Server exiting")
}
