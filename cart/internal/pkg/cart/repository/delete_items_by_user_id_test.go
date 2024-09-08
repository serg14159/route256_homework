package repository

import (
	"context"
	"route256/cart/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepository_DeleteItemsByUserID(t *testing.T) {
	// Init repo
	repo := NewCartRepository()

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
			ctx := context.Background()

			// Setup storage
			tt.setup(repo)

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
