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
func (c *LomsClient) OrderCreate(ctx context.Context, user int64, items []models.CartItem) (int64, error) {
	ctx, span := otel.Tracer("LomsClient").Start(ctx, "OrderCreate")
	defer span.End()

	start := time.Now()

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
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.IncExternalRequestCounter("LomsClient.OrderCreate", status)
	metrics.ObserveExternalRequestDuration("LomsClient.OrderCreate", duration)

	if err != nil {
		return 0, err
	}

	return res.OrderID, nil
}

// StocksInfo requests information about available stocks for specified SKU.
func (c *LomsClient) StocksInfo(ctx context.Context, SKU models.SKU) (int64, error) {
	ctx, span := otel.Tracer("LomsClient").Start(ctx, "StocksInfo")
	defer span.End()

	start := time.Now()

	req := &loms.StocksInfoRequest{
		Sku: uint32(SKU),
	}

	res, err := c.client.StocksInfo(ctx, req)

	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.IncExternalRequestCounter("LomsClient.StocksInfo", status)
	metrics.ObserveExternalRequestDuration("LomsClient.StocksInfo", duration)

	if err != nil {
		return 0, fmt.Errorf("failed to get stock info: %w", err)
	}

	return int64(res.Count), nil
}
