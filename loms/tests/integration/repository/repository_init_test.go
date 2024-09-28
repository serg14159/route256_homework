package repository_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"route256/loms/tests/integration/repository/migrations"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

var connTests *pgxpool.Pool

const DSN = "postgres://user:password@localhost:5434/postgres_test?sslmode=disable"

func TestMain(m *testing.M) {
	// DB connect for tests
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, DSN)
	if err != nil {
		panic(err)
	}
	connTests = pool
	defer connTests.Close()

	// DB connect for goose
	connGoose, err := sql.Open("pgx", DSN)
	if err != nil {
		panic(err)
	}
	defer connGoose.Close()

	// Run migrations
	runMigrations(connGoose, "up")

	code := m.Run()

	runMigrations(connGoose, "down")

	os.Exit(code)
}

// runMigrations
func runMigrations(db *sql.DB, action string) {
	// Configure goose
	goose.SetBaseFS(migrations.EmbedFS)

	// Init version
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %v", err)
	}
	if _, err := goose.GetDBVersion(db); err != nil {
		if err := goose.UpTo(db, ".", 0); err != nil {
			log.Fatalf("Error initializing goose_db_version: %v", err)
		}
	}
	// Run migration
	var err error
	switch action {
	case "up":
		err = goose.Up(db, ".")
		if err != nil {
			log.Fatalf("Error applying migrations up: %v", err)
		}
		log.Println("Migrations successfully applied up.")
	case "down":
		err = goose.DownTo(db, ".", 0)
		if err != nil {
			log.Fatalf("Error rolling back migrations down: %v", err)
		}
		log.Println("Migrations successfully rolled back down.")
	default:
		log.Fatalf("Unknown migration action: %s. Use 'up' or 'down'.", action)
	}
}
