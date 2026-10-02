package domain

import (
	"context"

	"github.com/google/uuid"
)

type RideUseCase interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	PublishRide(ctx context.Context, ride *Ride) error
	ActivateRide(ctx context.Context, driverID, rideID uuid.UUID) error
	StartRide(ctx context.Context, driverID, rideID uuid.UUID) error
	FindRides(ctx context.Context, params RideSearchParams) ([]Ride, error)
	BookRide(ctx context.Context, rideID, passengerID uuid.UUID, seats int32, isFullCarBuyout bool, offeredPrice *float64, pickupPreference, dropoffType string) (*Booking, error)
	AcceptBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error
	RejectBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error
	CounterBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID, newPrice float64) error
	TimeoutPendingNegotiations(ctx context.Context) error
	CancelRide(ctx context.Context, driverID, rideID uuid.UUID, reason string) error
	CompleteRide(ctx context.Context, driverID, rideID uuid.UUID) error
	ReportPassengerNoShow(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error
	CancelBookingByPassenger(ctx context.Context, passengerID, bookingID uuid.UUID, reason string) error
	HandleLimitExceeded(ctx context.Context, rideID uuid.UUID, bookingID *uuid.UUID, userID uuid.UUID) error
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserRideStats, error)
	GetUserActiveRide(ctx context.Context, userID uuid.UUID) (*Ride, error)
	GetMyRidesAsDriver(ctx context.Context, driverID uuid.UUID) ([]Ride, error)
	GetMyBookingsAsPassenger(ctx context.Context, passengerID uuid.UUID) ([]Booking, error)
	GetDriverStatistics(ctx context.Context, driverID uuid.UUID) (*DriverStatistics, error)
	GetRideByID(ctx context.Context, rideID uuid.UUID) (*Ride, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	CreateRouteSubscription(ctx context.Context, sub *RouteSubscription) error
	GetRouteSubscriptions(ctx context.Context, userID uuid.UUID) ([]RouteSubscription, error)
	DeleteRouteSubscription(ctx context.Context, subID, userID uuid.UUID) error
	GetRideRecommendations(ctx context.Context, userID uuid.UUID) (*RideRecommendations, error)
	GetAdminActiveRides(ctx context.Context) ([]Ride, error)
	GetRideDriverLocation(ctx context.Context, rideID, callerID uuid.UUID) (*DriverLocation, error)
	CreateWatchToken(ctx context.Context, rideID, callerID uuid.UUID) (*WatchTokenResponse, error)
	RevokeWatchToken(ctx context.Context, rideID, callerID uuid.UUID, token string) error
	GetRideDriverLocationByWatchToken(ctx context.Context, rideID uuid.UUID, token string) (*WatchLocationResponse, error)
	GetDriverPublicRides(ctx context.Context, driverID uuid.UUID, limit int) ([]DriverPublicRide, error)
}
