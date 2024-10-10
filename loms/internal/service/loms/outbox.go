package service

import (
	"context"
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
		var noMoreMessages bool
		err := s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			event, err := s.outboxRepository.FetchNextMsg(ctx, tx)
			if err != nil {
				if err == pgx.ErrNoRows {
					noMoreMessages = true
					return nil
				}
				return fmt.Errorf("failed to fetch next outbox message: %w", err)
			}

			// Send event in Kafka
			err = s.producer.SendOutboxEvent(ctx, event)
			if err != nil {
				return fmt.Errorf("failed to send event to Kafka: %w", err)
			}

			// Mark event as sent
			err = s.outboxRepository.MarkAsSent(ctx, tx, event.ID)
			if err != nil {
				return fmt.Errorf("failed to mark event as sent: %w", err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		if noMoreMessages {
			break
		}
	}

	return nil
}
