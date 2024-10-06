package repository

import (
	"context"
	"route256/cart/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRepository_DeleteItemsByUserID function for tests the DeleteItemsByUserID method of repository.
func TestRepository_DeleteItemsByUserID(t *testing.T) {
	// Init test data
	tests := []struct {
		name    string
		UID     models.UID
		setup   func(repo *Repository)
		wantErr bool
	}{
		{
			name: "successful delete user cart",
			UID:  1,
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
			setup:   func(repo *Repository) {},
			wantErr: true,
		},
		{
			name:    "delete empty cart",
			UID:     2,
			setup:   func(repo *Repository) {},
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
			err := repo.DeleteItemsByUserID(ctx, tt.UID)

			// Check want error
			if tt.wantErr {
				require.Error(t, err, "Error")
				return
			}
			require.NoError(t, err, "NoError")

			// Check storage
			_, exists := repo.storage[tt.UID]
			require.False(t, exists, "cart must be delete")
		})
	}
}

// TestRepository_DeleteItemsByUserID_Concurrent tests concurrent calls to DeleteItemsByUserID.
func TestRepository_DeleteItemsByUserID_Concurrent(t *testing.T) {
	// Run test parallel
	t.Parallel()

	repo := NewCartRepository()
	ctx := context.Background()

	const numGoroutines = 100
	const UID models.UID = 1
	item := models.CartItem{SKU: 1001, Count: 1}

	// Add item
	repo.mu.Lock()
	if repo.storage[UID] == nil {
		repo.storage[UID] = make(map[models.SKU]models.CartItem)
	}
	repo.storage[UID][item.SKU] = item
	repo.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrently delete user cart
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := repo.DeleteItemsByUserID(ctx, UID)
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	// Verify that the cart was deleted
	repo.mu.Lock()
	_, exists := repo.storage[UID]
	repo.mu.Unlock()
	require.False(t, exists, "the cart should be deleted")
}
