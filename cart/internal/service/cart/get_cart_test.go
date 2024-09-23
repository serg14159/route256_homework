package service

import (
	"context"
	"errors"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"route256/cart/internal/service/cart/mock"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the GetCart method of CartService.
func TestCartService_GetCart_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		setupMocks    func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock)
		expectedErr   error
		totalPrice    uint32
		errorContains string
	}{
		{
			name: "successful get cart with 1 item",
			UID:  1000000,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				items := []models.CartItem{{SKU: 700, Count: 3}}
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1000000)).Then(items, nil)
				productServiceMock.GetProductMock.When(ctx, models.SKU(700)).Then(&models.GetProductResponse{Name: "Product", Price: 100}, nil)
			},
			totalPrice:  300,
			expectedErr: nil,
		},
		{
			name: "successful get cart with 3 items",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				items := []models.CartItem{
					{SKU: 100, Count: 1},
					{SKU: 200, Count: 2},
					{SKU: 300, Count: 3},
				}
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1)).Then(items, nil)
				productServiceMock.GetProductMock.When(ctx, models.SKU(100)).Then(&models.GetProductResponse{Name: "Product 1", Price: 100}, nil)
				productServiceMock.GetProductMock.When(ctx, models.SKU(200)).Then(&models.GetProductResponse{Name: "Product 2", Price: 200}, nil)
				productServiceMock.GetProductMock.When(ctx, models.SKU(300)).Then(&models.GetProductResponse{Name: "Product 3", Price: 300}, nil)
			},
			totalPrice:  1400,
			expectedErr: nil,
		},
		{
			name: "bad request with UID 0",
			UID:  0,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
			},
			expectedErr: internal_errors.ErrBadRequest,
			totalPrice:  0,
		},
		{
			name: "repository error",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1)).Then(nil, ErrRepository)
			},
			expectedErr: ErrRepository,
			totalPrice:  0,
		},
		{
			name: "product service error",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				items := []models.CartItem{{SKU: 100, Count: 1}}
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1)).Then(items, nil)
				productServiceMock.GetProductMock.When(ctx, models.SKU(100)).Then(nil, internal_errors.ErrInternalServerError)
			},
			expectedErr: internal_errors.ErrInternalServerError,
			totalPrice:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repoMock, productServiceMock, _, service := setup(t)

			tt.setupMocks(ctx, repoMock, productServiceMock)

			items, totalPrice, err := service.GetCart(ctx, tt.UID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
				require.Nil(t, items)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.totalPrice, totalPrice)
				require.NotNil(t, items)
			}
		})
	}
}
