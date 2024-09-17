package service

import "route256/loms/internal/models"

type IOrderRepository interface {
	CreateOrder(order models.Order) (int64, error)
	GetByOrderID(orderID models.OID) (models.Order, error)
	SetOrderStatus(orderID models.OID, status models.OrderStatus) error
}

type IStockRepository interface {
	GetAvailableStockBySKU(SKU models.SKU) (uint64, error)
	ReserveItems(items []models.Item) error
	ReserveRemoveItems(items []models.Item) error
	ReserveCancelItems(items []models.Item) error
}

type CartService struct {
	orderRepository IOrderRepository
	stockRepository IStockRepository
}

func NewService(orderRepository IOrderRepository, stockRepository IStockRepository) *CartService {
	return &CartService{
		orderRepository: orderRepository,
		stockRepository: stockRepository,
	}
}
