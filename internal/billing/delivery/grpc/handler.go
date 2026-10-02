package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	pb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/billing"
)

type BillingGrpcServer struct {
	pb.UnimplementedBillingServiceServer
	useCase domain.BillingUseCase
}

func NewBillingGrpcServer(useCase domain.BillingUseCase) *BillingGrpcServer {
	return &BillingGrpcServer{
		useCase: useCase,
	}
}

func (s *BillingGrpcServer) CheckRideLimit(ctx context.Context, req *pb.CheckRideLimitRequest) (*pb.CheckRideLimitResponse, error) {
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid driver id")
	}

	isValid, planName, err := s.useCase.CheckRideLimit(ctx, driverID)
	if err != nil {
		return &pb.CheckRideLimitResponse{
			IsValid:      false,
			ErrorMessage: err.Error(),
			PlanName:     "",
		}, nil
	}

	return &pb.CheckRideLimitResponse{
		IsValid: isValid,
		PlanName: planName,
	}, nil
}

func (s *BillingGrpcServer) DeductParcelLimit(ctx context.Context, req *pb.DeductParcelLimitRequest) (*pb.DeductParcelLimitResponse, error) {
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid driver id")
	}

	success, err := s.useCase.DeductParcelLimit(ctx, driverID)
	if err != nil || !success {
		return &pb.DeductParcelLimitResponse{
			Success: false,
		}, nil
	}

	return &pb.DeductParcelLimitResponse{
		Success: true,
	}, nil
}
