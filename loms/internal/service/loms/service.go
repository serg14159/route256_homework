package service

import (
	"context"
	"route256/loms/internal/models"

	"github.com/jackc/pgx/v5"
)

type IOrderRepository interface {
	Create(ctx context.Context, order models.Order) (models.OID, error)
	GetByID(ctx context.Context, orderID models.OID) (models.Order, error)
	SetStatus(ctx context.Context, orderID models.OID, status models.OrderStatus) error
}

type IStockRepository interface {
	GetAvailableStockBySKU(ctx context.Context, SKU models.SKU) (uint64, error)
	ReserveItems(ctx context.Context, tx pgx.Tx, items []models.Item) error
	RemoveReservedItems(ctx context.Context, tx pgx.Tx, items []models.Item) error
	CancelReservedItems(ctx context.Context, tx pgx.Tx, items []models.Item) error
}

type WithTxFunc func(ctx context.Context, tx pgx.Tx) error

type ITxManager interface {
	WithTx(ctx context.Context, fn WithTxFunc) error
}

type IProducer interface {
	SendOutboxEvent(ctx context.Context, event *models.OutboxEvent) error
}

type IOutboxRepository interface {
	CreateEvent(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error
	FetchNextMsg(ctx context.Context, tx pgx.Tx) (*models.OutboxEvent, error)
	MarkAsSent(ctx context.Context, tx pgx.Tx, eventID int64) error
}

type LomsService struct {
	orderRepository  IOrderRepository
	stockRepository  IStockRepository
	outboxRepository IOutboxRepository
	txManager        ITxManager
	producer         IProducer
}

// NewService return instance of LomsService.
func NewService(orderRepository IOrderRepository, stockRepository IStockRepository, outboxRepository IOutboxRepository, txManager ITxManager, producer IProducer) *LomsService {
	return &LomsService{
		orderRepository:  orderRepository,
		stockRepository:  stockRepository,
		outboxRepository: outboxRepository,
		txManager:        txManager,
		producer:         producer,
	}
}
