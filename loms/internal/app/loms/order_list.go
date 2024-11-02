package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"

	"go.opentelemetry.io/otel"
)

// OrderList implements the GRPC OrderList method.
func (s *Service) OrderList(ctx context.Context, req *pb.OrderListRequest) (*pb.OrderListResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsHandlers").Start(ctx, "OrderList")
	defer span.End()

	orders, err := s.LomsService.OrderList(ctx)
	if err != nil {
		return nil, errorToStatus(err)
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		items := make([]*pb.Item, len(order.Items))
		for j, item := range order.Items {
			items[j] = &pb.Item{
				Sku:   uint32(item.SKU),
				Count: uint32(item.Count),
			}
		}

		pbOrders[i] = &pb.Order{
			OrderID: order.OrderID,
			Status:  string(order.Status),
			User:    order.UserID,
			Items:   items,
		}
	}

	return &pb.OrderListResponse{
		Orders: pbOrders,
	}, nil
}
