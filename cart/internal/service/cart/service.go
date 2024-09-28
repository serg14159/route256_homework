package service

import (
	"context"
	"fmt"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
)

type ICartRepository interface {
	AddItem(ctx context.Context, UID models.UID, item models.CartItem) error
	DeleteItem(ctx context.Context, UID models.UID, SKU models.SKU) error
	DeleteItemsByUserID(ctx context.Context, UID models.UID) error
	GetItemsByUserID(ctx context.Context, UID models.UID) ([]models.CartItem, error)
}

type IProductService interface {
	GetProduct(ctx context.Context, SKU models.SKU) (*models.GetProductResponse, error)
}

type ILomsService interface {
	OrderCreate(ctx context.Context, user int64, items []models.CartItem) (int64, error)
	StocksInfo(ctx context.Context, SKU models.SKU) (int64, error)
}

type CartService struct {
	repository     ICartRepository
	productService IProductService
	lomsService    ILomsService
}

// NewService return instance of CartService.
func NewService(repository ICartRepository, productService IProductService, lomsService ILomsService) *CartService {
	return &CartService{
		repository:     repository,
		productService: productService,
		lomsService:    lomsService,
	}
}

// AddProduct function for add product into cart.
func (s *CartService) AddProduct(ctx context.Context, UID models.UID, SKU models.SKU, Count uint16) error {
	if UID < 1 || SKU < 1 || Count < 1 {
		return fmt.Errorf("UID, SKU and Count must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	_, err := s.productService.GetProduct(ctx, SKU)
	if err != nil {
		return err
	}

	stocks, err := s.lomsService.StocksInfo(ctx, SKU)
	if err != nil {
		return err
	}

	if stocks < int64(Count) {
		return fmt.Errorf("number of stocks: %d less than required count: %d, err: %w", stocks, Count, internal_errors.ErrBadRequest)
	}

	item := models.CartItem{
		SKU:   SKU,
		Count: Count,
	}

	err = s.repository.AddItem(ctx, UID, item)
	if err != nil {
		return err
	}

	return nil
}

// DelProduct function for delete product from cart.
func (s *CartService) DelProduct(ctx context.Context, UID models.UID, SKU models.SKU) error {
	if UID < 1 || SKU < 1 {
		return fmt.Errorf("UID and SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	err := s.repository.DeleteItem(ctx, UID, SKU)
	if err != nil {
		return err
	}

	return nil
}

// DelCart function for delete user cart.
func (s *CartService) DelCart(ctx context.Context, UID models.UID) error {
	if UID < 1 {
		return fmt.Errorf("UID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	err := s.repository.DeleteItemsByUserID(ctx, UID)
	if err != nil {
		return err
	}

	return nil
}

// GetCart function for get user cart.
func (s *CartService) GetCart(ctx context.Context, UID models.UID) ([]models.CartItemResponse, uint32, error) {
	if UID < 1 {
		return nil, 0, fmt.Errorf("UID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	cartItems, err := s.repository.GetItemsByUserID(ctx, UID)
	if err != nil {
		return nil, 0, err
	}

	var items []models.CartItemResponse
	var totalPrice uint32

	for _, item := range cartItems {
		product, err := s.productService.GetProduct(ctx, item.SKU)
		if err != nil {
			return nil, 0, fmt.Errorf("get product err: %w", err)
		}

		items = append(items, models.CartItemResponse{
			SKU:   item.SKU,
			Name:  product.Name,
			Count: item.Count,
			Price: product.Price,
		})
		totalPrice += uint32(item.Count) * product.Price
	}

	return items, totalPrice, nil
}

// Checkout function for create order.
func (s *CartService) Checkout(ctx context.Context, UID models.UID) (int64, error) {
	cartItems, err := s.repository.GetItemsByUserID(ctx, UID)
	if err != nil {
		return 0, fmt.Errorf("failed to get cart items: %w", err)
	}

	orderID, err := s.lomsService.OrderCreate(ctx, int64(UID), cartItems)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	err = s.repository.DeleteItemsByUserID(ctx, UID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete items by user ID: %w", err)
	}

	return orderID, nil
}
