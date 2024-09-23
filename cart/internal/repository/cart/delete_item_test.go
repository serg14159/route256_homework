package repository

import (
	"context"
	"route256/cart/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the DeleteItem method of repository.
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
