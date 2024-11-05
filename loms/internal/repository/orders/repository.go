package repository

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/pkg/shard_manager"
	"route256/loms/internal/repository/sqlc"
	"sort"
	"strconv"
	"sync"
	"time"

	"route256/loms/internal/pkg/metrics"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

const allOrdersChannelBufferSize = 100

type IShardManager interface {
	GetShardIndex(key shard_manager.ShardKey) shard_manager.ShardIndex
	GetShardIndexFromID(id int64) shard_manager.ShardIndex
	GetShard(shard_manager.ShardIndex) (*pgxpool.Pool, error)
	GetShards() []*pgxpool.Pool
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
	defer setMetrics("Create", startTime)

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
	defer setMetrics("GetByID", startTime)

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
	defer setMetrics("SetStatus", startTime)

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

// GetOrders returns all orders.
func (r *OrderRepository) GetOrders(ctx context.Context) ([]models.Order, error) {
	// Tracer
	ctx, span := otel.Tracer("OrderRepository").Start(ctx, "GetOrders")
	defer span.End()

	startTime := time.Now()
	defer setMetrics("GetOrders", startTime)

	shards := r.shardManager.GetShards()
	allOrders := make(chan models.Order, allOrdersChannelBufferSize)
	errCh := make(chan error, len(shards))
	var wg sync.WaitGroup

	// Run
	for _, pool := range shards {
		wg.Add(1)
		go func(pool *pgxpool.Pool) {
			defer wg.Done()
			if err := r.processShard(ctx, pool, allOrders); err != nil {
				errCh <- err
			}
		}(pool)
	}

	// Close channels
	go func() {
		wg.Wait()
		close(allOrders)
		close(errCh)
	}()

	// Collecting all orders
	var orders []models.Order
	for order := range allOrders {
		orders = append(orders, order)
	}

	// Collecting errors
	if err := collectErrors(errCh); err != nil {
		return nil, err
	}

	// Sort orders by ID desc
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].OrderID > orders[j].OrderID
	})

	return orders, nil
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

// setMetrics set metrics of operation.
func setMetrics(operation string, startTime time.Time) {
	duration := time.Since(startTime)
	metrics.IncDBQueryCounter(operation)
	metrics.ObserveDBQueryDuration(operation, duration)
}

// processShard processes one shard.
func (r *OrderRepository) processShard(ctx context.Context, pool *pgxpool.Pool, allOrders chan<- models.Order) error {
	q := sqlc.New(pool)

	orders, err := q.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get orders from shard: %w", err)
	}

	for _, order := range orders {
		modelOrder, err := r.buildModelOrder(ctx, q, order)
		if err != nil {
			return err
		}
		allOrders <- modelOrder
	}

	return nil
}

// buildModelOrder build order model.
func (r *OrderRepository) buildModelOrder(ctx context.Context, q *sqlc.Queries, order *sqlc.GetAllOrdersRow) (models.Order, error) {
	modelOrder := models.Order{
		OrderID: order.ID,
		Status:  models.OrderStatus(order.Status),
		UserID:  order.UserID,
	}

	items, err := q.GetOrderItems(ctx, &order.ID)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get items for order %d: %w", order.ID, err)
	}

	for _, item := range items {
		modelOrder.Items = append(modelOrder.Items, models.Item{
			SKU:   models.SKU(item.Sku),
			Count: uint16(item.Count),
		})
	}

	return modelOrder, nil
}

// collectErrors collects errors from the error channel.
func collectErrors(errCh <-chan error) error {
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
