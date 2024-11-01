package repository

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/pkg/shard_manager"
	"route256/loms/internal/repository/sqlc"
	"strconv"
	"time"

	"route256/loms/internal/pkg/metrics"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

type IShardManager interface {
	GetShardIndex(key shard_manager.ShardKey) shard_manager.ShardIndex
	GetShardIndexFromID(id int64) shard_manager.ShardIndex
	GetShard(shard_manager.ShardIndex) (*pgxpool.Pool, error)
	CloseShards()
}

// OrderRepository.
type OrderRepository struct {
	shardManager IShardManager
}

// NewOrderRepository creates a new instance of OrderRepository.
func NewOrderRepository(shardManager IShardManager) *OrderRepository {
	return &OrderRepository{
		shardManager: shardManager,
	}
}

// Create adds a new order to repository and returns unique orderID.
func (r *OrderRepository) Create(ctx context.Context, order models.Order) (models.OID, error) {
	// Tracer
	ctx, span := otel.Tracer("OrderRepository").Start(ctx, "Create")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		operation := "Create"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if err := validateOrder(order); err != nil {
		return 0, err
	}

	// Determine shard
	shardIndex := r.shardManager.GetShardIndex(shard_manager.ShardKey(strconv.FormatInt(order.UserID, 10)))
	pool, err := r.shardManager.GetShard(shardIndex)
	if err != nil {
		return 0, err
	}

	// Start transaction on the shard
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	// Create order
	orderID, err := q.CreateOrder(ctx, &sqlc.CreateOrderParams{
		Column1: shardIndex,
		UserID:  order.UserID,
		Name:    string(order.Status),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range order.Items {
		_, err := q.CreateOrderItem(ctx, &sqlc.CreateOrderItemParams{
			OrderID: &orderID,
			Sku:     int32(item.SKU),
			Count:   int16(item.Count),
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return models.OID(orderID), nil
}

// GetByID return order by orderID.
func (r *OrderRepository) GetByID(ctx context.Context, orderID models.OID) (models.Order, error) {
	// Tracer
	ctx, span := otel.Tracer("OrderRepository").Start(ctx, "GetByID")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		operation := "GetByID"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if orderID < 1 {
		return models.Order{}, fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Determine shard
	shardIndex := r.shardManager.GetShardIndexFromID(orderID)
	pool, err := r.shardManager.GetShard(shardIndex)
	if err != nil {
		return models.Order{}, err
	}

	q := sqlc.New(pool)

	// Get order
	order, err := q.GetOrderByID(ctx, orderID)
	if err != nil {
		return models.Order{}, internal_errors.ErrNotFound
	}

	// Get items
	items, err := q.GetOrderItems(ctx, &orderID)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get order items: %w", err)
	}

	// Convert
	var modelItems []models.Item
	for _, item := range items {
		modelItems = append(modelItems, models.Item{
			SKU:   models.SKU(item.Sku),
			Count: uint16(item.Count),
		})
	}

	return models.Order{
		Status: models.OrderStatus(order.Status),
		UserID: order.UserID,
		Items:  modelItems,
	}, nil
}

// SetStatus updates the status of an existing order.
func (r *OrderRepository) SetStatus(ctx context.Context, orderID models.OID, status models.OrderStatus) error {
	// Tracer
	ctx, span := otel.Tracer("OrderRepository").Start(ctx, "SetStatus")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		operation := "SetStatus"
		metrics.IncDBQueryCounter(operation)
		metrics.ObserveDBQueryDuration(operation, duration)
	}()

	// Validate input data
	if orderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	if !isValidOrderStatus(status) {
		return fmt.Errorf("invalid order status: %w", internal_errors.ErrPreconditionFailed)
	}

	// Determine shard
	shardIndex := r.shardManager.GetShardIndexFromID(orderID)
	pool, err := r.shardManager.GetShard(shardIndex)
	if err != nil {
		return err
	}

	// Check transaction
	q := sqlc.New(pool)

	// Update order status
	err = q.SetOrderStatus(ctx, &sqlc.SetOrderStatusParams{
		ID:   orderID,
		Name: string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to set order status: %w", err)
	}

	return nil
}

// validateOrder validate the order.
func validateOrder(order models.Order) error {
	if order.UserID < 1 {
		return fmt.Errorf("userID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order must contain at least one item: %w", internal_errors.ErrBadRequest)
	}
	for _, item := range order.Items {
		if item.SKU < 1 {
			return fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
		if item.Count < 1 {
			return fmt.Errorf("count must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
	}
	return nil
}

// isValidOrderStatus check status is valid.
func isValidOrderStatus(status models.OrderStatus) bool {
	switch status {
	case models.OrderStatusNew,
		models.OrderStatusAwaitingPayment,
		models.OrderStatusFailed,
		models.OrderStatusPayed,
		models.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
