package repository

import (
	"context"
	"route256/cart/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRepository_AddItem function for tests the AddItem method of repository.
func TestRepository_AddItem(t *testing.T) {
	// Init test data
	tests := []struct {
		name          string
		UID           models.UID
		item          models.CartItem
		setup         func(repo *Repository)
		expectedSKU   models.SKU
		expectedCount uint16
		wantErr       bool
	}{
		{
			name: "successful add item",
			UID:  1,
			item: models.CartItem{
				SKU:   1001,
				Count: 2,
			},
			setup:         func(repo *Repository) {},
			expectedSKU:   1001,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "successful add item again",
			UID:  1,
			item: models.CartItem{
				SKU:   1001,
				Count: 2,
			},
			setup: func(repo *Repository) {
				repo.storage[1] = map[models.SKU]models.CartItem{
					1001: {SKU: 1001, Count: 2},
				}
			},
			expectedSKU:   1001,
			expectedCount: 4,
			wantErr:       false,
		},
		{
			name: "invalid UID",
			UID:  0,
			item: models.CartItem{
				SKU:   1001,
				Count: 2,
			},
			setup:   func(repo *Repository) {},
			wantErr: true,
		},
		{
			name: "invalid SKU",
			UID:  1,
			item: models.CartItem{
				SKU:   0,
				Count: 2,
			},
			setup:   func(repo *Repository) {},
			wantErr: true,
		},
		{
			name: "invalid count",
			UID:  1,
			item: models.CartItem{
				SKU:   1001,
				Count: 0,
			},
			setup:   func(repo *Repository) {},
			wantErr: true,
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
			err := repo.AddItem(ctx, tt.UID, tt.item)

			// Check want error
			if tt.wantErr {
				require.Error(t, err, "Error")
				return
			}
			require.NoError(t, err, "NoError")

			// Check storage
			storedCart, ok := repo.storage[tt.UID]
			require.True(t, ok, "cart for this UID must exist")
			require.NotNil(t, storedCart, "cart must not be nil")

			storedItem, ok := storedCart[tt.item.SKU]
			require.True(t, ok, "item must be in cart")
			require.Equal(t, tt.expectedSKU, storedItem.SKU, "SKU must match")
			require.Equal(t, tt.expectedCount, storedItem.Count, "Count must match")
		})
	}
}

// TestRepository_AddItem_Concurrent tests concurrent calls to AddItem.
func TestRepository_AddItem_Concurrent(t *testing.T) {
	// Run test parallel
	t.Parallel()

	repo := NewCartRepository()
	ctx := context.Background()

	const numGoroutines = 100
	const UID models.UID = 1
	item := models.CartItem{SKU: 1001, Count: 1}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Add items
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := repo.AddItem(ctx, UID, item)
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	// Verify the final count
	repo.mu.Lock()
	storedCart, ok := repo.storage[UID]
	repo.mu.Unlock()
	require.True(t, ok)
	storedItem, ok := storedCart[item.SKU]
	require.True(t, ok)
	require.Equal(t, uint16(numGoroutines*int(item.Count)), storedItem.Count)
}
