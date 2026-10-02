package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/telematics/domain"
	redis "github.com/redis/go-redis/v9"
)

const (
	telematicsGeoKey     = "telematics:drivers:geo"
	telematicsDataKey    = "telematics:drivers:data:"
	telematicsBatchQueue = "telematics:batch:queue"
	expireTime           = 5 * time.Minute
)

type redisTelematicsRepository struct {
	client *redis.Client
}

func NewRedisTelematicsRepository(client *redis.Client) *redisTelematicsRepository {
	return &redisTelematicsRepository{
		client: client,
	}
}

func (r *redisTelematicsRepository) SaveLocationCache(ctx context.Context, loc *domain.Location) error {
	// 1. Save geospatial data (for GEORADIUS)
	err := r.client.GeoAdd(ctx, telematicsGeoKey, &redis.GeoLocation{
		Name:      loc.DriverID.String(),
		Longitude: loc.Longitude,
		Latitude:  loc.Latitude,
	}).Err()
	if err != nil {
		return err
	}

	// 2. Save detailed data (speed, heading, timestamp)
	data, err := json.Marshal(loc)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%s", telematicsDataKey, loc.DriverID.String())
	return r.client.Set(ctx, key, data, expireTime).Err()
}

func (r *redisTelematicsRepository) GetNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]domain.NearbyDriver, error) {
	// If radiusKm > 100, caller passed meters (e.g. 1000m), so convert to km
	actualRadiusKm := radiusKm
	if actualRadiusKm > 100 {
		actualRadiusKm = actualRadiusKm / 1000.0
	}
	if actualRadiusKm <= 0 {
		actualRadiusKm = 10.0 // default 10 km
	}

	res, err := r.client.GeoRadius(ctx, telematicsGeoKey, lon, lat, &redis.GeoRadiusQuery{
		Radius:      actualRadiusKm,
		Unit:        "km",
		WithCoord:   true,
		WithDist:    true,
		WithGeoHash: false,
		Count:       100,
		Sort:        "ASC",
	}).Result()

	if err != nil {
		return nil, err
	}

	drivers := make([]domain.NearbyDriver, 0, len(res))
	for _, l := range res {
		id, err := uuid.Parse(l.Name)
		if err != nil {
			continue
		}

		nd := domain.NearbyDriver{
			DriverID:  id,
			Distance:  l.Dist * 1000, // in meters
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		}

		// Enrich with Heading, Speed and Timestamp from hot cache
		if loc, err := r.GetLocationCache(ctx, id); err == nil && loc != nil {
			nd.Heading = loc.Heading
			nd.Speed = loc.Speed
			nd.UpdatedAt = loc.Timestamp
		}

		drivers = append(drivers, nd)
	}

	return drivers, nil
}

func (r *redisTelematicsRepository) GetLocationCache(ctx context.Context, driverID uuid.UUID) (*domain.Location, error) {
	key := fmt.Sprintf("%s%s", telematicsDataKey, driverID.String())
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // not found
		}
		return nil, err
	}

	var loc domain.Location
	if err := json.Unmarshal([]byte(val), &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *redisTelematicsRepository) GetAllActiveLocations(ctx context.Context) ([]domain.Location, error) {
	driverIDs, err := r.client.ZRange(ctx, telematicsGeoKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var locations []domain.Location
	for _, idStr := range driverIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		loc, err := r.GetLocationCache(ctx, id)
		if err == nil && loc != nil {
			locations = append(locations, *loc)
		}
	}
	return locations, nil
}

// Ensure interface compatibility
// postgres batching logic will go to a combined repository or we separate it.
// Let's create a wrapper that combines both in usecase or a master repo.

func (r *redisTelematicsRepository) PushLocationToQueue(ctx context.Context, loc *domain.Location) error {
	data, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	return r.client.RPush(ctx, telematicsBatchQueue, data).Err()
}

func (r *redisTelematicsRepository) PopLocationsFromQueue(ctx context.Context, count int) ([]domain.Location, error) {
	pipe := r.client.Pipeline()
	// LPop with count requires Redis 6.2+. go-redis LPopCount is suitable
	cmds := make([]*redis.StringCmd, count)
	for i := 0; i < count; i++ {
		cmds[i] = pipe.LPop(ctx, telematicsBatchQueue)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var locations []domain.Location
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			break // queue is empty
		}
		if err != nil {
			continue // some other error, skip
		}
		var loc domain.Location
		if err := json.Unmarshal([]byte(val), &loc); err == nil {
			locations = append(locations, loc)
		}
	}

	return locations, nil
}
