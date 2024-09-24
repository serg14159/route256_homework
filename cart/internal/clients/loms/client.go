package client

import (
	"context"
	"fmt"
	"route256/cart/internal/models"
	"route256/loms/pkg/api/loms/v1"

	"google.golang.org/grpc"
)

type LomsClient struct {
	client loms.LomsClient
}

func NewLomsClient(conn *grpc.ClientConn) *LomsClient {
	return &LomsClient{
		client: loms.NewLomsClient(conn),
	}
}

// Function OrderCreate create order with items for user.
func (c *LomsClient) OrderCreate(ctx context.Context, user int64, items []models.CartItem) (int64, error) {
	lomsItems := make([]*loms.Item, 0, len(items))
	for _, item := range items {
		lomsItems = append(lomsItems, &loms.Item{
			Sku:   uint32(item.SKU),
			Count: uint32(item.Count),
		})
	}

	res, err := c.client.OrderCreate(ctx, &loms.OrderCreateRequest{
		User:  user,
		Items: lomsItems,
	})
	if err != nil {
		return 0, err
	}

	return res.OrderID, nil
}

// Function StocksInfo requests information about available stocks for specified SKU.
func (c *LomsClient) StocksInfo(ctx context.Context, SKU models.SKU) (int64, error) {
	req := &loms.StocksInfoRequest{
		Sku: uint32(SKU),
	}

	res, err := c.client.StocksInfo(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get stock info: %w", err)
	}

	return int64(res.Count), nil
}
