package service

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

func TestCartService_Checkout_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		setupMocks    func(ctx context.Context, cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock)
		expectedOrder int64
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful checkout",
			UID:  1,
			setupMocks: func(ctx context.Context, cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				items := []models.CartItem{
					{SKU: 1001, Count: 2},
					{SKU: 1002, Count: 3},
				}
				cartRepoMock.GetItemsByUserIDMock.When(ctx, models.UID(1)).Then(items, nil)
				lomsServiceMock.OrderCreateMock.When(ctx, int64(1), items).Then(int64(2), nil)
				cartRepoMock.DeleteItemsByUserIDMock.When(ctx, models.UID(1)).Then(nil)
			},
			expectedOrder: 2,
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "error getting cart items",
			UID:  2,
			setupMocks: func(ctx context.Context, cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				cartRepoMock.GetItemsByUserIDMock.When(ctx, models.UID(2)).Then(nil, errors.New("db error"))
			},
			expectedOrder: 0,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get cart items",
		},
		{
			name: "error creating order in LomsService",
			UID:  3,
			setupMocks: func(ctx context.Context, cartRepoMock *mock.ICartRepositoryMock, productServiceMock *mock.IProductServiceMock, lomsServiceMock *mock.ILomsServiceMock) {
				items := []models.CartItem{
					{SKU: 1003, Count: 1},
				}
				cartRepoMock.GetItemsByUserIDMock.When(ctx, models.UID(3)).Then(items, nil)
				lomsServiceMock.OrderCreateMock.When(ctx, int64(3), items).Then(int64(0), errors.New("order create error"))
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

			tt.setupMocks(ctx, cartRepoMock, productServiceMock, lomsServiceMock)

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
