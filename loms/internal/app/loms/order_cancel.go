package loms

import (
	"context"
	"log"

	"route256/loms/internal/models"
	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderCancel implements the gRPC OrderCancel method.
func (s *Service) OrderCancel(ctx context.Context, req *pb.OrderCancelRequest) (*pb.OrderCancelResponse, error) {
	log.Printf("OrderCancel called with OrderID: %d", req.OrderID)
	defer log.Printf("OrderCancel finished with OrderID: %d", req.OrderID)
	orderCancelRequest := &models.OrderCancelRequest{
		OrderID: models.OID(req.OrderID),
	}

	err := s.LomsService.OrderCancel(ctx, orderCancelRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}

	return &pb.OrderCancelResponse{}, nil
}
