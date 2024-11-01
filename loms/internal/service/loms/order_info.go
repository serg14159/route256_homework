package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"go.opentelemetry.io/otel"
)

// OrderInfo аunction.
func (s *LomsService) OrderInfo(ctx context.Context, req *models.OrderInfoRequest) (*models.OrderInfoResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsService").Start(ctx, "OrderInfo")
	defer span.End()

	// Validate input data
	if req.OrderID < 1 {
		return nil, fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Get info about order
	order, err := s.orderRepository.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Return order
	return &models.OrderInfoResponse{
		Status: order.Status,
		User:   order.UserID,
		Items:  order.Items,
	}, nil
}
