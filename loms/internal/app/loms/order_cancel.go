package loms

import (
	"context"

	"route256/loms/internal/models"
	pb "route256/loms/pkg/api/loms/v1"

	"go.opentelemetry.io/otel"
)

// OrderCancel implements the gRPC OrderCancel method.
func (s *Service) OrderCancel(ctx context.Context, req *pb.OrderCancelRequest) (*pb.OrderCancelResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("LomsHandlers").Start(ctx, "OrderCancel")
	defer span.End()

	orderCancelRequest := &models.OrderCancelRequest{
		OrderID: models.OID(req.OrderID),
	}

	err := s.LomsService.OrderCancel(ctx, orderCancelRequest)
	if err != nil {
		return nil, errorToStatus(err)
	}

	return &pb.OrderCancelResponse{}, nil
}
