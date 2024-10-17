package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
)

// OrderCancel function for cancel order.
func (s *LomsService) OrderCancel(ctx context.Context, req *models.OrderCancelRequest) error {
	// Validate input data
	if err := s.validateOrderCancelRequest(req); err != nil {
		return err
	}

	// Use WithTx for transaction
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.processOrderCancel(ctx, tx, req)
	})
}

// validateOrderCancelRequest validates OrderCancelRequest.
func (s *LomsService) validateOrderCancelRequest(req *models.OrderCancelRequest) error {
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	return nil
}

// processOrderCancel processes cancel within transaction.
func (s *LomsService) processOrderCancel(ctx context.Context, tx pgx.Tx, req *models.OrderCancelRequest) error {
	// Get order by orderID
	order, err := s.orderRepository.GetByID(ctx, tx, req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check order status
	if err := s.validateOrderStatusForCancel(&order); err != nil {
		return err
	}

	// Cancel reserved stock
	if err := s.cancelReservedStock(ctx, tx, order.Items); err != nil {
		return err
	}

	// Set order status to "cancelled"
	if err := s.updateOrderStatus(ctx, tx, req.OrderID, models.OrderStatusCancelled); err != nil {
		return err
	}

	// Write event in outbox
	if err := s.createOutboxEvent(ctx, tx, req.OrderID, models.OrderStatusCancelled, "OrderCancelled"); err != nil {
		return err
	}

	return nil
}

// validateOrderStatusForCancel check if order can be cancelled.
func (s *LomsService) validateOrderStatusForCancel(order *models.Order) error {
	if order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusPayed {
		return fmt.Errorf("order is already in a final state: %w", internal_errors.ErrInvalidOrderStatus)
	}
	return nil
}

// cancelReservedStock cancel reservation of items in stock.
func (s *LomsService) cancelReservedStock(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	if err := s.stockRepository.CancelReservedItems(ctx, tx, items); err != nil {
		return fmt.Errorf("failed to cancel stock reservation: %w", err)
	}
	return nil
}
