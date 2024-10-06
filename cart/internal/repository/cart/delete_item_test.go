package repository

import (
	"context"
	"route256/cart/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRepository_DeleteItem function for tests the DeleteItem method of repository.
func TestRepository_DeleteItem(t *testing.T) {
	// Init test data
	tests := []struct {
		name    string
		UID     models.UID
		SKU     models.SKU
		setup   func(repo *Repository)
		wantErr bool
	}{
		{
			name: "successful delete item",
			UID:  1,
			SKU:  1001,
			setup: func(repo *Repository) {
				repo.storage[1] = map[models.SKU]models.CartItem{
					1001: {SKU: 1001, Count: 2},
				}
			},
			wantErr: false,
		},
		{
			name:    "invalid UID",
			UID:     0,
			SKU:     1001,
			setup:   func(repo *Repository) {},
			wantErr: true,
		},
		{
			name:    "invalid SKU",
			UID:     1,
			SKU:     0,
			setup:   func(repo *Repository) {},
			wantErr: true,
		},
		{
			name: "item does not exist",
			UID:  1,
			SKU:  9999,
			setup: func(repo *Repository) {
				repo.storage[1] = map[models.SKU]models.CartItem{
					1001: {SKU: 1001, Count: 2},
				}
			},
			wantErr: false,
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
			err := repo.DeleteItem(ctx, tt.UID, tt.SKU)

			// Check want error
			if tt.wantErr {
				require.Error(t, err, "Error")
				return
			}
			require.NoError(t, err, "NoError")

			// Check storage
			if storedCart, exists := repo.storage[tt.UID]; exists {
				_, itemExists := storedCart[tt.SKU]
				require.False(t, itemExists, "item must be delete")
			}
		})
	}
}

// TestRepository_DeleteItem_Concurrent tests concurrent calls to DeleteItem.
func TestRepository_DeleteItem_Concurrent(t *testing.T) {
	// Run test parallel
	t.Parallel()

	repo := NewCartRepository()
	ctx := context.Background()

	const numGoroutines = 100
	const UID models.UID = 1
	SKU := models.SKU(1001)
	item := models.CartItem{SKU: SKU, Count: 1}

	// Add item
	repo.mu.Lock()
	if repo.storage[UID] == nil {
		repo.storage[UID] = make(map[models.SKU]models.CartItem)
	}
	repo.storage[UID][SKU] = item
	repo.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrently delete items from repository
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := repo.DeleteItem(ctx, UID, SKU)
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	// Verify the item has been deleted
	repo.mu.Lock()
	_, exists := repo.storage[UID][SKU]
	repo.mu.Unlock()
	require.False(t, exists, "item should be deleted")
}
