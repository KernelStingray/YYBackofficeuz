package integration

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rideDomain "github.com/hextec/yolyolakay-backend/internal/ride/domain"
	billingGrpc "github.com/hextec/yolyolakay-backend/pkg/api/grpc/billing"
	userGrpc "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
)

type GRPCClient interface {
	CheckRideLimit(ctx context.Context, driverID uuid.UUID) (bool, string, error)
	CheckCarStatus(ctx context.Context, driverID uuid.UUID) (bool, error)
	GetDriverVehicle(ctx context.Context, driverID uuid.UUID) (*userGrpc.GetDriverVehicleResponse, error)
	GetDriverProfile(ctx context.Context, driverID uuid.UUID) (*rideDomain.DriverProfile, error)
	GetPassengerProfile(ctx context.Context, passengerID uuid.UUID) (*rideDomain.PassengerProfile, error)
	Close()
}

type GrpcClient struct {
	billingConn *grpc.ClientConn
	userConn    *grpc.ClientConn
	billingCli  billingGrpc.BillingServiceClient
	userCli     userGrpc.UserServiceClient
}

func NewGrpcClient(billingURL, userURL string) (*GrpcClient, error) {
	// Connect to Billing Service
	bConn, err := grpc.Dial(billingURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to billing grpc: %w", err)
	}

	// Connect to User Service
	uConn, err := grpc.Dial(userURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		bConn.Close()
		return nil, fmt.Errorf("failed to connect to user grpc: %w", err)
	}

	return &GrpcClient{
		billingConn: bConn,
		userConn:    uConn,
		billingCli:  billingGrpc.NewBillingServiceClient(bConn),
		userCli:     userGrpc.NewUserServiceClient(uConn),
	}, nil
}

func (c *GrpcClient) Close() {
	if c.billingConn != nil {
		c.billingConn.Close()
	}
	if c.userConn != nil {
		c.userConn.Close()
	}
}

func (c *GrpcClient) CheckRideLimit(ctx context.Context, driverID uuid.UUID) (bool, string, error) {
	req := &billingGrpc.CheckRideLimitRequest{
		DriverId: driverID.String(),
	}
	res, err := c.billingCli.CheckRideLimit(ctx, req)
	if err != nil {
		return false, "", err
	}
	if !res.IsValid && res.ErrorMessage != "" {
		return false, "", fmt.Errorf("billing error: %s", res.ErrorMessage)
	}
	return res.IsValid, res.PlanName, nil
}

func (c *GrpcClient) CheckCarStatus(ctx context.Context, driverID uuid.UUID) (bool, error) {
	req := &userGrpc.CheckCarStatusRequest{
		DriverId: driverID.String(),
	}
	res, err := c.userCli.CheckCarStatus(ctx, req)
	if err != nil {
		return false, err
	}
	if !res.IsApproved && res.ErrorMessage != "" {
		return false, fmt.Errorf("user car error: %s", res.ErrorMessage)
	}
	return res.IsApproved, nil
}

func (c *GrpcClient) GetDriverVehicle(ctx context.Context, driverID uuid.UUID) (*userGrpc.GetDriverVehicleResponse, error) {
	req := &userGrpc.GetDriverVehicleRequest{
		DriverId: driverID.String(),
	}
	return c.userCli.GetDriverVehicle(ctx, req)
}

func (c *GrpcClient) GetDriverProfile(ctx context.Context, driverID uuid.UUID) (*rideDomain.DriverProfile, error) {
	req := &userGrpc.GetDriverProfileRequest{
		DriverId: driverID.String(),
	}
	res, err := c.userCli.GetDriverProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	return &rideDomain.DriverProfile{
		ID:          driverID,
		Name:        res.Name,
		Avatar:      res.AvatarUrl,
		Rating:      res.Rating,
		RatingCount: res.RatingCount,
		Gender:      res.Gender,
		Phone:       res.PhoneNumber,
		KycStatus:   res.KycStatus,
		CreatedAt:   res.CreatedAt,
	}, nil
}

func (c *GrpcClient) GetPassengerProfile(ctx context.Context, passengerID uuid.UUID) (*rideDomain.PassengerProfile, error) {
	req := &userGrpc.GetPassengerProfileRequest{
		PassengerId: passengerID.String(),
	}
	res, err := c.userCli.GetPassengerProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	return &rideDomain.PassengerProfile{
		ID:          passengerID,
		Name:        res.Name,
		Avatar:      res.AvatarUrl,
		Rating:      res.Rating,
		PhoneNumber: res.PhoneNumber,
		KycStatus:   res.KycStatus,
		CreatedAt:   res.CreatedAt,
	}, nil
}

