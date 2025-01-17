package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"route256/cart/internal/service/cart/mock"

	"github.com/stretchr/testify/require"
)

// TestCartService_Checkout_Table function for tests the Checkout method of CartService.
func TestCartService_Checkout_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		setupMocks    func(cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock)
		expectedOrder int64
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful checkout",
			UID:  1,
			setupMocks: func(cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				items := []models.CartItem{
					{SKU: 1001, Count: 2},
					{SKU: 1002, Count: 3},
				}
				cartRepoMock.GetItemsByUserIDMock.Set(func(ctx context.Context, uid models.UID) ([]models.CartItem, error) {
					require.Equal(t, models.UID(1), uid)
					return items, nil
				})
				lomsServiceMock.OrderCreateMock.Set(func(ctx context.Context, user int64, itemsParam []models.CartItem) (int64, error) {
					require.Equal(t, int64(1), user)
					require.Equal(t, items, itemsParam)
					return int64(2), nil
				})
				cartRepoMock.DeleteItemsByUserIDMock.Set(func(ctx context.Context, uid models.UID) error {
					require.Equal(t, models.UID(1), uid)
					return nil
				})
			},
			expectedOrder: 2,
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "error getting cart items",
			UID:  2,
			setupMocks: func(cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				cartRepoMock.GetItemsByUserIDMock.Set(func(ctx context.Context, uid models.UID) ([]models.CartItem, error) {
					require.Equal(t, models.UID(2), uid)
					return nil, errors.New("db error")
				})
			},
			expectedOrder: 0,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get cart items",
		},
		{
			name: "error creating order in LomsService",
			UID:  3,
			setupMocks: func(cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				items := []models.CartItem{
					{SKU: 1003, Count: 1},
				}
				cartRepoMock.GetItemsByUserIDMock.Set(func(ctx context.Context, uid models.UID) ([]models.CartItem, error) {
					require.Equal(t, models.UID(3), uid)
					return items, nil
				})
				lomsServiceMock.OrderCreateMock.Set(func(ctx context.Context, user int64, itemsParam []models.CartItem) (int64, error) {
					require.Equal(t, int64(3), user)
					require.Equal(t, items, itemsParam)
					return int64(0), errors.New("order create error")
				})
			},
			expectedOrder: 0,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to create order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cartRepoMock, productServiceMock, lomsServiceMock, service := setup(t)

			tt.setupMocks(cartRepoMock, productServiceMock, lomsServiceMock)

			orderID, err := service.Checkout(ctx, tt.UID)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOrder, orderID)
			}
		})
	}
}
