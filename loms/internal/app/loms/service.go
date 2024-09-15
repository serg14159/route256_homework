package loms

import (
	pb "route256/loms/pkg/api/loms/v1"
)

var _ pb.LomsServer = (*Service)(nil)

// type ILomsService interface {
// 	OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error)
// 	OrderInfo(ctx context.Context, req *models.OrderInfoRequest) (*models.OrderInfoResponse, error)
// 	OrderPay(ctx context.Context, req *models.OrderPayRequest) (*models.OrderPayResponse, error)
// 	OrderCancel(ctx context.Context, req *models.OrderCancelRequest) (*models.OrderCancelResponse, error)
// 	StocksInfo(ctx context.Context, req *models.StocksInfoRequest) (*models.StocksInfoResponse, error)
// }

type Service struct {
	pb.UnimplementedLomsServer
	//lomsService ILomsService
}

// func NewService(lomsService ILomsService) *Service {
// 	return &Service{lomsService: lomsService}
// }

func NewService() *Service {
	return &Service{}
}
