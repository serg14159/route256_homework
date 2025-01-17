package repository_test

import (
	"context"
	"route256/loms/internal/models"
	ordersRepository "route256/loms/internal/repository/orders"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test for CreateOrder.
func TestCreateOrder(t *testing.T) {
	orderRepo := ordersRepository.NewOrderRepository(shardManager)

	order := models.Order{
		UserID: 1,
		Status: "new",
		Items: []models.Item{
			{SKU: 1, Count: 5},
		},
	}
	ctx := context.Background()
	// Run test
	orderID, err := orderRepo.Create(ctx, order)
	require.NoError(t, err)
	require.Greater(t, orderID, int64(0))
}

// Test for GetOrderByID.
func TestGetOrderByID(t *testing.T) {
	orderRepo := ordersRepository.NewOrderRepository(shardManager)

	ctx := context.Background()
	orderID := models.OID(1)

	order, err := orderRepo.GetByID(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, order.UserID, int64(1))
	require.Equal(t, order.Items[0].SKU, models.SKU(1))
}

// Test for SetOrderStatus.
func TestSetOrderStatus(t *testing.T) {
	orderRepo := ordersRepository.NewOrderRepository(shardManager)

	ctx := context.Background()

	err := orderRepo.SetStatus(ctx, 1, "paid")
	require.NoError(t, err)

	order, err := orderRepo.GetByID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, order.Status, models.OrderStatus("paid"))
}
