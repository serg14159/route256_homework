package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderPay implements the gRPC OrderPay method.
func (s *Service) OrderPay(ctx context.Context, req *pb.OrderPayRequest) (*pb.OrderPayResponse, error) {

	return &pb.OrderPayResponse{}, nil
}
