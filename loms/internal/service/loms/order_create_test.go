package service

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

// Function for tests the OrderCreate method of LomsService.
func TestLomsService_OrderCreate_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.OrderCreateRequest
		setupMocks    func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest)
		expectedResp  *models.OrderCreateResponse
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful create order",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1001, Count: 2},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: int64(req.User),
					Items:  req.Items,
				}
				orderRepoMock.CreateMock.Expect(ctx, order).Return(int64(1), nil)
				stockRepoMock.ReserveItemsMock.Expect(ctx, req.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, models.OID(1), models.OrderStatusAwaitingPayment).Return(nil)
			},
			expectedResp: &models.OrderCreateResponse{
				OrderID: 1,
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "invalid UserID",
			req: &models.OrderCreateRequest{
				User: 0,
				Items: []models.Item{
					{SKU: 1001, Count: 2},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "userID must be greater than zero",
		},
		{
			name: "empty items list",
			req: &models.OrderCreateRequest{
				User:  1,
				Items: []models.Item{},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "order must contain at least one item",
		},
		{
			name: "invalid SKU",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 0, Count: 2},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "SKU must be greater than zero",
		},
		{
			name: "invalid count",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1001, Count: 0},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "count must be greater than zero",
		},
		{
			name: "error create order",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1002, Count: 1},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: int64(req.User),
					Items:  req.Items,
				}
				orderRepoMock.CreateMock.Expect(ctx, order).Return(int64(0), errors.New("create order error"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to create order",
		},
		{
			name: "error reservation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1003, Count: 3},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: int64(req.User),
					Items:  req.Items,
				}
				orderRepoMock.CreateMock.Expect(ctx, order).Return(int64(2), nil)
				stockRepoMock.ReserveItemsMock.Expect(ctx, req.Items).Return(errors.New("reserve items error"))
				orderRepoMock.SetStatusMock.Expect(ctx, models.OID(2), models.OrderStatusFailed).Return(nil)
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to reserve stock",
		},
		{
			name: "status update error after failed reservation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1004, Count: 4},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: int64(req.User),
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Expect(ctx, order).Return(int64(3), nil)
				stockRepoMock.ReserveItemsMock.Expect(ctx, req.Items).Return(errors.New("reserve items error"))
				orderRepoMock.SetStatusMock.Expect(ctx, models.OID(3), models.OrderStatusFailed).Return(errors.New("set status failed"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to set order status failed",
		},
		{
			name: "status update error after successful reservation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1005, Count: 5},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCreateRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: int64(req.User),
					Items:  req.Items,
				}
				orderRepoMock.CreateMock.Expect(ctx, order).Return(int64(4), nil)
				stockRepoMock.ReserveItemsMock.Expect(ctx, req.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, models.OID(4), models.OrderStatusAwaitingPayment).Return(errors.New("set status awaiting payment error"))
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to set order status awaiting payment",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			orderRepoMock, stockRepoMock, svc := setup(t)

			tt.setupMocks(ctx, orderRepoMock, stockRepoMock, tt.req)

			resp, err := svc.OrderCreate(ctx, tt.req)
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
		})
	}
}
