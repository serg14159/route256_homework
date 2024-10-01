package service_test

import (
	"context"
	"errors"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"route256/cart/internal/service/cart/mock"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCartService_GetCart_Table function for tests the GetCart method of CartService.
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
			name: "successful retrieval of cart with 1 item",
			UID:  1000000,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				items := []models.CartItem{{SKU: 700, Count: 3}}
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1000000)).Then(items, nil)

				var mu sync.Mutex
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					mu.Lock()
					defer mu.Unlock()
					if sku == 700 {
						return &models.GetProductResponse{Name: "Product", Price: 100}, nil
					}
					return nil, errors.New("product not found")
				})
			},
			totalPrice:  300,
			expectedErr: nil,
		},
		{
			name: "successful retrieval of cart with 3 items",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock) {
				items := []models.CartItem{
					{SKU: 100, Count: 1},
					{SKU: 200, Count: 2},
					{SKU: 300, Count: 3},
				}
				repoMock.GetItemsByUserIDMock.When(ctx, models.UID(1)).Then(items, nil)

				var mu sync.Mutex
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					mu.Lock()
					defer mu.Unlock()
					switch sku {
					case 100:
						return &models.GetProductResponse{Name: "Product 1", Price: 100}, nil
					case 200:
						return &models.GetProductResponse{Name: "Product 2", Price: 200}, nil
					case 300:
						return &models.GetProductResponse{Name: "Product 3", Price: 300}, nil
					default:
						return nil, errors.New("product not found")
					}
				})
			},
			totalPrice:  1400,
			expectedErr: nil,
		},
		{
			name: "invalid request with UID 0",
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

				var mu sync.Mutex
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					mu.Lock()
					defer mu.Unlock()
					return nil, internal_errors.ErrInternalServerError
				})
			},
			expectedErr: internal_errors.ErrInternalServerError,
			totalPrice:  0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repoMock, productServiceMock, _, service := setup(t)

			tt.setupMocks(ctx, repoMock, productServiceMock)

			items, totalPrice, err := service.GetCart(ctx, tt.UID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"expected error %v or message to contain: %s", tt.expectedErr, tt.errorContains)
				require.Nil(t, items)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.totalPrice, totalPrice)
				require.NotNil(t, items)
			}
		})
	}
}
