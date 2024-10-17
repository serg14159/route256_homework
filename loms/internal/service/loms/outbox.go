package service

import (
	"context"
	"errors"
	"fmt"
	"route256/loms/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

// writeEventInOutbox writes event to outbox table.
func (s *LomsService) writeEventInOutbox(ctx context.Context, tx pgx.Tx, eventType string, orderID models.OID, status models.OrderStatus, additional string) error {
	event := models.OrderEvent{
		OrderID:    orderID,
		Status:     status,
		Time:       time.Now(),
		Additional: additional,
	}

	err := s.outboxRepository.CreateEvent(ctx, tx, eventType, event)
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return nil
}

// ProcessOutbox continuously processes outbox events, sending them to Kafka and marking them as processed.
func (s *LomsService) ProcessOutbox(ctx context.Context) error {
	for {
		err := s.processNextOutboxEvent(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
	}
}

// processNextOutboxEvent processes the next outbox event.
func (s *LomsService) processNextOutboxEvent(ctx context.Context) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		event, err := s.outboxRepository.FetchNextMsg(ctx, tx)
		if err != nil {
			return err
		}

		// Send event to Kafka
		if err := s.producer.SendOutboxEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to send event to Kafka: %w", err)
		}

		// Mark event as sent
		if err := s.outboxRepository.MarkAsSent(ctx, tx, event.ID); err != nil {
			return fmt.Errorf("failed to mark event as sent: %w", err)
		}

		return nil
	})
}
