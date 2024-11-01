package service

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

// OrderCreate function.
func (s *LomsService) OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsService").Start(ctx, "OrderCreate")
	defer span.End()

	// Validate input data
	if err := validateOrderCreateRequest(req); err != nil {
		return nil, err
	}

	// Create order with status "new" and write event in outbox
	orderID, err := s.createOrderWithEvent(ctx, req, "OrderCreated")
	if err != nil {
		return nil, err
	}

	// Reserve stocks and update order
	err = s.reserveStocksAndUpdateOrder(ctx, orderID, req.Items, "OrderAwaitingPayment")
	if err != nil {
		// Set order status "failed"
		errUpdate := s.updateOrderStatusAndCreateEvent(ctx, orderID, models.OrderStatusFailed, "OrderFailed")
		if errUpdate != nil {
			return nil, fmt.Errorf("%w : %w", err, errUpdate)
		}
		return nil, fmt.Errorf("%w : %w", err, internal_errors.ErrPreconditionFailed)
	}

	// Return orderID
	return &models.OrderCreateResponse{
		OrderID: orderID,
	}, nil
}

// createOrderWithEvent creates an order and writes an event to the outbox.
func (s *LomsService) createOrderWithEvent(ctx context.Context, req *models.OrderCreateRequest, eventType string) (models.OID, error) {
	// Create order
	orderID, err := s.createOrder(ctx, req)
	if err != nil {
		return orderID, err
	}

	// Write event in outbox
	if err := s.createOutboxEvent(ctx, nil, orderID, models.OrderStatusNew, eventType); err != nil {
		return orderID, err
	}

	return orderID, err
}

// reserveStocksAndUpdateOrder reserves stocks and updates order status.
func (s *LomsService) reserveStocksAndUpdateOrder(ctx context.Context, orderID models.OID, items []models.Item, eventType string) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Reserve stocks
		if err := s.reserveStocks(ctx, tx, items); err != nil {
			return fmt.Errorf("reserve stocks: %w", err)
		}

		// Update order status
		if err := s.updateOrderStatus(ctx, orderID, models.OrderStatusAwaitingPayment); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}

		// Write event in outbox
		if err := s.createOutboxEvent(ctx, tx, orderID, models.OrderStatusAwaitingPayment, eventType); err != nil {
			return err
		}

		return nil
	})
}

// updateOrderStatusAndCreateEvent updates order status and writes event within transaction.
func (s *LomsService) updateOrderStatusAndCreateEvent(ctx context.Context, orderID models.OID, status models.OrderStatus, eventType string) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Update order status
		if err := s.updateOrderStatus(ctx, orderID, status); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}

		// Write event in outbox
		if err := s.createOutboxEvent(ctx, tx, orderID, status, eventType); err != nil {
			return err
		}

		return nil
	})
}

// createOrder handles order creation and returns the created order ID.
func (s *LomsService) createOrder(ctx context.Context, req *models.OrderCreateRequest) (models.OID, error) {
	order := models.Order{
		Status: models.OrderStatusNew,
		UserID: req.User,
		Items:  req.Items,
	}

	orderID, err := s.orderRepository.Create(ctx, order)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	return orderID, nil
}

// reserveStocks handles stock reservation and sets order status accordingly.
func (s *LomsService) reserveStocks(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	err := s.stockRepository.ReserveItems(ctx, tx, items)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	return err
}

// validateOrderCreateRequest function for validate request data.
func validateOrderCreateRequest(req *models.OrderCreateRequest) error {
	if req.User < 1 {
		return fmt.Errorf("userID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item: %w", internal_errors.ErrBadRequest)
	}

	for _, item := range req.Items {
		if item.SKU < 1 {
			return fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
		if item.Count < 1 {
			return fmt.Errorf("count must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
	}

	return nil
}
