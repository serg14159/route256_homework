package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderInfo implements the gRPC OrderInfo method.
func (s *Service) OrderInfo(ctx context.Context, req *pb.OrderInfoRequest) (*pb.OrderInfoResponse, error) {

	return &pb.OrderInfoResponse{}, nil
}
