package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	"route256/loms/migrations"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	// Flags
	var action string
	var dsn string

	flag.StringVar(&action, "action", "up", "Migration action: up or down")
	flag.StringVar(&dsn, "dsn", "", "Database DSN")
	flag.Parse()

	// Check DSN
	if dsn == "" {
		log.Fatal("Database DSN must be provided")
	}

	// Database connect
	ctx := context.Background()
	db, err := NewConnect(ctx, dsn)
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
func NewConnect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
