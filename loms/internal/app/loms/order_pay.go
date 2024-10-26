package loms

import (
	"context"
	"fmt"

	"route256/loms/internal/models"
	internal_errors "route256/loms/internal/pkg/errors"
	pb "route256/loms/pkg/api/loms/v1"

	"go.opentelemetry.io/otel"
)

// OrderPay implements the gRPC OrderPay method.
func (s *Service) OrderPay(ctx context.Context, req *pb.OrderPayRequest) (*pb.OrderPayResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsHandlers").Start(ctx, "OrderPay")
	defer span.End()

	orderPayRequest, err := toModelOrderPayRequest(req)
	if err != nil {
		return nil, errorToStatus(err)
	}

	err = s.LomsService.OrderPay(ctx, orderPayRequest)
	if err != nil {
		return nil, err
	}

	return &pb.OrderPayResponse{}, nil
}

// toModelOrderPayRequest convert request.
func toModelOrderPayRequest(req *pb.OrderPayRequest) (*models.OrderPayRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid input data: %w", internal_errors.ErrBadRequest)
	}

	return &models.OrderPayRequest{
		OrderID: models.OID(req.OrderID),
	}, nil
}
