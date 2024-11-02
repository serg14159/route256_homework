package loms

import (
	"context"
	"errors"
	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	pb "route256/loms/pkg/api/loms/v1"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ILomsService interface {
	OrderCreate(ctx context.Context, req *models.OrderCreateRequest) (*models.OrderCreateResponse, error)
	OrderInfo(ctx context.Context, req *models.OrderInfoRequest) (*models.OrderInfoResponse, error)
	OrderPay(ctx context.Context, req *models.OrderPayRequest) error
	OrderCancel(ctx context.Context, req *models.OrderCancelRequest) error
	OrderList(ctx context.Context) ([]models.Order, error)
	StocksInfo(ctx context.Context, req *models.StocksInfoRequest) (*models.StocksInfoResponse, error)
}

type Service struct {
	pb.UnimplementedLomsServer
	LomsService ILomsService
}

// NewService return instance of lomsService.
func NewService(lomsService ILomsService) *Service {
	return &Service{LomsService: lomsService}
}

// errorToStatus convert errors to status.
func errorToStatus(err error) error {
	var st *status.Status
	switch {
	case errors.Is(err, internal_errors.ErrNotFound):
		st = status.New(codes.NotFound, "not found")
	case errors.Is(err, internal_errors.ErrBadRequest):
		st = status.New(codes.InvalidArgument, "invalid argument")
	case errors.Is(err, internal_errors.ErrPreconditionFailed):
		st = status.New(codes.FailedPrecondition, "precondition failed")
	default:
		st = status.New(codes.Internal, "internal server error")
	}

	stDetails, errDetails := st.WithDetails(&errdetails.ErrorInfo{
		Reason:   err.Error(),
		Domain:   "loms",
		Metadata: map[string]string{"details": err.Error()},
	})
	if errDetails != nil {
		return st.Err()
	}
	return stDetails.Err()
}
