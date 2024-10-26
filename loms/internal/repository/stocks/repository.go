package repository

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/pkg/metrics"
	"time"

	"route256/loms/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

// StockRepository
type StockRepository struct {
	queries sqlc.Querier
	pool    *pgxpool.Pool
}

// NewStockRepository creates a new instance of StockRepository.
func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// GetAvailableStockBySKU returns the available stock for specified SKU.
func (r *StockRepository) GetAvailableStockBySKU(ctx context.Context, SKU models.SKU) (uint64, error) {
	// Tracer
	ctx, span := otel.Tracer("StockRepository").Start(ctx, "GetAvailableStockBySKU")
	defer span.End()

	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		operation := "select"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if err := r.validateSKU(SKU); err != nil {
		return 0, err
	}

	// Get stock by SKU
	available, err := r.queries.GetAvailableStockBySKU(ctx, int32(SKU))
	if err != nil {
		return 0, fmt.Errorf("failed to get available stock: %w", err)
	}

	return uint64(available), nil
}

// ReserveItems reserves the specified count of products in the provided array of items.
func (r *StockRepository) ReserveItems(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	// Tracer
	ctx, span := otel.Tracer("StockRepository").Start(ctx, "ReserveItems")
	defer span.End()

	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		operation := "update"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	// Check transaction
	q := r.getQuerier(tx)

	// Reserve
	for _, item := range items {
		// Get available stock
		available, err := q.GetAvailableStockBySKU(ctx, int32(item.SKU))
		if err != nil {
			return fmt.Errorf("failed to get available stock for SKU %d: %w", item.SKU, err)
		}

		// Check available stock
		if available < int32(item.Count) {
			return fmt.Errorf("not enough stock for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Reserve product
		err = q.ReserveItems(ctx, &sqlc.ReserveItemsParams{
			Sku:      int32(item.SKU),
			Reserved: int64(item.Count),
		})
		if err != nil {
			return fmt.Errorf("failed to reserve items for SKU %d: %w", item.SKU, err)
		}
	}

	return nil
}

// RemoveReservedItems removes reserved stock for product.
func (r *StockRepository) RemoveReservedItems(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	// Tracer
	ctx, span := otel.Tracer("StockRepository").Start(ctx, "RemoveReservedItems")
	defer span.End()

	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		operation := "update"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	// Check transaction
	q := r.getQuerier(tx)

	// Remove reserved items
	for _, item := range items {
		// Get stock
		stock, err := q.GetStockBySKU(ctx, int32(item.SKU))
		if err != nil {
			return fmt.Errorf("failed to get stock for SKU %d: %w", item.SKU, err)
		}

		// Check
		if stock.Reserved < int64(item.Count) {
			return fmt.Errorf("not enough reserved stock for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Remove reserved stock and update total count
		err = q.RemoveReservedItems(ctx, &sqlc.RemoveReservedItemsParams{
			Sku:      int32(item.SKU),
			Reserved: int64(item.Count),
		})
		if err != nil {
			return fmt.Errorf("failed to remove reserved items for SKU %d: %w", item.SKU, err)
		}
	}

	return nil
}

// CancelReservedItems cancels reservation and makes the stock available again.
func (r *StockRepository) CancelReservedItems(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	// Tracer
	ctx, span := otel.Tracer("StockRepository").Start(ctx, "CancelReservedItems")
	defer span.End()

	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		operation := "update"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	// Check transaction
	q := r.getQuerier(tx)

	// Cancel reserved items
	for _, item := range items {
		// Get stock
		stock, err := q.GetStockBySKU(ctx, int32(item.SKU))
		if err != nil {
			return fmt.Errorf("failed to get stock for SKU %d: %w", item.SKU, err)
		}

		// Check
		if stock.Reserved < int64(item.Count) {
			return fmt.Errorf("not enough reserved stock to cancel for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Cancel reserved stock
		err = q.CancelReservedItems(ctx, &sqlc.CancelReservedItemsParams{
			Sku:      int32(item.SKU),
			Reserved: int64(item.Count),
		})
		if err != nil {
			return fmt.Errorf("failed to cancel reserved items for SKU %d: %w", item.SKU, err)
		}
	}

	return nil
}

// validateSKU function for validate SKU.
func (r *StockRepository) validateSKU(SKU models.SKU) error {
	if SKU < 1 {
		return fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	return nil
}

// validateItems function for validate items.
func (r *StockRepository) validateItems(items []models.Item) error {
	for _, item := range items {
		if err := r.validateSKU(item.SKU); err != nil {
			return err
		}
		if item.Count < 1 {
			return fmt.Errorf("count must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
	}
	return nil
}

// getQuerier returns sqlc.Querier based on provided transaction.
func (r *StockRepository) getQuerier(tx pgx.Tx) sqlc.Querier {
	if tx != nil {
		return sqlc.New(tx)
	}
	return r.queries
}
