package client

import (
	"context"
	"fmt"
	"route256/cart/internal/models"
	"route256/cart/internal/pkg/metrics"
	"route256/loms/pkg/api/loms/v1"
	"time"

	"go.opentelemetry.io/otel"
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

// OrderCreate create order with items for user.
func (c *LomsClient) OrderCreate(ctx context.Context, user int64, items []models.CartItem) (orderID int64, err error) {
	// Tracer
	ctx, span := otel.Tracer("LomsClient").Start(ctx, "OrderCreate")
	defer span.End()

	// Start time for metrics
	start := time.Now()
	defer metrics.LogExternalRequest("LomsClient.OrderCreate", start, &err)

	lomsItems := make([]*loms.Item, 0, len(items))
	for _, item := range items {
		lomsItems = append(lomsItems, &loms.Item{
			Sku:   uint32(item.SKU),
			Count: uint32(item.Count),
		})
	}

	// Call client
	var res *loms.OrderCreateResponse
	res, err = c.client.OrderCreate(ctx, &loms.OrderCreateRequest{
		User:  user,
		Items: lomsItems,
	})

	if err != nil {
		return 0, err
	}

	return res.OrderID, nil
}

// StocksInfo requests information about available stocks for specified SKU.
func (c *LomsClient) StocksInfo(ctx context.Context, SKU models.SKU) (count int64, err error) {
	// Tracer
	ctx, span := otel.Tracer("LomsClient").Start(ctx, "StocksInfo")
	defer span.End()

	// Start time for metrics
	start := time.Now()
	defer metrics.LogExternalRequest("LomsClient.StocksInfo", start, &err)

	req := &loms.StocksInfoRequest{
		Sku: uint32(SKU),
	}

	// Call client
	var res *loms.StocksInfoResponse
	res, err = c.client.StocksInfo(ctx, req)

	if err != nil {
		err = fmt.Errorf("failed to get stock info: %w", err)
		return 0, err
	}

	return int64(res.Count), nil
}
