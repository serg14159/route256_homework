package service

import (
	"context"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the DelProduct method of CartService.
func TestCartService_DelProduct_Table(t *testing.T) {
	tests := []struct {
		name      string
		UID       models.UID
		SKU       models.SKU
		deleteErr error
		wantErr   error
	}{
		{
			name:    "successful delete",
			UID:     1,
			SKU:     100,
			wantErr: nil,
		},
		{
			name:    "bad request with UID 0",
			UID:     0,
			SKU:     100,
			wantErr: internal_errors.ErrBadRequest,
		},
		{
			name:    "bad request with SKU 0",
			UID:     1,
			SKU:     0,
			wantErr: internal_errors.ErrBadRequest,
		},
		{
			name:    "bad request with UID 0 and SKU 0",
			UID:     0,
			SKU:     0,
			wantErr: internal_errors.ErrBadRequest,
		},
		{
			name:      "repository error",
			UID:       1,
			SKU:       100,
			deleteErr: ErrRepository,
			wantErr:   ErrRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock, _, service := setup(t)

			ctx := context.Background()

			if tt.UID < 1 || tt.SKU < 1 {
				err := service.DelProduct(ctx, tt.UID, tt.SKU)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			repoMock.DeleteItemMock.Expect(ctx, tt.UID, tt.SKU).Return(tt.deleteErr)

			err := service.DelProduct(ctx, tt.UID, tt.SKU)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
