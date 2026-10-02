package domain

import (
	"context"

	"github.com/google/uuid"
)

type TelematicsRepository interface {
	// Hot Data (Redis)
	SaveLocationCache(ctx context.Context, loc *Location) error
	GetNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]NearbyDriver, error)
	GetLocationCache(ctx context.Context, driverID uuid.UUID) (*Location, error)
	GetAllActiveLocations(ctx context.Context) ([]Location, error)

	// Redis Queue for Batching
	PushLocationToQueue(ctx context.Context, loc *Location) error
	PopLocationsFromQueue(ctx context.Context, count int) ([]Location, error)

	// Cold Data (Postgres)
	SaveLocationsBatch(ctx context.Context, locations []Location) error
	IsConfirmedPassenger(ctx context.Context, driverID, passengerID uuid.UUID) (bool, error)
}
