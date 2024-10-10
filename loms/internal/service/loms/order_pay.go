package service

import (
	"context"
	"fmt"
	"log"

	"route256/loms/internal/models"

	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
)

// OrderPay function.
func (s *LomsService) OrderPay(ctx context.Context, req *models.OrderPayRequest) error {
	// Validate input data
	if req.OrderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Use WithTx for transaction
	err := s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Get info about order
		order, err := s.orderRepository.GetByID(ctx, tx, req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		// Check order status
		if order.Status != models.OrderStatusAwaitingPayment {
			return fmt.Errorf("order is not in awaiting payment status: %w", internal_errors.ErrInvalidOrderStatus)
		}

		// Remove reserved stock
		err = s.stockRepository.RemoveReservedItems(ctx, tx, order.Items)
		if err != nil {
			return fmt.Errorf("failed to remove reserved stock: %w", err)
		}

		// Set order status "payed"
		err = s.orderRepository.SetStatus(ctx, tx, req.OrderID, models.OrderStatusPayed)
		if err != nil {
			return fmt.Errorf("failed to set order status to payed: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Send order status "payed" to Kafka
	err = s.sendEventToKafka(ctx, req.OrderID, models.OrderStatusPayed, "OrderPay")
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
	}

	return nil
}
