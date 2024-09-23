package service

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
)

// Function OrderCreate.
func (s *LomsService) OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error) {
	// Validate input data
	if err := validateOrderCreateRequest(req); err != nil {
		return nil, err
	}

	// Create order with status "new"
	order := models.Order{
		Status: models.OrderStatusNew,
		UserID: req.User,
		Items:  req.Items,
	}

	// Save order
	orderID, err := s.orderRepository.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Reserve stocks
	err = s.stockRepository.ReserveItems(ctx, order.Items)
	if err != nil {
		setStatusErr := s.orderRepository.SetStatus(ctx, orderID, models.OrderStatusFailed)
		if setStatusErr != nil {
			return nil, fmt.Errorf("failed to set order status failed: %w", setStatusErr)
		}
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Set status "awaiting payment"
	err = s.orderRepository.SetStatus(ctx, orderID, models.OrderStatusAwaitingPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to set order status awaiting payment: %w", err)
	}

	// Return orderID
	return &models.OrderCreateResponse{
		OrderID: orderID,
	}, nil
}

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
