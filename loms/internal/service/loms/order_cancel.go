package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
)

// OrderCancel function.
func (s *LomsService) OrderCancel(ctx context.Context, req *models.OrderCancelRequest) error {
	// Validate input data
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Use WithTx for transaction
	err := s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Get order by orderID
		order, err := s.orderRepository.GetByID(ctx, tx, req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		// Check status
		if order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusPayed {
			return fmt.Errorf("order is already in a final state: %w", internal_errors.ErrInvalidOrderStatus)
		}

		// Reserve сancel
		err = s.stockRepository.CancelReservedItems(ctx, tx, order.Items)
		if err != nil {
			return fmt.Errorf("failed to cancel stock reservation: %w", err)
		}

		// Set order status "cancelled"
		err = s.orderRepository.SetStatus(ctx, tx, req.OrderID, models.OrderStatusCancelled)
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		return nil
	})
	return err
}
