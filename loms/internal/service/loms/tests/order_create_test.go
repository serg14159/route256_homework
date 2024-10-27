package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	loms_service "route256/loms/internal/service/loms"
	"route256/loms/internal/service/loms/mock"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// Test function for OrderCreate method of LomsService.
func TestLomsService_OrderCreate_Table(t *testing.T) {
	tests := []struct {
		name       string
		req        *models.OrderCreateRequest
		setupMocks func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
			stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
			txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest)
		expectedResp  *models.OrderCreateResponse
		expectedErr   error
		errorContains string
	}{
		{
			name: "successful order creation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1001, Count: 2},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {

				newOrder := models.Order{
					Status: models.OrderStatusNew,
					UserID: req.User,
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
					require.Equal(t, order, newOrder)
					return models.OID(1), nil
				})

				callCount := 0

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					callCount++
					switch callCount {
					case 1:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderCreated", eventType)
						require.Equal(t, models.OID(1), event.OrderID)
						require.Equal(t, models.OrderStatusNew, event.Status)
						require.Equal(t, "OrderCreated", event.Additional)
						require.False(t, event.Time.IsZero())
					case 2:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderAwaitingPayment", eventType)
						require.Equal(t, models.OID(1), event.OrderID)
						require.Equal(t, models.OrderStatusAwaitingPayment, event.Status)
						require.Equal(t, "OrderAwaitingPayment", event.Additional)
						require.False(t, event.Time.IsZero())
					default:
						t.Errorf("unexpected number of CreateEvent calls: %d", callCount)
					}
					return nil
				})

				stockRepoMock.ReserveItemsMock.Set(func(ctx context.Context, tx pgx.Tx, items []models.Item) error {
					require.Equal(t, newOrder.Items, items)
					return nil
				})

				orderRepoMock.SetStatusMock.Set(func(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
					require.Equal(t, models.OID(1), orderID)
					require.Equal(t, models.OrderStatusAwaitingPayment, status)
					return nil
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn loms_service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
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
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {
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
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {
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
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {
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
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrBadRequest,
			errorContains: "count must be greater than zero",
		},
		{
			name: "error creating order",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1002, Count: 1},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {

				newOrder := models.Order{
					Status: models.OrderStatusNew,
					UserID: req.User,
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
					require.Equal(t, order, newOrder)
					return models.OID(0), errors.New("create order error")
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn loms_service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedResp:  nil,
			expectedErr:   errors.New("failed to create order"),
			errorContains: "failed to create order",
		},
		{
			name: "error reserving stocks",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1003, Count: 3},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {

				newOrder := models.Order{
					Status: models.OrderStatusNew,
					UserID: req.User,
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
					require.Equal(t, order, newOrder)
					return models.OID(2), nil
				})

				callCount := 0

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					callCount++
					switch callCount {
					case 1:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderCreated", eventType)
						require.Equal(t, models.OID(2), event.OrderID)
						require.Equal(t, models.OrderStatusNew, event.Status)
						require.Equal(t, "OrderCreated", event.Additional)
						require.False(t, event.Time.IsZero())
					case 2:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderFailed", eventType)
						require.Equal(t, models.OID(2), event.OrderID)
						require.Equal(t, models.OrderStatusFailed, event.Status)
						require.Equal(t, "OrderFailed", event.Additional)
						require.False(t, event.Time.IsZero())
					default:
						t.Errorf("unexpected number of CreateEvent calls: %d", callCount)
					}
					return nil
				})

				stockRepoMock.ReserveItemsMock.Set(func(ctx context.Context, tx pgx.Tx, items []models.Item) error {
					require.Equal(t, newOrder.Items, items)
					return errors.New("reserve items error")
				})

				orderRepoMock.SetStatusMock.Set(func(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
					require.Equal(t, models.OID(2), orderID)
					require.Equal(t, models.OrderStatusFailed, status)
					return nil
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn loms_service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedResp:  nil,
			expectedErr:   internal_errors.ErrStockReservation,
			errorContains: "failed to reserve stock",
		},
		{
			name: "error setting status after failed reservation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1004, Count: 4},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {

				newOrder := models.Order{
					Status: models.OrderStatusNew,
					UserID: req.User,
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
					require.Equal(t, order, newOrder)
					return models.OID(3), nil
				})

				callCount := 0

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					callCount++
					switch callCount {
					case 1:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderCreated", eventType)
						require.Equal(t, models.OID(3), event.OrderID)
						require.Equal(t, models.OrderStatusNew, event.Status)
						require.Equal(t, "OrderCreated", event.Additional)
						require.False(t, event.Time.IsZero())
					case 2:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderFailed", eventType)
						require.Equal(t, models.OID(3), event.OrderID)
						require.Equal(t, models.OrderStatusFailed, event.Status)
						require.Equal(t, "OrderFailed", event.Additional)
						require.False(t, event.Time.IsZero())
						return errors.New("write outbox failed")
					default:
						t.Errorf("unexpected number of CreateEvent calls: %d", callCount)
					}
					return nil
				})

				stockRepoMock.ReserveItemsMock.Set(func(ctx context.Context, tx pgx.Tx, items []models.Item) error {
					require.Equal(t, newOrder.Items, items)
					return errors.New("reserve items error")
				})

				callCountSetStatus := 0

				orderRepoMock.SetStatusMock.Set(func(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
					callCountSetStatus++
					switch callCountSetStatus {
					case 1:
						require.Equal(t, models.OrderStatusFailed, status)
						return errors.New("set status failed")
					default:
						t.Errorf("unexpected number of SetStatus calls: %d", callCountSetStatus)
					}
					return nil
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn loms_service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedResp:  nil,
			expectedErr:   errors.New("set status failed"),
			errorContains: "set status failed",
		},
		{
			name: "error setting status after successful reservation",
			req: &models.OrderCreateRequest{
				User: 1,
				Items: []models.Item{
					{SKU: 1005, Count: 5},
				},
			},
			setupMocks: func(ctx context.Context, orderRepoMock *mock.IOrderRepositoryMock,
				stockRepoMock *mock.IStockRepositoryMock, outboxRepoMock *mock.IOutboxRepositoryMock,
				txManagerMock *mock.ITxManagerMock, txMock *mock.TxMock, req *models.OrderCreateRequest) {

				newOrder := models.Order{
					Status: models.OrderStatusNew,
					UserID: req.User,
					Items:  req.Items,
				}

				orderRepoMock.CreateMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
					require.Equal(t, order, newOrder)
					return models.OID(4), nil
				})

				callCount := 0

				outboxRepoMock.CreateEventMock.Set(func(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
					callCount++
					switch callCount {
					case 1:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderCreated", eventType)
						require.Equal(t, models.OID(4), event.OrderID)
						require.Equal(t, models.OrderStatusNew, event.Status)
						require.Equal(t, "OrderCreated", event.Additional)
						require.False(t, event.Time.IsZero())
					case 2:
						event, ok := payload.(models.OrderEvent)
						require.True(t, ok)
						require.Equal(t, "OrderFailed", eventType)
						require.Equal(t, models.OID(4), event.OrderID)
						require.Equal(t, models.OrderStatusFailed, event.Status)
						require.Equal(t, "OrderFailed", event.Additional)
						require.False(t, event.Time.IsZero())
						return errors.New("write outbox failed")
					default:
						t.Errorf("unexpected number of CreateEvent calls: %d", callCount)
					}
					return nil
				})

				stockRepoMock.ReserveItemsMock.Set(func(ctx context.Context, tx pgx.Tx, items []models.Item) error {
					require.Equal(t, newOrder.Items, items)
					return nil
				})

				callCountSetStatus := 0
				orderRepoMock.SetStatusMock.Set(func(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
					callCountSetStatus++
					switch callCountSetStatus {
					case 1:
						require.Equal(t, models.OrderStatusAwaitingPayment, status)
						return errors.New("set status awaiting payment error")
					case 2:
						require.Equal(t, models.OrderStatusFailed, status)
						return nil
					default:
						t.Errorf("unexpected number of SetStatus calls: %d", callCountSetStatus)
					}
					return nil
				})

				txManagerMock.WithTxMock.Set(func(ctx context.Context, fn loms_service.WithTxFunc) error {
					return fn(ctx, txMock)
				})
			},
			expectedResp:  nil,
			expectedErr:   errors.New("set status awaiting payment error"),
			errorContains: "set status awaiting payment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			orderRepoMock, stockRepoMock, outboxRepoMock, _, txManagerMock, service := setup(t)
			txMock := mock.NewTxMock(t)

			tt.setupMocks(ctx, orderRepoMock, stockRepoMock, outboxRepoMock, txManagerMock, txMock, tt.req)

			resp, err := service.OrderCreate(ctx, tt.req)
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
			outboxRepoMock.MinimockFinish()
			txManagerMock.MinimockFinish()
		})
	}
}
