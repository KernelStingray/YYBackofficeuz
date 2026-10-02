package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TelematicsUseCase interface {
	ProcessPing(ctx context.Context, loc *Location) error
	GetNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]NearbyDriver, error)
	GetDriverLocation(ctx context.Context, driverID uuid.UUID) (*Location, error)
	CalculateETA(ctx context.Context, driverID uuid.UUID, targetLat, targetLon float64) (time.Duration, error)
	GetAllActiveLocations(ctx context.Context) ([]Location, error)
	IsAuthorizedPassenger(ctx context.Context, driverID, callerID uuid.UUID) (bool, error)
}
