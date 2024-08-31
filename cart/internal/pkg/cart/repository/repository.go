package repository

import (
	"context"
	"errors"
	"route256/cart/internal/models"
	"sync"
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
func (r *Repository) AddCartItem(ctx context.Context, UID models.UID, item models.CartItem) error {
	if UID < 1 || item.SKU < 1 {
		return errors.New("UID and SKU must be defined")
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
func (r *Repository) DelItem(UID models.UID, SKU models.SKU) error {
	if UID < 1 || SKU < 1 {
		return errors.New("UID and SKU must be defined")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.storage[UID] != nil {
		delete(r.storage[UID], SKU)
	}

	return nil
}

// Function for delete cart.
func (r *Repository) DelCart(UID models.UID) error {
	if UID < 1 {
		return errors.New("UID must be defined")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.storage, UID)

	return nil
}

// Function for getting items from cart.
func (r *Repository) GetCart(UID models.UID) ([]models.CartItem, uint32, error) {
	if UID < 1 {
		return nil, 0, errors.New("UID must be defined")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cart, ok := r.storage[UID]
	if !ok || len(cart) == 0 {
		return nil, 0, errors.New("cart not found")
	}

	items := make([]models.CartItem, 0, len(cart))
	var totalPrice uint32
	for _, item := range cart {
		items = append(items, item)
		totalPrice += uint32(item.Count) * item.Price
	}

	return items, totalPrice, nil
}
