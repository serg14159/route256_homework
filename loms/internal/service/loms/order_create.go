package service

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
)

// OrderCreate function.
func (s *LomsService) OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error) {
	// Validate input data
	if err := validateOrderCreateRequest(req); err != nil {
		return nil, err
	}

	// Create var
	var orderID models.OID
	var orderStatus models.OrderStatus
	var eventType string
	var err error

	// Create order with status "new"
	orderStatus = models.OrderStatusNew
	err = s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Create order
		orderID, err = s.createOrder(ctx, tx, req)
		if err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		// Write event in outbox
		eventType = "OrderCreated"
		err = s.writeEventInOutbox(ctx, tx, eventType, orderID, orderStatus, eventType)
		if err != nil {
			return fmt.Errorf("write event in outbox: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reserve stocks
	orderStatus = models.OrderStatusAwaitingPayment
	err = s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Reserve stocks
		err := s.reserveStocks(ctx, tx, req.Items)
		if err != nil {
			return fmt.Errorf("reserve stocks: %w", err)
		}
		// Set order status "awaiting payment"
		err = s.updateOrderStatus(ctx, tx, orderID, orderStatus)
		if err != nil {
			return fmt.Errorf("update order status: %w", err)
		}
		// Write event in outbox
		eventType = "OrderAwaitingPayment"
		err = s.writeEventInOutbox(ctx, tx, eventType, orderID, orderStatus, eventType)
		if err != nil {
			return fmt.Errorf("write event in outbox: %w", err)
		}

		return nil
	})

	if err != nil {
		// Set order status "failed"
		orderStatus = models.OrderStatusFailed
		errSetStatusFailed := s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			// Set order status
			errSetStatus := s.updateOrderStatus(ctx, tx, orderID, orderStatus)
			if err != nil {
				return fmt.Errorf("update order status: %w : %w", errSetStatus, internal_errors.ErrInternalServerError)
			}
			// Write event in outbox
			eventType = "OrderFailed"
			errSetStatus = s.writeEventInOutbox(ctx, tx, eventType, orderID, orderStatus, eventType)
			if err != nil {
				return fmt.Errorf("write event in outbox: %w : %w", errSetStatus, internal_errors.ErrInternalServerError)
			}

			return nil
		})

		if errSetStatusFailed != nil {
			return nil, fmt.Errorf("%w : %w", err, errSetStatusFailed)
		}
		return nil, fmt.Errorf("%w : %w", err, internal_errors.ErrPreconditionFailed)
	}

	// Return orderID
	return &models.OrderCreateResponse{
		OrderID: orderID,
	}, nil
}

// createOrder handles order creation and returns the created order ID.
func (s *LomsService) createOrder(ctx context.Context, tx pgx.Tx, req *models.OrderCreateRequest) (models.OID, error) {
	order := models.Order{
		Status: models.OrderStatusNew,
		UserID: req.User,
		Items:  req.Items,
	}

	orderID, err := s.orderRepository.Create(ctx, tx, order)
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

// updateOrderStatus function for update order status.
func (s *LomsService) updateOrderStatus(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
	err := s.orderRepository.SetStatus(ctx, tx, orderID, status)
	if err != nil {
		return fmt.Errorf("failed to set order status '%s': %w", status, err)
	}
	return nil
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
