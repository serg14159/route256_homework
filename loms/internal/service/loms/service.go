package service

import (
	"context"
	"route256/loms/internal/models"

	"github.com/jackc/pgx/v5"
)

type IOrderRepository interface {
	Create(ctx context.Context, tx *pgx.Tx, order models.Order) (models.OID, error)
	GetByID(ctx context.Context, tx *pgx.Tx, orderID models.OID) (models.Order, error)
	SetStatus(ctx context.Context, tx *pgx.Tx, orderID models.OID, status models.OrderStatus) error
}

type IStockRepository interface {
	GetAvailableStockBySKU(ctx context.Context, SKU models.SKU) (uint64, error)
	ReserveItems(ctx context.Context, tx *pgx.Tx, items []models.Item) error
	RemoveReservedItems(ctx context.Context, tx *pgx.Tx, items []models.Item) error
	CancelReservedItems(ctx context.Context, tx *pgx.Tx, items []models.Item) error
}

type WithTxFunc func(ctx context.Context, tx *pgx.Tx) error

type ITxManager interface {
	WithTx(ctx context.Context, fn WithTxFunc) error
}

type LomsService struct {
	orderRepository IOrderRepository
	stockRepository IStockRepository
	txManager       ITxManager
}

func NewService(orderRepository IOrderRepository, stockRepository IStockRepository, txManager ITxManager) *LomsService {
	return &LomsService{
		orderRepository: orderRepository,
		stockRepository: stockRepository,
		txManager:       txManager,
	}
}
