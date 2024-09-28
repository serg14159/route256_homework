package loms

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	pb "route256/loms/pkg/api/loms/v1"
)

// StocksInfo implements the gRPC StocksInfo method.
func (s *Service) StocksInfo(ctx context.Context, req *pb.StocksInfoRequest) (*pb.StocksInfoResponse, error) {
	stocksInfoRequest, err := toModelStocksInfoRequest(req)
	if err != nil {
		return nil, errorToStatus(err)
	}

	stocksInfoResponse, err := s.LomsService.StocksInfo(ctx, stocksInfoRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}

	return toPBStocksInfoResponse(stocksInfoResponse), nil
}

func toModelStocksInfoRequest(req *pb.StocksInfoRequest) (*models.StocksInfoRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid input data: %w", internal_errors.ErrBadRequest)
	}

	return &models.StocksInfoRequest{
		SKU: models.SKU(req.Sku),
	}, nil
}

func toPBStocksInfoResponse(res *models.StocksInfoResponse) *pb.StocksInfoResponse {
	if res == nil {
		return &pb.StocksInfoResponse{}
	}

	return &pb.StocksInfoResponse{
		Count: res.Count,
	}
}
