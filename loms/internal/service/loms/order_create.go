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

	// Create a transaction using WithTx
	var orderID models.OID
	err := s.txManager.WithTx(ctx, func(ctx context.Context, tx *pgx.Tx) error {
		// Create order with status "new"
		order := models.Order{
			Status: models.OrderStatusNew,
			UserID: req.User,
			Items:  req.Items,
		}

		// Save order
		var err error
		orderID, err = s.orderRepository.Create(ctx, tx, order)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Reserve stocks
		err = s.stockRepository.ReserveItems(ctx, tx, order.Items)
		if err != nil {
			setStatusErr := s.orderRepository.SetStatus(ctx, tx, orderID, models.OrderStatusFailed)
			if setStatusErr != nil {
				return fmt.Errorf("failed to set order status failed: %w", setStatusErr)
			}
			return fmt.Errorf("failed to reserve stock: %w", err)
		}

		// Set status "awaiting payment"
		err = s.orderRepository.SetStatus(ctx, tx, orderID, models.OrderStatusAwaitingPayment)
		if err != nil {
			return fmt.Errorf("failed to set order status awaiting payment: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return orderID
	return &models.OrderCreateResponse{
		OrderID: orderID,
	}, nil
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
