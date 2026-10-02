package usecase

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/telematics/domain"
)

type telematicsUseCase struct {
	repo   domain.TelematicsRepository
	worker *BatchWorker
}

func NewTelematicsUseCase(repo domain.TelematicsRepository, worker *BatchWorker) domain.TelematicsUseCase {
	return &telematicsUseCase{
		repo:   repo,
		worker: worker,
	}
}

func (u *telematicsUseCase) ProcessPing(ctx context.Context, loc *domain.Location) error {
	// Set timestamp if not provided
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}

	// 1. Hot Data: Update Redis immediately for real-time tracking
	if err := u.repo.SaveLocationCache(ctx, loc); err != nil {
		return err
	}

	// 2. Cold Data: Push to Redis Batch Queue for the background worker to pick up
	if err := u.repo.PushLocationToQueue(ctx, loc); err != nil {
		// Log the error but don't fail the request since hot data succeeded
		// slog.Error("Failed to push to batch queue", "error", err)
	}

	return nil
}

func (u *telematicsUseCase) GetNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]domain.NearbyDriver, error) {
	return u.repo.GetNearbyDrivers(ctx, lat, lon, radiusKm)
}

func (u *telematicsUseCase) GetDriverLocation(ctx context.Context, driverID uuid.UUID) (*domain.Location, error) {
	loc, err := u.repo.GetLocationCache(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, errors.New("driver location not found")
	}
	return loc, nil
}

func (u *telematicsUseCase) GetAllActiveLocations(ctx context.Context) ([]domain.Location, error) {
	return u.repo.GetAllActiveLocations(ctx)
}

func (u *telematicsUseCase) IsAuthorizedPassenger(ctx context.Context, driverID, callerID uuid.UUID) (bool, error) {
	if driverID == callerID {
		return true, nil
	}
	return u.repo.IsConfirmedPassenger(ctx, driverID, callerID)
}

// CalculateETA estimates time of arrival using a simple Haversine formula
// and an assumed average speed of 40 km/h.
// In a real production system, this would call an OSRM or Google Maps API.
func (u *telematicsUseCase) CalculateETA(ctx context.Context, driverID uuid.UUID, targetLat, targetLon float64) (time.Duration, error) {
	// 1. Get driver's current location from Redis (Hot Data)
	loc, err := u.repo.GetLocationCache(ctx, driverID)
	if err != nil {
		return 0, err
	}
	if loc == nil {
		return 0, nil // Driver not found or not online
	}

	// 2. Calculate distance using Haversine formula
	distKm := haversineDistance(loc.Latitude, loc.Longitude, targetLat, targetLon)

	// 3. Estimate time based on average speed (e.g., 40 km/h)
	avgSpeedKmh := 40.0
	hours := distKm / avgSpeedKmh

	return time.Duration(hours * float64(time.Hour)), nil
}

// haversineDistance calculates the great-circle distance between two points on a sphere
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1 = degreesToRadians(lat1)
	lat2 = degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}
