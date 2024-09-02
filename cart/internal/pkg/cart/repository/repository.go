package repository

import (
	"context"
	"fmt"
	"route256/cart/internal/models"
	"sort"
	"sync"

	internal_errors "route256/cart/internal/pkg/errors"
)

type Storage = map[models.UID]map[models.SKU]models.CartItem

type Repository struct {
	mu      sync.Mutex
	storage Storage
}

func NewCartRepository() *Repository {
	return &Repository{
		mu:      sync.Mutex{},
		storage: make(Storage),
	}
}

// Function for adding item to cart.
func (r *Repository) AddItem(ctx context.Context, UID models.UID, item models.CartItem) error {
	if UID < 1 || item.SKU < 1 {
		return fmt.Errorf("UID and SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.storage[UID] == nil {
		r.storage[UID] = make(map[models.SKU]models.CartItem)
	}

	foundItem, ok := r.storage[UID][item.SKU]
	if ok {
		item.Count += foundItem.Count
	}
	r.storage[UID][item.SKU] = item

	return nil
}

// Function for delete item from cart.
func (r *Repository) DeleteItem(ctx context.Context, UID models.UID, SKU models.SKU) error {
	if UID < 1 || SKU < 1 {
		return fmt.Errorf("UID and SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.storage[UID] != nil {
		delete(r.storage[UID], SKU)
	}

	return nil
}

// Function for delete cart.
func (r *Repository) DeleteItemsByUserID(ctx context.Context, UID models.UID) error {
	if UID < 1 {
		return fmt.Errorf("UID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.storage, UID)

	return nil
}

// Function for getting items from cart.
func (r *Repository) GetItemsByUserID(ctx context.Context, UID models.UID) ([]models.CartItem, error) {
	if UID < 1 {
		return nil, fmt.Errorf("UID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cart, ok := r.storage[UID]
	if !ok || len(cart) == 0 {
		return nil, fmt.Errorf("cart for UID not found in storage: %w", internal_errors.ErrNotFound)
	}

	items := make([]models.CartItem, 0, len(cart))
	for _, item := range cart {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SKU < items[j].SKU
	})

	return items, nil
}
