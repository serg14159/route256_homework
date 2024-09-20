package client

import (
	"context"
	service "route256/cart/internal/service/cart"
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

func (c *LomsClient) OrderCreate(ctx context.Context, user int64, items []*loms.Item) (int64, error) {
	res, err := c.client.OrderCreate(ctx, &loms.OrderCreateRequest{
		User:  user,
		Items: items,
	})
	if err != nil {
		return 0, err
	}
	return res.OrderID, nil
}

var _ service.ILomsService = (*LomsClient)(nil)
