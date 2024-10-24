package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"route256/cart/internal/app"

	loggerPkg "route256/cart/internal/pkg/logger"
)

const quitChannelBufferSize = 1

func main() {
	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the application
	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application run error: %v", err)
	}

	// Wait for OS interrupt signal
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt)
	<-quit
	loggerPkg.Infow(ctx, "Shutdown Server ...")

	// Shutdown app
	if err := application.Shutdown(); err != nil {
		loggerPkg.Errorw(ctx, "Failed to shutdown application", "error", err)
	}

	loggerPkg.Infow(ctx, "Server exiting")
}
