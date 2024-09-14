package service

import (
	"context"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the GetCart method of CartService.
func TestCartService_GetCart_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		items         []models.CartItem
		getItemsErr   error
		getProduct    []models.GetProductResponse
		getProductErr error
		wantErr       error
		totalPrice    uint32
	}{
		{
			name:          "successful get cart with 1 item",
			UID:           1000000,
			items:         []models.CartItem{{SKU: 700, Count: 3}},
			getItemsErr:   nil,
			getProduct:    []models.GetProductResponse{{Name: "Product", Price: 100}},
			getProductErr: nil,
			totalPrice:    300,
		},
		{
			name: "successful get cart with 3 items",
			UID:  1,
			items: []models.CartItem{
				{SKU: 100, Count: 1},
				{SKU: 200, Count: 2},
				{SKU: 300, Count: 3},
			},
			getItemsErr: nil,
			getProduct: []models.GetProductResponse{
				{Name: "Product 1", Price: 100},
				{Name: "Product 2", Price: 200},
				{Name: "Product 3", Price: 300},
			},
			getProductErr: nil,
			totalPrice:    1400,
		},
		{
			name:          "bad request with UID 0",
			UID:           0,
			items:         nil,
			getItemsErr:   nil,
			getProduct:    nil,
			getProductErr: nil,
			wantErr:       internal_errors.ErrBadRequest,
			totalPrice:    0,
		},
		{
			name:          "repository error",
			UID:           1,
			items:         nil,
			getItemsErr:   ErrRepository,
			getProduct:    nil,
			getProductErr: nil,
			wantErr:       ErrRepository,
			totalPrice:    0,
		},
		{
			name:          "product service error",
			UID:           1,
			items:         []models.CartItem{{SKU: 100, Count: 1}},
			getItemsErr:   nil,
			getProduct:    nil,
			getProductErr: internal_errors.ErrInternalServerError,
			wantErr:       internal_errors.ErrInternalServerError,
			totalPrice:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repoMock, productServiceMock, service := setup(t)

			ctx := context.Background()

			if tt.UID < 1 {
				_, totalPrice, err := service.GetCart(ctx, tt.UID)
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.totalPrice, totalPrice)
				return
			}

			repoMock.GetItemsByUserIDMock.When(ctx, tt.UID).Then(tt.items, tt.getItemsErr)

			if tt.getItemsErr == nil {
				for i, item := range tt.items {
					itemSKU := item.SKU

					if tt.getProduct == nil {
						productServiceMock.GetProductMock.When(ctx, itemSKU).Then(nil, tt.getProductErr)
					} else {
						product := tt.getProduct[i]
						productServiceMock.GetProductMock.When(ctx, itemSKU).Then(&models.GetProductResponse{
							Name:  product.Name,
							Price: product.Price,
						}, tt.getProductErr)
					}

				}
			}

			_, totalPrice, err := service.GetCart(ctx, tt.UID)

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.totalPrice, totalPrice)

		})
	}
}
