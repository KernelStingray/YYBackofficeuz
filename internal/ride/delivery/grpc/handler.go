package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	ride "github.com/hextec/yolyolakay-backend/pkg/api/grpc/ride"
)

type Handler struct {
	ride.UnimplementedRideServiceServer
	useCase domain.RideUseCase
}

func NewHandler(useCase domain.RideUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) GetRideByID(ctx context.Context, req *ride.GetRideByIDRequest) (*ride.GetRideByIDResponse, error) {
	rideID, err := uuid.Parse(req.GetRideId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ride id")
	}

	r, err := h.useCase.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ride not found")
	}

	return &ride.GetRideByIDResponse{
		Id:       r.ID.String(),
		DriverId: r.DriverID.String(),
		Status:   string(r.Status),
	}, nil
}
