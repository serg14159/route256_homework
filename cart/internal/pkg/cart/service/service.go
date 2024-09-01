package service

import (
	"context"
	"errors"
	"route256/cart/internal/models"
)

type ICartRepository interface {
	AddItem(ctx context.Context, UID models.UID, item models.CartItem) error
	DeleteItem(ctx context.Context, UID models.UID, SKU models.SKU) error
	DeleteItemsByUserID(ctx context.Context, UID models.UID) error
	GetItemsByUserID(ctx context.Context, UID models.UID) ([]models.CartItem, error)
}

type CartService struct {
	repository ICartRepository
}

func NewService(repository ICartRepository) *CartService {
	return &CartService{repository: repository}
}

func (s *CartService) AddProduct(ctx context.Context, UID models.UID, SKU models.SKU, Count uint16) error {
	if UID < 1 || SKU < 1 || Count < 1 {
		return errors.New("fail validation")
	}

	return nil
}

func (s *CartService) DelProduct(ctx context.Context, UID models.UID, SKU models.SKU) error {
	if UID < 1 || SKU < 1 {
		return errors.New("fail validation")
	}

	return nil
}

func (s *CartService) DelCart(ctx context.Context, UID models.UID) error {
	if UID < 1 {
		return errors.New("fail validation")
	}

	return nil
}

func (s *CartService) GetCart(ctx context.Context, UID models.UID) ([]models.CartItemResponse, uint32, error) {
	if UID < 1 {
		return nil, 0, errors.New("fail validation")
	}

	return nil, 0, nil
}
