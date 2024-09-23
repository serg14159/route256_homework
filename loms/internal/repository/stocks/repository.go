package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	"sync"
)

// Storage
type Storage = map[models.SKU]models.Stock

// StockRepository
type StockRepository struct {
	mu     sync.Mutex
	stocks Storage
}

// Function NewStockRepository creates a new instance of StockRepository.
func NewStockRepository() *StockRepository {
	repo := &StockRepository{
		mu:     sync.Mutex{},
		stocks: make(Storage),
	}
	return repo
}

// Function LoadStocks loads data in StockRepository from the specified file.
func (r *StockRepository) LoadStocks(ctx context.Context, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read stock data file: %w", err)
	}

	var stocks []models.Stock
	if err := json.Unmarshal(data, &stocks); err != nil {
		return fmt.Errorf("failed to unmarshal stock data: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range stocks {
		stock := stocks[i]
		r.stocks[stock.SKU] = stock
	}

	return nil
}

// Function for validate SKU.
func (r *StockRepository) validateSKU(SKU models.SKU) error {
	if SKU < 1 {
		return fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}
	return nil
}

// Function for validate items.
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

// Function computeAvailableStock calculate available stock.
func (r *StockRepository) computeAvailableStock(stock models.Stock) uint64 {
	return stock.TotalCount - stock.Reserved
}

// Function GetAvailableStockBySKU returns the available stock for specified SKU.
func (r *StockRepository) GetAvailableStockBySKU(ctx context.Context, SKU models.SKU) (uint64, error) {
	// Validate input data
	if err := r.validateSKU(SKU); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get stock by SKU
	stock, exists := r.stocks[SKU]
	if !exists {
		return 0, internal_errors.ErrNotFound
	}

	available := r.computeAvailableStock(stock)
	return available, nil
}

// Function ReserveItems reserves the specified count of products in the provided array of items.
func (r *StockRepository) ReserveItems(ctx context.Context, items []models.Item) error {
	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// List of successfully reserved items for potential rollback
	var reservedItems []models.Item

	// Reserve
	for _, item := range items {
		stock, exists := r.stocks[item.SKU]
		if !exists {
			r.rollbackReservations(reservedItems)
			return fmt.Errorf("not found stock for SKU %d: %w", item.SKU, internal_errors.ErrNotFound)
		}

		// Check available
		available := r.computeAvailableStock(stock)
		if available < uint64(item.Count) {
			r.rollbackReservations(reservedItems)
			return fmt.Errorf("not enough stock for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Reserve product
		stock.Reserved += uint64(item.Count)
		r.stocks[item.SKU] = stock

		// Add successfully reserved items for potential rollback
		reservedItems = append(reservedItems, item)
	}

	return nil
}

// Function cancelReservations rolls back all previously successful reservations.
func (r *StockRepository) rollbackReservations(reservedItems []models.Item) {
	for _, item := range reservedItems {
		stock := r.stocks[item.SKU]
		stock.Reserved -= uint64(item.Count)
		r.stocks[item.SKU] = stock
	}
}

// Function RemoveReservedItems removes reserved stock for product.
func (r *StockRepository) RemoveReservedItems(ctx context.Context, items []models.Item) error {
	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// List of successfully removed items for potential rollback
	var removedItems []models.Item

	// Remove reserved
	for _, item := range items {
		stock, exists := r.stocks[item.SKU]
		if !exists {
			r.RollbackRemoveReserved(removedItems)
			return fmt.Errorf("not found stock for SKU %d: %w", item.SKU, internal_errors.ErrNotFound)
		}

		// Check
		if stock.Reserved < uint64(item.Count) {
			r.RollbackRemoveReserved(removedItems)
			return fmt.Errorf("not enough reserved stock for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Remove reserved stock and update total count
		stock.Reserved -= uint64(item.Count)
		stock.TotalCount -= uint64(item.Count)
		r.stocks[item.SKU] = stock

		// Add to the list of removed reservations for potential rollback
		removedItems = append(removedItems, item)
	}

	return nil
}

// Function RollbackRemoveReserved rolls back all previously successful remove from reserved stock.
func (r *StockRepository) RollbackRemoveReserved(removedItems []models.Item) {
	for _, item := range removedItems {
		stock := r.stocks[item.SKU]
		stock.Reserved += uint64(item.Count)
		stock.TotalCount += uint64(item.Count)
		r.stocks[item.SKU] = stock
	}
}

// Function CancelReservedItems cancels reservation and makes the stock available again.
func (r *StockRepository) CancelReservedItems(ctx context.Context, items []models.Item) error {
	// Validate input data
	if err := r.validateItems(items); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// List of successfully cancelled reservations for potential rollback
	var cancelledItems []models.Item

	// Cancel reserved
	for _, item := range items {
		stock, exists := r.stocks[item.SKU]
		if !exists {
			r.rollbackCancelReservation(cancelledItems)
			return fmt.Errorf("not found stock for SKU %d: %w", item.SKU, internal_errors.ErrNotFound)
		}

		// Check
		if stock.Reserved < uint64(item.Count) {
			r.rollbackCancelReservation(cancelledItems)
			return fmt.Errorf("not enough reserved stock to cancel for SKU %d: %w", item.SKU, internal_errors.ErrPreconditionFailed)
		}

		// Cancel reserved stock
		stock.Reserved -= uint64(item.Count)
		r.stocks[item.SKU] = stock

		// Add to the list of cancelled reservations for potential rollback
		cancelledItems = append(cancelledItems, item)
	}

	return nil
}

// Function rollbackCancelReservation rolls back all previously successful cancel of reserved stock
func (r *StockRepository) rollbackCancelReservation(cancelledItems []models.Item) {
	for _, item := range cancelledItems {
		stock := r.stocks[item.SKU]
		stock.Reserved += uint64(item.Count)
		r.stocks[item.SKU] = stock
	}
}
