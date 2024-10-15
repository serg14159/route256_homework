package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	service "route256/loms/internal/service/loms"
	"route256/loms/internal/service/loms/mock"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// Test function for OrderCancel method of LomsService
func TestLomsService_OrderCancel_Table(t *testing.T) {
	tests := []struct {
		name       string
		req        *models.OrderCancelRequest
		setupMocks func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
			stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
			txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest)
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful cancel",
			req: &models.OrderCancelRequest{
				OrderID: 1,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1001, Count: 2}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.CancelReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusCancelled).Return(nil)

				EventType := "OrderCancelled"
				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					require.Equal(t, EventType, eventType)
					event, ok := payload.(models.OrderEvent)
					require.True(t, ok)
					require.Equal(t, event.OrderID, req.OrderID)
					require.Equal(t, event.Status, models.OrderStatusCancelled)
					require.Equal(t, event.Additional, EventType)
					return nil
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   nil,
			errorContains: "",
		},
		{
			name: "invalid OrderID",
			req: &models.OrderCancelRequest{
				OrderID: 0,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {

			},
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "orderID must be greater than zero",
		},
		{
			name: "error when receiving order",
			req: &models.OrderCancelRequest{
				OrderID: 2,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {
				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(models.Order{}, errors.New("db error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get order",
		},
		{
			name: "error cancel reservation",
			req: &models.OrderCancelRequest{
				OrderID: 3,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1003, Count: 1}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.CancelReservedItemsMock.Expect(ctx, txMock, order.Items).Return(errors.New("reserve cancel error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to cancel stock reservation",
		},
		{
			name: "error updating order status",
			req: &models.OrderCancelRequest{
				OrderID: 4,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1004, Count: 5}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.CancelReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusCancelled).Return(errors.New("update status error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to update order status",
		},
		{
			name: "error writing event to outbox",
			req: &models.OrderCancelRequest{
				OrderID: 5,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCancelRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1005, Count: 3}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.CancelReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusCancelled).Return(nil)

				EventType := "OrderCancelled"
				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					require.Equal(t, EventType, eventType)

					event, ok := payload.(models.OrderEvent)
					require.True(t, ok)
					require.Equal(t, event.OrderID, req.OrderID)
					require.Equal(t, event.Status, models.OrderStatusCancelled)
					require.Equal(t, event.Additional, EventType)
					return errors.New("outbox error")
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "write event in outbox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			orderRepoMock, stockRepoMock, outboxRepoMock, _, txManagerMock, service := setup(t)
			txMock := mock.NewTxMock(t)

			tt.setupMocks(ctx, orderRepoMock, stockRepoMock, outboxRepoMock, txManagerMock, txMock, tt.req)

			err := service.OrderCancel(ctx, tt.req)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedErr) || (tt.errorContains != "" && strings.Contains(err.Error(), tt.errorContains)),
					"error must be %v or contain message: %s", tt.expectedErr, tt.errorContains)
			} else {
				require.NoError(t, err)
			}

			orderRepoMock.MinimockFinish()
			stockRepoMock.MinimockFinish()
			outboxRepoMock.MinimockFinish()
			txManagerMock.MinimockFinish()
		})
	}
}
