package main

import (
	"context"
	"os"

	"route256/loms/internal/app"
	"route256/utils/logger"
)

func main() {
	// Create the application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the application
	application, err := app.NewApp(ctx, cancel)
	if err != nil {
		logger.Errorw(ctx, "Failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Run application
	err = application.Run()
	if err != nil {
		logger.Errorw(ctx, "Application run error", "error", err)
		os.Exit(1)
	}
}
