package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"route256/loms/internal/models"
	"route256/loms/internal/service/loms/mock"

	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/stretchr/testify/require"
)

// Function for tests the OrderCancel method of LomsService.
func TestLomsService_OrderCancel_Table(t *testing.T) {
	tests := []struct {
		name          string
		req           *models.OrderCancelRequest
		setupMocks    func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest)
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful cancel",
			req: &models.OrderCancelRequest{
				OrderID: 1,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1001, Count: 2}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveCancelItemsMock.Expect(order.Items).Return(nil)
				orderRepoMock.SetOrderStatusMock.Expect(models.OID(req.OrderID), models.OrderStatusCancelled).Return(nil)
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "invalid OrderID",
			req: &models.OrderCancelRequest{
				OrderID: 0,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest) {
			},
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "orderID must be greater than zero",
		},
		{
			name: "error when receiving order",
			req: &models.OrderCancelRequest{
				OrderID: 2,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest) {
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(models.Order{}, errors.New("db error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get order",
		},
		{
			name: "error cancel reservation",
			req: &models.OrderCancelRequest{
				OrderID: 3,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1003, Count: 1}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveCancelItemsMock.Expect(order.Items).Return(errors.New("reserve cancel error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to cancel stock reservation",
		},
		{
			name: "error update order status",
			req: &models.OrderCancelRequest{
				OrderID: 4,
			},
			setupMocks: func(orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1004, Count: 5}},
				}
				orderRepoMock.GetByOrderIDMock.Expect(req.OrderID).Return(order, nil)
				stockRepoMock.ReserveCancelItemsMock.Expect(order.Items).Return(nil)
				orderRepoMock.SetOrderStatusMock.Expect(models.OID(req.OrderID), models.OrderStatusCancelled).Return(errors.New("update status error"))
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to update order status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepoMock, stockRepoMock, service := setup(t)

			tt.setupMocks(orderRepoMock, stockRepoMock, tt.req)

			err := service.OrderCancel(context.Background(), tt.req)
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
