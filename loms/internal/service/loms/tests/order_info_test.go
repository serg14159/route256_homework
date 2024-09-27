package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/service/loms/mock"

	"github.com/stretchr/testify/require"
)

// Test for OrderInfo method of LomsService.
func TestLomsService_OrderInfo_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.OrderInfoRequest
		setupMocks    func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.OrderInfoRequest)
		expectedResp  *models.OrderInfoResponse
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful get order info",
			req: &models.OrderInfoRequest{
				OrderID: 1,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.OrderInfoRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1001, Count: 2}},
				}
				orderRepoMock.GetByIDMock.Expect(ctx, nil, models.OID(req.OrderID)).Return(order, nil)
			},
			expectedResp: &models.OrderInfoResponse{
				Status: models.OrderStatusNew,
				User:   1,
				Items:  []models.Item{{SKU: 1001, Count: 2}},
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "bad request with OrderID 0",
			req: &models.OrderInfoRequest{
				OrderID: 0,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.OrderInfoRequest) {

			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "orderID must be greater than zero",
		},
		{
			name: "error receive order",
			req: &models.OrderInfoRequest{
				OrderID: 2,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, txManagerMock *mock.ITxManagerMock, req *models.OrderInfoRequest) {
				orderRepoMock.GetByIDMock.Expect(ctx, nil, models.OID(req.OrderID)).Return(models.Order{}, errors.New("db error"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get order",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			orderRepoMock, stockRepoMock, txManagerMock, service := setup(t)

			tt.setupMocks(ctx, orderRepoMock, stockRepoMock, txManagerMock, tt.req)

			resp, err := service.OrderInfo(ctx, tt.req)
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
