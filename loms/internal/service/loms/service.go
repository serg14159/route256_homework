package service

import (
	"context"
	"route256/loms/internal/models"
)

type IOrderRepository interface {
	Create(ctx context.Context, order models.Order) (int64, error)
	GetByID(ctx context.Context, orderID models.OID) (models.Order, error)
	SetStatus(ctx context.Context, orderID models.OID, status models.OrderStatus) error
}

type IStockRepository interface {
	GetAvailableStockBySKU(ctx context.Context, SKU models.SKU) (uint64, error)
	ReserveItems(ctx context.Context, items []models.Item) error
	RemoveReservedItems(ctx context.Context, items []models.Item) error
	CancelReservedItems(ctx context.Context, items []models.Item) error
	RollbackRemoveReserved(removedItems []models.Item)
}

type LomsService struct {
	orderRepository IOrderRepository
	stockRepository IStockRepository
}

func NewService(orderRepository IOrderRepository, stockRepository IStockRepository) *LomsService {
	return &LomsService{
		orderRepository: orderRepository,
		stockRepository: stockRepository,
	}
}
