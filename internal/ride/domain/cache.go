package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishNotification(ctx context.Context, userID uuid.UUID, title, message, eventType string) error
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

type RideCache interface {
	SetRideLocation(ctx context.Context, rideID uuid.UUID, lon, lat float64) error
	GetNearbyRideIDs(ctx context.Context, lon, lat float64, radiusKm float64) ([]uuid.UUID, error)
	GetDriverProfile(ctx context.Context, driverID uuid.UUID) (*DriverProfile, error)
	SetDriverProfile(ctx context.Context, driverID uuid.UUID, profile *DriverProfile) error
	InvalidateDriverProfile(ctx context.Context, driverID uuid.UUID) error
	SetRideDriver(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) error
	AddRidePassenger(ctx context.Context, rideID uuid.UUID, passengerID uuid.UUID) error
	GetSearchResults(ctx context.Context, hash string) ([]Ride, error)
	SetSearchResults(ctx context.Context, hash string, rides []Ride, ttl time.Duration) error
	InvalidateSearchResults(ctx context.Context) error
	GetDriverLiveLocation(ctx context.Context, driverID uuid.UUID) (*DriverLocation, error)
	SetWatchToken(ctx context.Context, token string, rideID, creatorID uuid.UUID, ttl time.Duration) error
	GetWatchToken(ctx context.Context, token string) (rideID uuid.UUID, exists bool, err error)
	DeleteWatchToken(ctx context.Context, token string) error
}
