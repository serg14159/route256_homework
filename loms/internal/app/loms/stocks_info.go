package loms

import (
	"context"

	pb "route256/loms/pkg/api/loms/v1"
)

// Function StocksInfo implements the gRPC StocksInfo method.
func (s *Service) StocksInfo(ctx context.Context, req *pb.StocksInfoRequest) (*pb.StocksInfoResponse, error) {

	return &pb.StocksInfoResponse{}, nil
}
