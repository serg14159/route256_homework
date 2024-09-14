package service

import (
	"context"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the AddProduct method of CartService.
func TestCartService_AddProduct_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		SKU           models.SKU
		count         uint16
		getProduct    *models.GetProductResponse
		getProductErr error
		addItemErr    error
		wantErr       error
	}{
		{
			name:  "successful add",
			UID:   1,
			SKU:   100,
			count: 2,
			getProduct: &models.GetProductResponse{
				Name:  "Книга",
				Price: 400,
			},
			getProductErr: nil,
			addItemErr:    nil,
			wantErr:       nil,
		},
		{
			name:          "bad request with UID 0",
			UID:           0,
			SKU:           100,
			count:         1,
			getProduct:    nil,
			getProductErr: nil,
			addItemErr:    nil,
			wantErr:       internal_errors.ErrBadRequest,
		},
		{
			name:          "bad request with SKU 0",
			UID:           1,
			SKU:           0,
			count:         2,
			getProduct:    nil,
			getProductErr: nil,
			addItemErr:    nil,
			wantErr:       internal_errors.ErrBadRequest,
		},
		{
			name:          "bad request with count 0",
			UID:           1,
			SKU:           100,
			count:         0,
			getProduct:    nil,
			getProductErr: nil,
			addItemErr:    nil,
			wantErr:       internal_errors.ErrBadRequest,
		},
		{
			name:          "product service error",
			UID:           1,
			SKU:           100,
			count:         1,
			getProduct:    nil,
			getProductErr: internal_errors.ErrInternalServerError,
			addItemErr:    nil,
			wantErr:       internal_errors.ErrInternalServerError,
		},
		{
			name:          "product service SKU not found",
			UID:           1,
			SKU:           100,
			count:         1,
			getProduct:    nil,
			getProductErr: internal_errors.ErrPreconditionFailed,
			addItemErr:    nil,
			wantErr:       internal_errors.ErrPreconditionFailed,
		},
		{
			name:  "repository error when adding item",
			UID:   1,
			SKU:   100,
			count: 3,
			getProduct: &models.GetProductResponse{
				Name:  "Книга",
				Price: 400,
			},
			getProductErr: nil,
			addItemErr:    ErrRepository,
			wantErr:       ErrRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repoMock, productServiceMock, service := setup(t)

			ctx := context.Background()

			if tt.UID < 1 || tt.SKU < 1 || tt.count < 1 {
				err := service.AddProduct(ctx, tt.UID, tt.SKU, tt.count)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			productServiceMock.GetProductMock.Expect(ctx, tt.SKU).Return(tt.getProduct, tt.getProductErr)

			if tt.getProductErr == nil {
				repoMock.AddItemMock.Expect(ctx, tt.UID, models.CartItem{SKU: tt.SKU, Count: tt.count}).Return(tt.addItemErr)
			}

			err := service.AddProduct(ctx, tt.UID, tt.SKU, tt.count)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
