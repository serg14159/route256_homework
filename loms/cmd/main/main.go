package main

import (
	"log"
	"os"

	"route256/loms/internal/app"
)

func main() {
	// Initialize the application
	service, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
		os.Exit(1)
	}

	// Run application
	err = service.Run()
	if err != nil {
		log.Fatalf("Service error: %v", err)
		os.Exit(1)
	}
}
