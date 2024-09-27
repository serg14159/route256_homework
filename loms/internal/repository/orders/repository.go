package repository

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"route256/loms/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
)

// OrderRepository
type OrderRepository struct {
	queries sqlc.Querier
	conn    *pgx.Conn
}

// Function NewOrderRepository creates a new instance of OrderRepository.
func NewOrderRepository(conn *pgx.Conn) *OrderRepository {
	return &OrderRepository{
		queries: sqlc.New(conn),
		conn:    conn,
	}
}

// Create adds a new order to repository and returns unique orderID.
func (r *OrderRepository) Create(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error) {
	// Validate input data
	if err := validateOrder(order); err != nil {
		return 0, err
	}

	// Check transaction
	var q sqlc.Querier
	if tx != nil {
		q = sqlc.New(tx)
	} else {
		q = r.queries
	}

	// Create order
	createdOrder, err := q.CreateOrder(ctx, &sqlc.CreateOrderParams{
		UserID: order.UserID,
		Status: string(order.Status),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range order.Items {
		_, err := q.CreateOrderItem(ctx, &sqlc.CreateOrderItemParams{
			OrderID: &createdOrder.ID,
			Sku:     int32(item.SKU),
			Count:   int16(item.Count),
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	return models.OID(createdOrder.ID), nil
}

// Function GetByID return order by orderID.
func (r *OrderRepository) GetByID(ctx context.Context, tx pgx.Tx, orderID models.OID) (models.Order, error) {
	// Validate input data
	if orderID < 1 {
		return models.Order{}, fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	// Check transaction
	var q sqlc.Querier
	if tx != nil {
		q = sqlc.New(tx)
	} else {
		q = r.queries
	}

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
func (r *OrderRepository) SetStatus(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
	// Validate input data
	if orderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	if !isValidOrderStatus(status) {
		return fmt.Errorf("invalid order status: %w", internal_errors.ErrPreconditionFailed)
	}

	// Check transaction
	var q sqlc.Querier
	if tx != nil {
		q = sqlc.New(tx)
	} else {
		q = r.queries
	}

	// Update order status
	err := q.SetOrderStatus(ctx, &sqlc.SetOrderStatusParams{
		ID:     orderID,
		Status: string(status),
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
