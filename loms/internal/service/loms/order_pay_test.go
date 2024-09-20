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

// Function for tests the OrderPay method of LomsService.
func TestLomsService_OrderPay_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.OrderPayRequest
		setupMocks    func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest)
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful payment",
			req: &models.OrderPayRequest{
				OrderID: 1,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1001, Count: 2}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveRemoveItemsMock.Expect(order.Items).Return(nil)
				orderRepoMock.SetOrderStatusMock.Expect(req.OrderID, models.OrderStatusPayed).Return(nil)
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "invalid OrderID",
			req: &models.OrderPayRequest{
				OrderID: 0,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
			},
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "orderID must be greater than zero",
		},
		{
			name: "error getting order",
			req: &models.OrderPayRequest{
				OrderID: 2,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(models.Order{}, errors.New("db error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get order",
		},
		{
			name: "order not in awaiting payment status",
			req: &models.OrderPayRequest{
				OrderID: 3,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1002, Count: 1}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
			},
			expectedErr:   internal_errors.ErrInvalidOrderStatus,
			errorContains: "order is not in awaiting payment status",
		},
		{
			name: "error removing reserved stock",
			req: &models.OrderPayRequest{
				OrderID: 4,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1003, Count: 3}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveRemoveItemsMock.Expect(order.Items).Return(errors.New("reserve remove error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to remove reserved stock",
		},
		{
			name: "error setting order status to paid",
			req: &models.OrderPayRequest{
				OrderID: 5,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1004, Count: 4}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveRemoveItemsMock.Expect(order.Items).Return(nil)
				orderRepoMock.SetOrderStatusMock.Expect(req.OrderID, models.OrderStatusPayed).Return(errors.New("set status paid error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to set order status to payed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepoMock, stockRepoMock, svc := setup(t)

			tt.setupMocks(orderRepoMock, stockRepoMock, tt.req)

			err := svc.OrderPay(context.Background(), tt.req)
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
