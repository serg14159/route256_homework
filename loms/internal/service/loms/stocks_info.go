package service

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"go.opentelemetry.io/otel"
)

// StocksInfo function.
func (s *LomsService) StocksInfo(ctx context.Context, req *models.StocksInfoRequest) (*models.StocksInfoResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsService").Start(ctx, "StocksInfo")
	defer span.End()

	// Validate input data
	if req.SKU < 1 {
		return nil, fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Get available stock by SKU
	count, err := s.stockRepository.GetAvailableStockBySKU(ctx, req.SKU)
	if err != nil {
		return nil, fmt.Errorf("failed to get available stock for SKU %d: %w", req.SKU, err)
	}

	return &models.StocksInfoResponse{
		Count: count,
	}, nil
}
