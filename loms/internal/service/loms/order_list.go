package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"

	"go.opentelemetry.io/otel"
)

// OrderList returns all orders.
func (s *LomsService) OrderList(ctx context.Context) ([]models.Order, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsService").Start(ctx, "OrderList")
	defer span.End()

	orders, err := s.orderRepository.GetOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}
