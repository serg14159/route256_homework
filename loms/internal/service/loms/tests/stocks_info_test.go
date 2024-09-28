package service_test

import (
	"context"
	"errors"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/service/loms/mock"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test function for StocksInfo method of LomsService.
func TestLomsService_StocksInfo_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.StocksInfoRequest
		setupMocks    func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.StocksInfoRequest)
		expectedResp  *models.StocksInfoResponse
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful stock retrieval",
			req: &models.StocksInfoRequest{
				SKU: 1001,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.StocksInfoRequest) {
				stockRepoMock.GetAvailableStockBySKUMock.Expect(ctx, req.SKU).Return(uint64(50), nil)
			},
			expectedResp: &models.StocksInfoResponse{
				Count: 50,
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "invalid SKU",
			req: &models.StocksInfoRequest{
				SKU: 0,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.StocksInfoRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "SKU must be greater than zero",
		},
		{
			name: "error getting available stock",
			req: &models.StocksInfoRequest{
				SKU: 1002,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.StocksInfoRequest) {
				stockRepoMock.GetAvailableStockBySKUMock.Expect(ctx, req.SKU).Return(uint64(0), errors.New("db error"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get available stock for SKU 1002",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			orderRepoMock, stockRepoMock, txManagerMock, service := setup(t)

			tt.setupMocks(ctx, orderRepoMock, stockRepoMock, txManagerMock, tt.req)

			resp, err := service.StocksInfo(ctx, tt.req)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, resp)
			}

			orderRepoMock.MinimockFinish()
			stockRepoMock.MinimockFinish()
			txManagerMock.MinimockFinish()
		})
	}
}
