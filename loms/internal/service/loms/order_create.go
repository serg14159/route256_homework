package service

import (
	"context"
	"fmt"
	"log"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5"
)

// OrderCreate function.
func (s *LomsService) OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error) {
	// Validate input data
	if err := validateOrderCreateRequest(req); err != nil {
		return nil, err
	}

	// Create a transaction using WithTx
	var orderID models.OID
	var orderStatus models.OrderStatus
	var err error

	// Create order with status "new"
	orderStatus = models.OrderStatusNew
	orderID, err = s.createOrder(ctx, nil, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Send order status "new" to Kafka
	err = s.sendEventToKafka(ctx, orderID, orderStatus, "OrderCreate")
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
	}

	// Tx
	orderStatus = models.OrderStatusAwaitingPayment
	err = s.txManager.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Reserve stocks
		err := s.reserveStocks(ctx, tx, req.Items)
		if err != nil {
			return fmt.Errorf("failed to reserve stocks: %w", err)
		}
		// Set order status "awaiting payment"
		err = s.updateOrderStatus(ctx, tx, orderID, orderStatus)
		if err != nil {
			return fmt.Errorf("failed to reserve stocks: %w", err)
		}

		return nil
	})

	if err != nil {
		// Set order status "failed"
		orderStatus = models.OrderStatusFailed
		errSetStatus := s.updateOrderStatus(ctx, nil, orderID, orderStatus)
		if errSetStatus != nil {
			return nil, fmt.Errorf("%w : %w", err, internal_errors.ErrInternalServerError)
		}

		// Send order status "failed" to Kafka
		err = s.sendEventToKafka(ctx, orderID, orderStatus, "OrderCreate")
		if err != nil {
			log.Printf("Failed to send Kafka message: %v", err)
		}

		return nil, fmt.Errorf("%w : %w", err, internal_errors.ErrPreconditionFailed)
	}

	// Send order status "awaiting payment" to Kafka
	err = s.sendEventToKafka(ctx, orderID, orderStatus, "OrderCreate")
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
	}

	// Return orderID
	return &models.OrderCreateResponse{
		OrderID: orderID,
	}, nil
}

// createOrder handles order creation and returns the created order ID.
func (s *LomsService) createOrder(ctx context.Context, tx pgx.Tx, req *models.OrderCreateRequest) (models.OID, error) {
	order := models.Order{
		Status: models.OrderStatusNew,
		UserID: req.User,
		Items:  req.Items,
	}

	orderID, err := s.orderRepository.Create(ctx, tx, order)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	return orderID, nil
}

// reserveStocks handles stock reservation and sets order status accordingly.
func (s *LomsService) reserveStocks(ctx context.Context, tx pgx.Tx, items []models.Item) error {
	err := s.stockRepository.ReserveItems(ctx, tx, items)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	return err
}

// updateOrderStatus function for update order status.
func (s *LomsService) updateOrderStatus(ctx context.Context, tx pgx.Tx, orderID models.OID, status models.OrderStatus) error {
	err := s.orderRepository.SetStatus(ctx, tx, orderID, status)
	if err != nil {
		return fmt.Errorf("failed to set order status '%s': %w", status, err)
	}
	return nil
}

// validateOrderCreateRequest function for validate request data.
func validateOrderCreateRequest(req *models.OrderCreateRequest) error {
	if req.User < 1 {
		return fmt.Errorf("userID must be greater than zero: %w", internal_errors.ErrBadRequest)
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item: %w", internal_errors.ErrBadRequest)
	}

	for _, item := range req.Items {
		if item.SKU < 1 {
			return fmt.Errorf("SKU must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
		if item.Count < 1 {
			return fmt.Errorf("count must be greater than zero: %w", internal_errors.ErrBadRequest)
		}
	}

	return nil
}
