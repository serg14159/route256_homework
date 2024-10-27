package service_test

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

// TestCartService_AddProduct_Table function for tests the AddProduct method of CartService.
func TestCartService_AddProduct_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		SKU           models.SKU
		count         uint16
		setupMocks    func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock)
		expectedErr   error
		errorContains string
	}{
		{
			name:  "successful add",
			UID:   1,
			SKU:   100,
			count: 2,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return &models.GetProductResponse{Name: "Книга", Price: 400}, nil
				})
				lomsServiceMock.StocksInfoMock.Set(func(ctx context.Context, sku models.SKU) (int64, error) {
					require.Equal(t, models.SKU(100), sku)
					return int64(5), nil
				})
				repoMock.AddItemMock.Set(func(ctx context.Context, uid models.UID, item models.CartItem) error {
					require.Equal(t, models.UID(1), uid)
					require.Equal(t, models.CartItem{SKU: 100, Count: 2}, item)
					return nil
				})
			},
			expectedErr: nil,
		},
		{
			name:  "bad request with UID 0",
			UID:   0,
			SKU:   100,
			count: 1,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
			},
			expectedErr: internal_errors.ErrBadRequest,
		},
		{
			name:  "bad request with SKU 0",
			UID:   1,
			SKU:   0,
			count: 2,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
			},
			expectedErr: internal_errors.ErrBadRequest,
		},
		{
			name:  "bad request with count 0",
			UID:   1,
			SKU:   100,
			count: 0,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
			},
			expectedErr: internal_errors.ErrBadRequest,
		},
		{
			name:  "product service error",
			UID:   1,
			SKU:   100,
			count: 1,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return nil, internal_errors.ErrInternalServerError
				})
			},
			expectedErr: internal_errors.ErrInternalServerError,
		},
		{
			name:  "product service SKU not found",
			UID:   1,
			SKU:   100,
			count: 1,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return nil, internal_errors.ErrPreconditionFailed
				})
			},
			expectedErr: internal_errors.ErrPreconditionFailed,
		},
		{
			name:  "insufficient stocks",
			UID:   1,
			SKU:   100,
			count: 5,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return &models.GetProductResponse{Name: "Книга", Price: 400}, nil
				})
				lomsServiceMock.StocksInfoMock.Set(func(ctx context.Context, sku models.SKU) (int64, error) {
					require.Equal(t, models.SKU(100), sku)
					return int64(3), nil
				})
			},
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "number of stocks: 3 less than required count: 5",
		},
		{
			name:  "stocks service error",
			UID:   1,
			SKU:   100,
			count: 1,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return &models.GetProductResponse{Name: "Книга", Price: 400}, nil
				})
				lomsServiceMock.StocksInfoMock.Set(func(ctx context.Context, sku models.SKU) (int64, error) {
					require.Equal(t, models.SKU(100), sku)
					return int64(0), internal_errors.ErrInternalServerError
				})
			},
			expectedErr: internal_errors.ErrInternalServerError,
		},
		{
			name:  "repository error when adding item",
			UID:   1,
			SKU:   100,
			count: 3,
			setupMocks: func(repoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				productServiceMock.GetProductMock.Set(func(ctx context.Context, sku models.SKU) (*models.GetProductResponse, error) {
					require.Equal(t, models.SKU(100), sku)
					return &models.GetProductResponse{Name: "Книга", Price: 400}, nil
				})
				lomsServiceMock.StocksInfoMock.Set(func(ctx context.Context, sku models.SKU) (int64, error) {
					require.Equal(t, models.SKU(100), sku)
					return int64(10), nil
				})
				repoMock.AddItemMock.Set(func(ctx context.Context, uid models.UID, item models.CartItem) error {
					require.Equal(t, models.UID(1), uid)
					require.Equal(t, models.CartItem{SKU: 100, Count: 3}, item)
					return ErrRepository
				})
			},
			expectedErr: ErrRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repoMock, productServiceMock, lomsServiceMock, service := setup(t)

			tt.setupMocks(repoMock, productServiceMock, lomsServiceMock)

			err := service.AddProduct(ctx, tt.UID, tt.SKU, tt.count)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
