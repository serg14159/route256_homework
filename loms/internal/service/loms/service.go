package service

import (
	"context"
	"route256/loms/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type IOrderRepository interface {
	Create(ctx context.Context, tx pgx.Tx, order models.Order) (models.OID, error)
	GetByID(ctx context.Context, tx pgx.Tx, orderID models.OID) (models.Order, error)
	SetStatus(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error
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
	SendOrderEvent(ctx context.Context, event *models.OrderEvent) error
}

type LomsService struct {
	orderRepository IOrderRepository
	stockRepository IStockRepository
	txManager       ITxManager
	producer        IProducer
}

// NewService return instance of LomsService.
func NewService(orderRepository IOrderRepository, stockRepository IStockRepository, txManager ITxManager, producer IProducer) *LomsService {
	return &LomsService{
		orderRepository: orderRepository,
		stockRepository: stockRepository,
		txManager:       txManager,
		producer:        producer,
	}
}

// sendEventToKafka sends OrderEvent to Kafka.
func (s *LomsService) sendEventToKafka(ctx context.Context, orderID models.OID, status models.OrderStatus, additional string) error {
	event := &models.OrderEvent{
		OrderID:    models.OID(orderID),
		Status:     status,
		Time:       time.Now(),
		Additional: additional,
	}
	err := s.producer.SendOrderEvent(ctx, event)
	return err
}
