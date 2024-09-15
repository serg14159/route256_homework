package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"
)

// Function OrderCreate implements the GRPC OrderCreate method.
func (s *Service) OrderCreate(ctx context.Context, req *pb.OrderCreateRequest) (*pb.OrderCreateResponse, error) {

	return &pb.OrderCreateResponse{}, nil
}
