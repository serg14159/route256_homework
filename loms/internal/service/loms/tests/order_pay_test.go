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

// Test function for OrderPay method of LomsService.
func TestLomsService_OrderPay_Table(t *testing.T) {
	tests := []struct {
		name       string
		req        *models.OrderPayRequest
		setupMocks func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
			stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
			txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest)
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful payment",
			req: &models.OrderPayRequest{
				OrderID: 1,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock, stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock, txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1001, Count: 2}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.RemoveReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusPayed).Return(nil)

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					event, ok := payload.(models.OrderEvent)
					require.True(t, ok)
					require.Equal(t, "OrderPayed", eventType)
					require.Equal(t, models.OID(req.OrderID), event.OrderID)
					require.Equal(t, models.OrderStatusPayed, event.Status)
					require.Equal(t, "OrderPayed", event.Additional)
					require.False(t, event.Time.IsZero())
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
			req: &models.OrderPayRequest{
				OrderID: 0,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
			},
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "orderID must be greater than zero",
		},
		{
			name: "error getting order",
			req: &models.OrderPayRequest{
				OrderID: 2,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(models.Order{}, errors.New("db error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to get order",
		},
		{
			name: "order not in awaiting payment status",
			req: &models.OrderPayRequest{
				OrderID: 3,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusNew,
					UserID: 1,
					Items:  []models.Item{{SKU: 1002, Count: 1}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInvalidOrderStatus,
			errorContains: "order is not in awaiting payment status",
		},
		{
			name: "error removing reserved stock",
			req: &models.OrderPayRequest{
				OrderID: 4,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1003, Count: 3}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.RemoveReservedItemsMock.Expect(ctx, txMock, order.Items).Return(errors.New("reserve remove error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to remove reserved stock",
		},
		{
			name: "error setting order status to payed",
			req: &models.OrderPayRequest{
				OrderID: 5,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1004, Count: 4}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.RemoveReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusPayed).Return(errors.New("set status payed error"))

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedErr:   internal_errors.ErrInternalServerError,
			errorContains: "failed to set order status to payed",
		},
		{
			name: "error writing event to outbox",
			req: &models.OrderPayRequest{
				OrderID: 6,
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderPayRequest) {
				order := models.Order{
					Status: models.OrderStatusAwaitingPayment,
					UserID: 1,
					Items:  []models.Item{{SKU: 1005, Count: 5}},
				}

				orderRepoMock.GetByIDMock.Expect(ctx, txMock, models.OID(req.OrderID)).Return(order, nil)
				stockRepoMock.RemoveReservedItemsMock.Expect(ctx, txMock, order.Items).Return(nil)
				orderRepoMock.SetStatusMock.Expect(ctx, txMock, models.OID(req.OrderID), models.OrderStatusPayed).Return(nil)

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					event, ok := payload.(models.OrderEvent)
					require.True(t, ok)
					require.Equal(t, "OrderPayed", eventType)
					require.Equal(t, models.OID(req.OrderID), event.OrderID)
					require.Equal(t, models.OrderStatusPayed, event.Status)
					require.Equal(t, "OrderPayed", event.Additional)
					require.False(t, event.Time.IsZero())
					return errors.New("outbox write error")
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

			err := service.OrderPay(ctx, tt.req)
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
