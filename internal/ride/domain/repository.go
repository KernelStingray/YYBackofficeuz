package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RideRepository interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	CreateRide(ctx context.Context, ride *Ride) error
	UpdateRideRoutePolygon(ctx context.Context, rideID uuid.UUID, wkt string, radius float64) error
	SearchRides(ctx context.Context, params RideSearchParams) ([]Ride, error)
	CreateBookingTx(ctx context.Context, rideID, passengerID uuid.UUID, seats int32, isFullCarBuyout bool, offeredPrice *float64, pickupPreference, dropoffType string) (*Booking, error)
	AcceptBookingTx(ctx context.Context, rideID, bookingID uuid.UUID) (*Booking, error)
	RejectBooking(ctx context.Context, bookingID uuid.UUID) error
	CancelBookingTx(ctx context.Context, passengerID, bookingID uuid.UUID) (*Booking, error)
	GetRideByID(ctx context.Context, rideID uuid.UUID) (*Ride, error)
	GetRidesByIDs(ctx context.Context, ids []uuid.UUID) ([]Ride, error)
	DeleteRidesByUserID(ctx context.Context, userID uuid.UUID) error
	UpdateRideStatus(ctx context.Context, rideID uuid.UUID, status RideStatus) error
	CountPassengerNoShows(ctx context.Context, passengerID uuid.UUID, since time.Time) (int, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	GetBookingsByRideID(ctx context.Context, rideID uuid.UUID) ([]Booking, error)
	UpdateBookingStatus(ctx context.Context, id uuid.UUID, status string) error
	ReportPassengerNoShowTx(ctx context.Context, rideID, bookingID uuid.UUID) error
	UpdateBookingPriceAndStatus(ctx context.Context, id uuid.UUID, price float64, status string) error
	RejectExpiredNegotiations(ctx context.Context) ([]RejectedBooking, error)
	GetRidesByDriverID(ctx context.Context, driverID uuid.UUID) ([]Ride, error)
	GetBookingsByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]Booking, error)
	GetDriverStatisticsAggregated(ctx context.Context, driverID uuid.UUID) (*DriverStatistics, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserRideStats, error)
	GetUserMostVisitedDestinations(ctx context.Context, userID uuid.UUID, limit int32) ([]RouteStats, error)
	GetTopGlobalDestinations(ctx context.Context, limit int32) ([]RouteStats, error)
	GetActiveRidesToDestinations(ctx context.Context, destinations []string, limit int32) ([]Ride, error)
	GetAdminActiveRides(ctx context.Context) ([]Ride, error)
	GetPublicRidesByDriverID(ctx context.Context, driverID uuid.UUID, limit int) ([]DriverPublicRide, error)
}

type RouteSubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub *RouteSubscription) error
	GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]RouteSubscription, error)
	DeleteSubscription(ctx context.Context, id, userID uuid.UUID) error
	FindMatchingSubscriptions(ctx context.Context, originLat, originLon, destLat, destLon float64, departureTime time.Time, radiusKm float64) ([]RouteSubscription, error)
}
