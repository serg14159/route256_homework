package loms

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderInfo implements the GRPC OrderInfo method.
func (s *Service) OrderInfo(ctx context.Context, req *pb.OrderInfoRequest) (*pb.OrderInfoResponse, error) {
	orderInfoRequest, err := toModelOrderInfoRequest(req)
	if err != nil {
		return nil, errorToStatus(err)
	}

	orderInfoResponse, err := s.LomsService.OrderInfo(ctx, orderInfoRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}
	return toPbOrderInfoResponse(orderInfoResponse), nil
}

func toModelOrderInfoRequest(req *pb.OrderInfoRequest) (*models.OrderInfoRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid input data: %w", internal_errors.ErrBadRequest)
	}

	return &models.OrderInfoRequest{
		OrderID: models.OID(req.OrderID),
	}, nil
}

func toPbOrderInfoResponse(res *models.OrderInfoResponse) *pb.OrderInfoResponse {
	if res == nil {
		return &pb.OrderInfoResponse{}
	}

	items := make([]*pb.Item, len(res.Items))
	for i, item := range res.Items {
		items[i] = &pb.Item{
			Sku:   uint32(item.SKU),
			Count: uint32(item.Count),
		}
	}

	return &pb.OrderInfoResponse{
		Status: string(res.Status),
		User:   res.User,
		Items:  items,
	}
}
