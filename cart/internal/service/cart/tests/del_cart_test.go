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

// TestCartService_DelCart_Table function for tests the DelCart method of CartService.
func TestCartService_DelCart_Table(t *testing.T) {
	tests := []struct {
		name          string
		UID           models.UID
		setupMocks    func(ctx context.Context, repoMock *mock.ICartRepositoryMock)
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful delete cart",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock) {
				repoMock.DeleteItemsByUserIDMock.When(ctx, models.UID(1)).Then(nil)
			},
			expectedErr: nil,
		},
		{
			name: "delete empty cart",
			UID:  5,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock) {
				repoMock.DeleteItemsByUserIDMock.When(ctx, models.UID(5)).Then(nil)
			},
			expectedErr: nil,
		},
		{
			name: "bad request with UID 0",
			UID:  0,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock) {
			},
			expectedErr: internal_errors.ErrBadRequest,
		},
		{
			name: "repository error",
			UID:  1,
			setupMocks: func(ctx context.Context, repoMock *mock.ICartRepositoryMock) {
				repoMock.DeleteItemsByUserIDMock.When(ctx, models.UID(1)).Then(ErrRepository)
			},
			expectedErr: ErrRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repoMock, _, _, service := setup(t)

			tt.setupMocks(ctx, repoMock)

			err := service.DelCart(ctx, tt.UID)

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
