package loms

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderCreate implements the GRPC OrderCreate method.
func (s *Service) OrderCreate(ctx context.Context, req *pb.OrderCreateRequest) (*pb.OrderCreateResponse, error) {
	orderCreateRequest, err := toModelOrderCreateRequest(req)
	if err != nil {
		return nil, errorToStatus(err)
	}

	orderCreateResponse, err := s.LomsService.OrderCreate(ctx, orderCreateRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}

	return toPbOrderCreateResponse(orderCreateResponse), nil
}

func toModelOrderCreateRequest(req *pb.OrderCreateRequest) (*models.OrderCreateRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid input data: %w", internal_errors.ErrBadRequest)
	}

	items := make([]models.Item, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.Item{
			SKU:   models.SKU(item.Sku),
			Count: uint16(item.Count),
		}
	}

	return &models.OrderCreateRequest{
		User:  models.UID(req.User),
		Items: items,
	}, nil
}

func toPbOrderCreateResponse(res *models.OrderCreateResponse) *pb.OrderCreateResponse {
	if res == nil {
		return &pb.OrderCreateResponse{}
	}

	return &pb.OrderCreateResponse{
		OrderID: res.OrderID,
	}
}
