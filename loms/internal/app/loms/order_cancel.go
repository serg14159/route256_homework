package loms

import (
	"context"

	"route256/loms/internal/models"
	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderCancel implements the gRPC OrderCancel method.
func (s *Service) OrderCancel(ctx context.Context, req *pb.OrderCancelRequest) (*pb.OrderCancelResponse, error) {
	orderCancelRequest := &models.OrderCancelRequest{
		OrderID: models.OID(req.OrderID),
	}

	err := s.lomsService.OrderCancel(ctx, orderCancelRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}

	return &pb.OrderCancelResponse{}, nil
}
