package repository_test

import (
	"context"
	"route256/loms/internal/models"
	repository "route256/loms/internal/repository/stocks"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test for GetAvailableStockBySKU
func TestGetAvailableStockBySKU(t *testing.T) {
	stockRepo := repository.NewStockRepository(connTests)

	ctx := context.Background()
	availableStock, err := stockRepo.GetAvailableStockBySKU(ctx, models.SKU(1))
	require.NoError(t, err)
	require.Equal(t, uint64(90), availableStock)
}

// Test for ReserveItems
func TestReserveItems(t *testing.T) {
	stockRepo := repository.NewStockRepository(connTests)

	items := []models.Item{
		{SKU: 1, Count: 5},
		{SKU: 2, Count: 10},
	}

	ctx := context.Background()

	// Run tx
	tx, err := connTests.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// reserve
	err = stockRepo.ReserveItems(ctx, tx, items)
	require.NoError(t, err)

	// Commit
	err = tx.Commit(ctx)
	require.NoError(t, err)
}

// Test for RemoveReservedItems
func TestRemoveReservedItems(t *testing.T) {
	stockRepo := repository.NewStockRepository(connTests)

	items := []models.Item{
		{SKU: 1, Count: 5},
		{SKU: 2, Count: 10},
	}

	ctx := context.Background()

	// Run tx
	tx, err := connTests.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Remove reserved
	err = stockRepo.RemoveReservedItems(ctx, tx, items)
	require.NoError(t, err)

	// Commit
	err = tx.Commit(ctx)
	require.NoError(t, err)
}

// Test for CancelReservedItems
func TestCancelReservedItems(t *testing.T) {
	stockRepo := repository.NewStockRepository(connTests)

	items := []models.Item{
		{SKU: 1, Count: 5},
		{SKU: 2, Count: 10},
	}
	ctx := context.Background()
	// Run tx
	tx, err := connTests.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Cancel reserved
	err = stockRepo.CancelReservedItems(ctx, tx, items)
	require.NoError(t, err)

	// Commit
	err = tx.Commit(ctx)
	require.NoError(t, err)
}
