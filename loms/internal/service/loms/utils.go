package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"

	"github.com/jackc/pgx/v5"
)

// updateOrderStatus updates order status in repository.
func (s *LomsService) updateOrderStatus(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
	if err := s.orderRepository.SetStatus(ctx, tx, orderID, status); err != nil {
		return fmt.Errorf("failed to set order status %s: %w", status, err)
	}
	return nil
}

// createOutboxEvent writes event to outbox.
func (s *LomsService) createOutboxEvent(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus, eventType string) error {
	if err := s.writeEventInOutbox(ctx, tx, eventType, orderID, status, eventType); err != nil {
		return fmt.Errorf("write event in outbox: %w", err)
	}
	return nil
}
