package main

import (
	"context"
	"os"
	"os/signal"
	"route256/cart/internal/app"

	"route256/utils/logger"
)

const quitChannelBufferSize = 1

func main() {
	// Create the application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the application
	application, err := app.NewApp(ctx)
	if err != nil {
		logger.Errorw(ctx, "Failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Run the application
	if err := application.Run(); err != nil {
		logger.Errorw(ctx, "Application run error", "error", err)
		os.Exit(1)
	}

	// Wait for OS interrupt signal
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Infow(ctx, "Shutdown Server ...")

	// Shutdown application
	if err := application.Shutdown(); err != nil {
		logger.Errorw(ctx, "Failed to shutdown application", "error", err)
	}

	logger.Infow(ctx, "Server exiting")
}
