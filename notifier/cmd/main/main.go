package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"route256/notifier/internal/app/handler"
	"route256/notifier/internal/config"
	"route256/notifier/internal/consumer"
	service "route256/notifier/internal/service/notifier"

	"log"

	"github.com/joho/godotenv"
)

const quitChannelBufferSize = 1

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
	log.Printf("Cfg: %v", cfg)

	// Service
	notifierService := service.NewNotifierService()

	// Handle
	messageHandler := handler.NewMessageHandler(notifierService)

	// Consumer
	kafkaConsumer, err := consumer.NewKafkaConsumer(&cfg.Kafka, messageHandler)
	if err != nil {
		log.Printf("Failed creating consumer, err: %s", err)
		os.Exit(1)
	}

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run consumer
	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("Failed start consumer, err: %s", err)
			os.Exit(1)
		}
	}()

	// Wait os interrupt
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Printf("Shutdown ...")
}
