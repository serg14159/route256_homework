package repository

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"sync"
)

// Storage
type Storage = map[models.OID]models.Order

// OrderRepository
type OrderRepository struct {
	mu           sync.Mutex
	storage      Storage
	countOrderID models.OID
}

// Function NewOrderRepository creates a new instance of OrderRepository.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		mu:           sync.Mutex{},
		storage:      make(Storage),
		countOrderID: 0,
	}
}

// Function Create add new order to repository and returns unique orderID.
func (r *OrderRepository) Create(ctx context.Context, order models.Order) (models.OID, error) {
	// Validate input data
	err := validateOrder(order)
	if err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Increment countOrderID and generate new orderID
	r.countOrderID++
	orderID := r.countOrderID

	// Save order in storage
	r.storage[orderID] = order

	return orderID, nil
}

// Function GetByID return order by orderID.
func (r *OrderRepository) GetByID(ctx context.Context, orderID models.OID) (models.Order, error) {
	// Validate input data
	if orderID < 1 {
		return models.Order{}, fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get order by orderID
	order, exists := r.storage[orderID]
	if !exists {
		return models.Order{}, internal_errors.ErrNotFound
	}

	return order, nil
}

// Function SetStatus update status of existing order.
func (r *OrderRepository) SetStatus(ctx context.Context, orderID models.OID, status models.OrderStatus) error {
	// Validate input data
	if orderID < 1 {
		return fmt.Errorf("orderID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	if !isValidOrderStatus(status) {
		return fmt.Errorf("invalid order status: %w", internal_errors.ErrPreconditionFailed)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get order by orderID
	order, exists := r.storage[orderID]
	if !exists {
		return fmt.Errorf("order with orderID not found: %w", internal_errors.ErrPreconditionFailed)
	}

	// Update order status
	order.Status = status
	r.storage[orderID] = order

	return nil
}

// Function validateOrder validate the order.
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

// Function isValidOrderStatus check status is valid.
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
