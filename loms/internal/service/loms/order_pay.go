package service

import (
	"context"
	"fmt"

	"route256/loms/internal/models"

	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

// OrderPay function for payment of order.
func (s *LomsService) OrderPay(ctx context.Context, req *models.OrderPayRequest) error {
	// Tracer
	ctx, span := otel.Tracer("LomsService").Start(ctx, "OrderPay")
	defer span.End()

	// Validate input data
	if err := s.validateOrderPayRequest(req); err != nil {
		return err
	}

	// Use WithTx for transaction
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.processOrderPayment(ctx, tx, req)
	})
}

// validateOrderPayRequest validates the OrderPayRequest.
func (s *LomsService) validateOrderPayRequest(req *models.OrderPayRequest) error {
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	return nil
}

// processOrderPayment processes payment of order within transaction.
func (s *LomsService) processOrderPayment(ctx context.Context, tx pgx.Tx, req *models.OrderPayRequest) error {
	// Get info about order
	order, err := s.orderRepository.GetByID(ctx, req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check order status
	if err := s.validateOrderStatus(&order); err != nil {
		return err
	}

	// Remove reserved stock
	if err := s.removeReservedStock(ctx, tx, order.Items); err != nil {
		return err
	}

	// Set order status to "payed"
	if err := s.updateOrderStatus(ctx, req.OrderID, models.OrderStatusPayed); err != nil {
		return err
	}

	// Write event in outbox
	if err := s.createOutboxEvent(ctx, tx, req.OrderID, models.OrderStatusPayed, "OrderPayed"); err != nil {
		return err
	}

	return nil
}

// validateOrderStatus checks if order is in correct status for payment.
func (s *LomsService) validateOrderStatus(order *models.Order) error {
	if order.Status != models.OrderStatusAwaitingPayment {
		return fmt.Errorf("order is not in awaiting payment status: %w", internal_errors.ErrInvalidOrderStatus)
	}
	return nil
}

// removeReservedStock removes reserved items from stock.
func (s *LomsService) removeReservedStock(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	if err := s.stockRepository.RemoveReservedItems(ctx, tx, items); err != nil {
		return fmt.Errorf("failed to remove reserved stock: %w", err)
	}
	return nil
}
