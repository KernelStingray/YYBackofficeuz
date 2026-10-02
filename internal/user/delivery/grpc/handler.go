package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hextec/yolyolakay-backend/internal/user/domain"
	pb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
)


type UserGrpcServer struct {
	pb.UnimplementedUserServiceServer
	useCase domain.UserUseCase
}

func NewUserGrpcServer(useCase domain.UserUseCase) *UserGrpcServer {
	return &UserGrpcServer{
		useCase: useCase,
	}
}

func (s *UserGrpcServer) CheckCarStatus(ctx context.Context, req *pb.CheckCarStatusRequest) (*pb.CheckCarStatusResponse, error) {
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid driver id")
	}

	isApproved, err := s.useCase.CheckCarApproved(ctx, driverID)
	if err != nil {
		return &pb.CheckCarStatusResponse{
			IsApproved:   false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.CheckCarStatusResponse{
		IsApproved: isApproved,
	}, nil
}

func (s *UserGrpcServer) GetUserLocale(ctx context.Context, req *pb.GetUserLocaleRequest) (*pb.GetUserLocaleResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	user, _, err := s.useCase.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	locale := user.Locale
	if locale == "" {
		locale = "uz"
	}

	return &pb.GetUserLocaleResponse{
		Locale: locale,
	}, nil
}

func (s *UserGrpcServer) GetDriverVehicle(ctx context.Context, req *pb.GetDriverVehicleRequest) (*pb.GetDriverVehicleResponse, error) {
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid driver id")
	}

	car, err := s.useCase.GetActiveCar(ctx, driverID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "active car not found")
	}

	return &pb.GetDriverVehicleResponse{
		Id:               car.ID.String(),
		Brand:            car.Brand,
		Model:            car.Model,
		Color:            car.Color,
		Year:             car.Year,
		LicensePlate:     car.GovNumber,
		ModerationStatus: car.ModerationStatus,
	}, nil
}

func (s *UserGrpcServer) GetDriverProfile(ctx context.Context, req *pb.GetDriverProfileRequest) (*pb.GetDriverProfileResponse, error) {
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid driver id")
	}

	user, _, err := s.useCase.GetUserProfile(ctx, driverID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	var gender string
	if user.Gender != nil {
		gender = *user.Gender
	}

	var avatarURL string
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.GetDriverProfileResponse{
		Id:          user.ID.String(),
		Name:        user.FirstName + " " + user.LastName,
		AvatarUrl:   avatarURL,
		Rating:      user.Rating,
		Gender:      gender,
		PhoneNumber: user.PhoneNumber,
		KycStatus:   string(user.KycStatus),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		RatingCount: user.RatingCount,
	}, nil
}

func (s *UserGrpcServer) GetPassengerProfile(ctx context.Context, req *pb.GetPassengerProfileRequest) (*pb.GetPassengerProfileResponse, error) {
	passengerID, err := uuid.Parse(req.GetPassengerId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid passenger id")
	}

	user, _, err := s.useCase.GetUserProfile(ctx, passengerID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	var avatarURL string
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.GetPassengerProfileResponse{
		Id:          user.ID.String(),
		Name:        user.FirstName + " " + user.LastName,
		AvatarUrl:   avatarURL,
		Rating:      user.Rating,
		PhoneNumber: user.PhoneNumber,
		KycStatus:   string(user.KycStatus),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}, nil
}



