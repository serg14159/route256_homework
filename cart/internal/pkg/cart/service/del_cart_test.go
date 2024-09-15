package service

import (
	"context"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function for tests the DelCart method of CartService.
func TestCartService_DelCart_Table(t *testing.T) {
	tests := []struct {
		name      string
		UID       models.UID
		deleteErr error
		wantErr   error
	}{
		{
			name:    "successful delete cart",
			UID:     1,
			wantErr: nil,
		},
		{
			name:    "delete empty cart",
			UID:     5,
			wantErr: nil,
		},
		{
			name:    "bad request",
			UID:     0,
			wantErr: internal_errors.ErrBadRequest,
		},
		{
			name:      "repository error",
			UID:       1,
			deleteErr: ErrRepository,
			wantErr:   ErrRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repoMock, _, service := setup(t)

			ctx := context.Background()

			if tt.UID < 1 {
				err := service.DelCart(ctx, tt.UID)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			repoMock.DeleteItemsByUserIDMock.Expect(ctx, tt.UID).Return(tt.deleteErr)

			err := service.DelCart(ctx, tt.UID)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
