package repository

import (
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
func (r *StockRepository) LoadStocks(filePath string) error {
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

// Function GetAvailableStockBySKU returns the available stock for specified SKU.
func (r *StockRepository) GetAvailableStockBySKU(SKU models.SKU) (uint64, error) {
	// Validate input data
	if SKU < 1 {
		return 0, fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get stock by SKU
	stock, exists := r.stocks[SKU]
	if !exists {
		return 0, internal_errors.ErrNotFound
	}

	available := stock.TotalCount - stock.Reserved
	return available, nil
}

// Function Reserve reserves the specified count of product.
func (r *StockRepository) Reserve(SKU models.SKU, count uint16) error {
	// Validate input data
	if SKU < 1 || count < 1 {
		return fmt.Errorf("SKU and count must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get stock by SKU
	stock, exists := r.stocks[SKU]
	if !exists {
		return internal_errors.ErrNotFound
	}

	// Check
	available := stock.TotalCount - stock.Reserved
	if available < uint64(count) {
		return fmt.Errorf("not enough stock: %w", internal_errors.ErrPreconditionFailed)
	}

	// Update
	stock.Reserved += uint64(count)
	r.stocks[SKU] = stock

	return nil
}

// Function ReserveRemove removes reserved stock for product.
func (r *StockRepository) ReserveRemove(SKU models.SKU, count uint16) error {
	// Validate input data
	if SKU < 1 || count < 1 {
		return fmt.Errorf("SKU and count must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get stock by SKU
	stock, exists := r.stocks[SKU]
	if !exists {
		return internal_errors.ErrNotFound
	}

	// Check
	if stock.Reserved < uint64(count) {
		return fmt.Errorf("not enough reserved stock: %w", internal_errors.ErrPreconditionFailed)
	}

	// Update
	stock.Reserved -= uint64(count)
	stock.TotalCount -= uint64(count)
	r.stocks[SKU] = stock

	return nil
}

// Function ReserveCancel cancels reservation and makes the stock available again.
func (r *StockRepository) ReserveCancel(SKU models.SKU, count uint16) error {
	// Validate input data
	if SKU < 1 || count < 1 {
		return fmt.Errorf("SKU and count must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check
	stock, exists := r.stocks[SKU]
	if !exists {
		return internal_errors.ErrNotFound
	}

	if stock.Reserved < uint64(count) {
		return fmt.Errorf("not enough reserved stock to cancel: %w", internal_errors.ErrPreconditionFailed)
	}

	// Update
	stock.Reserved -= uint64(count)
	r.stocks[SKU] = stock

	return nil
}
