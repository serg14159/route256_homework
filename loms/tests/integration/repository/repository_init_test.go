package repository_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"route256/loms/tests/integration/repository/migrations"
	"testing"

	"route256/loms/internal/pkg/shard_manager"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

var connTests *pgxpool.Pool

const (
	DSN              = "postgres://user:password@localhost:5432/postgres_test?sslmode=disable"
	ShardDSN1        = "postgres://user:password@localhost:5430/postgres_test?sslmode=disable"
	ShardDSN2        = "postgres://user:password@localhost:5431/postgres_test?sslmode=disable"
	ShardCount       = 2
	ShardBucketCount = 1000
)

var shardManager *shard_manager.ShardManager

// TestMain
func TestMain(m *testing.M) {
	// Context
	ctx := context.Background()

	// DB connection
	pool, err := pgxpool.New(ctx, DSN)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	connTests = pool
	defer connTests.Close()

	// Initialize shard pools
	shards := []string{ShardDSN1, ShardDSN2}
	shardPools := make([]*pgxpool.Pool, len(shards))
	for i, dsn := range shards {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			log.Fatalf("Failed to connect to shard %d: %v", i, err)
		}
		shardPools[i] = pool
		defer shardPools[i].Close()
	}

	// Initialize ShardManager
	shardFn := shard_manager.GetMurmur3ShardFn(len(shardPools))
	shardManager = shard_manager.NewShardManager(shardFn, shardPools, ShardBucketCount)

	// Initialize goose pools
	dsns := []string{DSN, ShardDSN1, ShardDSN2}
	gooseConns := make([]*sql.DB, len(dsns))
	for i, dsn := range dsns {
		conn, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to shard %d: %v", i, err)
		}
		gooseConns[i] = conn
		defer gooseConns[i].Close()
	}

	// Run migrations
	for i, conn := range gooseConns {
		if err := runMigrations(conn, "up"); err != nil {
			log.Fatalf("Failed to run migrations on shard %d: %v", i, err)
		}
		defer func(c *sql.DB, i int) {
			if err := runMigrations(c, "down"); err != nil {
				log.Printf("Failed to rollback migrations on shard %d: %v", i, err)
			}
		}(conn, i)
	}

	code := m.Run()

	os.Exit(code)
}

// runMigrations
func runMigrations(db *sql.DB, action string) error {
	// Configure goose
	goose.SetBaseFS(migrations.EmbedFS)

	// Init version
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %v", err)
		return err
	}

	// Run migration
	var err error
	switch action {
	case "up":
		err = goose.Up(db, ".")
		if err != nil {
			log.Fatalf("Error applying migrations up: %v", err)
			return err
		}
		log.Println("Migrations successfully applied up.")
	case "down":
		err = goose.DownTo(db, ".", 0)
		if err != nil {
			log.Fatalf("Error rolling back migrations down: %v", err)
			return err
		}
		log.Println("Migrations successfully rolled back down.")
	default:
		log.Fatalf("Unknown migration action: %s. Use 'up' or 'down'.", action)
		return err
	}
	return nil
}
