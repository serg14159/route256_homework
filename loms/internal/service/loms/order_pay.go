package service

import (
	"context"
	"fmt"

	"route256/loms/internal/models"

	internal_errors "route256/loms/internal/pkg/errors"
)

// Function OrderPay.
func (s *LomsService) OrderPay(ctx context.Context, req *models.OrderPayRequest) error {
	// Validate input data
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Get info about order
	order, err := s.orderRepository.GetByID(ctx, req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check order status
	if order.Status != models.OrderStatusAwaitingPayment {
		return fmt.Errorf("order is not in awaiting payment status: %w", internal_errors.ErrInvalidOrderStatus)
	}

	// Remove reserve stock
	err = s.stockRepository.RemoveReservedItems(ctx, order.Items)
	if err != nil {
		return fmt.Errorf("failed to remove reserved stock: %w", err)
	}

	// Set order status "payed"
	err = s.orderRepository.SetStatus(ctx, req.OrderID, models.OrderStatusPayed)
	if err != nil {
		//s.stockRepository.RollbackRemoveReserved(order.Items)
		return fmt.Errorf("failed to set order status to payed: %w", err)
	}

	return nil
}
