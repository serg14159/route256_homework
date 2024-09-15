package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderCancel implements the gRPC OrderCancel method.
func (s *Service) OrderCancel(ctx context.Context, req *pb.OrderCancelRequest) (*pb.OrderCancelResponse, error) {

	return &pb.OrderCancelResponse{}, nil
}
