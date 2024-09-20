package service

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

// Function for tests the OrderCancel method of LomsService.
func TestLomsService_StocksInfo_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.StocksInfoRequest
		setupMocks    func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.StocksInfoRequest)
		expectedResp  *models.StocksInfoResponse
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful stock retrieval",
			req: &models.StocksInfoRequest{
				SKU: 1001,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.StocksInfoRequest) {
				stockRepoMock.GetAvailableStockBySKUMock.Expect(req.SKU).Return(uint64(50), nil)
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
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.StocksInfoRequest) {
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
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.StocksInfoRequest) {
				stockRepoMock.GetAvailableStockBySKUMock.Expect(req.SKU).Return(uint64(0), errors.New("db error"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get available stock for SKU 1002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepoMock, stockRepoMock, svc := setup(t)

			tt.setupMocks(orderRepoMock, stockRepoMock, tt.req)

			resp, err := svc.StocksInfo(context.Background(), tt.req)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
