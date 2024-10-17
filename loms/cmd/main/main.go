package main

import (
	"log"

	"route256/loms/internal/app"
)

func main() {
	service, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	err = service.Run()
	if err != nil {
		log.Fatalf("Service error: %v", err)
	}
}
