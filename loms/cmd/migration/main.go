package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	config "route256/loms/internal/config"

	"route256/loms/migrations"

	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	// Load environment
	_ = godotenv.Load()

	// Flags
	var configPath string
	var action string

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&action, "action", "up", "Migration action: up or down")
	flag.Parse()

	// Set default config path
	if configPath == "" {
		configPath = "config.yml"
	}

	// Read config
	cfg := config.NewConfig()
	if err := cfg.ReadConfig(configPath); err != nil {
		log.Fatalf("Failed to initialize configuration, error: %s", err)
	}

	// Cfg
	fmt.Printf("cfg: %v", cfg)

	// Database connect
	ctx := context.Background()
	db, err := NewConnect(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to the database, error: %s", err)
	}
	defer db.Close()

	// Configure goose
	goose.SetBaseFS(migrations.EmbedFS)

	// Run migration
	switch action {
	case "up":
		err = goose.Up(db, ".")
		if err != nil {
			log.Fatalf("Error applying migrations up: %v", err)
		}
		log.Println("Migrations successfully applied up.")
	case "down":
		err = goose.Down(db, ".")
		if err != nil {
			log.Fatalf("Error rolling back migrations down: %v", err)
		}
		log.Println("Migrations successfully rolled back down.")
	default:
		log.Fatalf("Unknown migration action: %s. Use 'up' or 'down'.", action)
	}
}

// NewConnect return *sql.DB connection to database
func NewConnect(ctx context.Context, cfg *config.Database) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
