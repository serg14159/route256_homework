package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
)

// Function OrderCancel.
func (s *LomsService) OrderCancel(ctx context.Context, req *models.OrderCancelRequest) error {
	// Validate input data
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	// Get order by orderID
	order, err := s.orderRepository.GetByID(ctx, req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Reserve сancel
	err = s.stockRepository.CancelReservedItems(ctx, order.Items)
	if err != nil {
		return fmt.Errorf("failed to cancel stock reservation: %w", err)
	}

	// Set order status "cancelled"
	err = s.orderRepository.SetStatus(ctx, models.OID(req.OrderID), models.OrderStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}
