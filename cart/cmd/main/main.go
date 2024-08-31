package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"route256/cart/internal/app/server"
	"route256/cart/internal/config"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("Failed init configuration")
	}

	log.Info().
		Str("version", cfg.Project.GetVersion()).
		Str("commitHash", cfg.Project.GetCommitHash()).
		Bool("debug", cfg.Project.GetDebug()).
		Str("environment", cfg.Project.GetEnvironment()).
		Msgf("Starting service: %s", cfg.Project.GetName())

	// Set log level
	if cfg.Project.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Msgf("Cfg: %v", cfg)

	// Server
	var a any
	s := server.NewServer(&cfg.Server, a)

	err := s.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Wait os interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Debug().Msgf("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed server shutdown")
	}
	log.Debug().Msgf("Server exiting")
}
