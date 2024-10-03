package repository

import (
	"context"
	"route256/cart/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRepository_GetItemsByUserID function for tests the GetItemsByUserID method of repository.
func TestRepository_GetItemsByUserID(t *testing.T) {
	// Init test data
	tests := []struct {
		name          string
		UID           models.UID
		setup         func(repo *Repository)
		wantErr       bool
		expectedLen   int
		expectedItems []models.CartItem
	}{
		{
			name:        "get from empty cart",
			UID:         1,
			setup:       func(repo *Repository) {},
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name: "successful get user cart with 1 items",
			UID:  2,
			setup: func(repo *Repository) {
				repo.storage[2] = map[models.SKU]models.CartItem{
					1001: {SKU: 1001, Count: 2},
				}
			},
			wantErr:     false,
			expectedLen: 1,
			expectedItems: []models.CartItem{
				{SKU: 1001, Count: 2},
			},
		},
		{
			name: "successful get user cart with 3 items",
			UID:  3,
			setup: func(repo *Repository) {
				repo.storage[3] = map[models.SKU]models.CartItem{
					1003: {SKU: 1003, Count: 1},
					1002: {SKU: 1002, Count: 5},
					1001: {SKU: 1001, Count: 2},
				}
			},
			wantErr:     false,
			expectedLen: 3,
			expectedItems: []models.CartItem{
				{SKU: 1001, Count: 2},
				{SKU: 1002, Count: 5},
				{SKU: 1003, Count: 1},
			},
		},
		{
			name:          "invalid UID",
			UID:           0,
			setup:         func(repo *Repository) {},
			wantErr:       true,
			expectedLen:   0,
			expectedItems: []models.CartItem{},
		},
		{
			name: "cart for UID not found",
			UID:  4,
			setup: func(repo *Repository) {
				repo.storage[5] = map[models.SKU]models.CartItem{
					1001: {SKU: 1001, Count: 2},
				}
			},
			wantErr:       true,
			expectedLen:   0,
			expectedItems: []models.CartItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test parallel
			t.Parallel()

			// Init repo
			repo := NewCartRepository()

			// Setup storage
			tt.setup(repo)

			ctx := context.Background()

			// Run function
			items, err := repo.GetItemsByUserID(ctx, tt.UID)

			// Check want error
			if tt.wantErr {
				require.Error(t, err, "Error")
				return
			}
			require.NoError(t, err, "NoError")

			// Check items len
			require.Len(t, items, tt.expectedLen, "len must match")

			// Check items order
			for i, expectedItem := range tt.expectedItems {
				require.Equal(t, expectedItem, items[i], "items must match")
			}
		})
	}
}

// TestRepository_GetItemsByUserID_Concurrent tests concurrent calls to GetItemsByUserID.
func TestRepository_GetItemsByUserID_Concurrent(t *testing.T) {
	// Run test parallel
	t.Parallel()

	repo := NewCartRepository()
	ctx := context.Background()

	const numGoroutines = 100
	const UID models.UID = 1
	items := []models.CartItem{
		{SKU: 1001, Count: 2},
		{SKU: 1002, Count: 3},
	}

	// Add items
	for _, item := range items {
		err := repo.AddItem(ctx, UID, item)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrently get items from repository
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			retrievedItems, err := repo.GetItemsByUserID(ctx, UID)
			require.NoError(t, err)
			require.Len(t, retrievedItems, len(items))
		}()
	}

	wg.Wait()
}
