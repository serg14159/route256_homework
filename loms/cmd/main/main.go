package main

import (
	"flag"
	config "route256/loms/internal/config"
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

	// Cfg
	log.Printf("Cfg: %v", cfg)
}
