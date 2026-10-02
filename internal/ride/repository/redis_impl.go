package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
)

const (
	rideGeoKey       = "rides:geo"
	driverProfileKey = "driver:profile:%s"
	rideSearchKey    = "ride:search:%s"
)

type redisRideCache struct {
	client *cache.Client
}

func NewRedisRideCache(client *cache.Client) domain.RideCache {
	return &redisRideCache{client: client}
}

func (c *redisRideCache) SetRideDriver(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) error {
	key := fmt.Sprintf("ride:%s:driver", rideID.String())
	return c.client.GetRawClient().Set(ctx, key, driverID.String(), 48*time.Hour).Err()
}

func (c *redisRideCache) AddRidePassenger(ctx context.Context, rideID uuid.UUID, passengerID uuid.UUID) error {
	key := fmt.Sprintf("ride:%s:passengers", rideID.String())
	err := c.client.GetRawClient().SAdd(ctx, key, passengerID.String()).Err()
	if err == nil {
		c.client.GetRawClient().Expire(ctx, key, 48*time.Hour)
	}
	return err
}

func (c *redisRideCache) SetRideLocation(ctx context.Context, rideID uuid.UUID, lon, lat float64) error {
	return c.client.GeoAdd(ctx, rideGeoKey, lon, lat, rideID.String())
}

func (c *redisRideCache) GetNearbyRideIDs(ctx context.Context, lon, lat float64, radiusKm float64) ([]uuid.UUID, error) {
	members, err := c.client.GeoSearch(ctx, rideGeoKey, lon, lat, radiusKm)
	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, m := range members {
		id, err := uuid.Parse(m)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (c *redisRideCache) GetDriverProfile(ctx context.Context, driverID uuid.UUID) (*domain.DriverProfile, error) {
	key := fmt.Sprintf(driverProfileKey, driverID.String())
	var profile domain.DriverProfile

	err := c.client.Get(ctx, key, &profile)
	if err != nil {
		return nil, err // Returns error (e.g. key does not exist)
	}
	return &profile, nil
}

func (c *redisRideCache) InvalidateDriverProfile(ctx context.Context, driverID uuid.UUID) error {
	key := fmt.Sprintf(driverProfileKey, driverID.String())
	return c.client.Delete(ctx, key)
}

func (c *redisRideCache) SetDriverProfile(ctx context.Context, driverID uuid.UUID, profile *domain.DriverProfile) error {
	key := fmt.Sprintf(driverProfileKey, driverID.String())
	// Cache profile for 24 hours
	return c.client.Set(ctx, key, profile, 24*time.Hour)
}

func (c *redisRideCache) GetSearchResults(ctx context.Context, hash string) ([]domain.Ride, error) {
	key := fmt.Sprintf(rideSearchKey, hash)
	var rides []domain.Ride
	err := c.client.Get(ctx, key, &rides)
	if err != nil {
		return nil, err
	}
	return rides, nil
}

func (c *redisRideCache) SetSearchResults(ctx context.Context, hash string, rides []domain.Ride, ttl time.Duration) error {
	key := fmt.Sprintf(rideSearchKey, hash)
	return c.client.Set(ctx, key, rides, ttl)
}

func (c *redisRideCache) InvalidateSearchResults(ctx context.Context) error {
	iter := c.client.GetRawClient().Scan(ctx, 0, "ride:search:*", 0).Iterator()
	for iter.Next(ctx) {
		_ = c.client.GetRawClient().Del(ctx, iter.Val()).Err()
	}
	return iter.Err()
}

type telematicsLocationRaw struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *redisRideCache) GetDriverLiveLocation(ctx context.Context, driverID uuid.UUID) (*domain.DriverLocation, error) {
	key := fmt.Sprintf("telematics:drivers:data:%s", driverID.String())
	val, err := c.client.GetRawClient().Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var raw telematicsLocationRaw
	if err := json.Unmarshal([]byte(val), &raw); err != nil {
		return nil, err
	}

	return &domain.DriverLocation{
		DriverID:  raw.DriverID,
		Latitude:  raw.Latitude,
		Longitude: raw.Longitude,
		Speed:     raw.Speed,
		Heading:   raw.Heading,
		UpdatedAt: raw.Timestamp,
	}, nil
}

type watchTokenData struct {
	RideID    uuid.UUID `json:"ride_id"`
	CreatorID uuid.UUID `json:"creator_id"`
}

func (c *redisRideCache) SetWatchToken(ctx context.Context, token string, rideID, creatorID uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf("ride:watch_token:%s", token)
	data, err := json.Marshal(watchTokenData{RideID: rideID, CreatorID: creatorID})
	if err != nil {
		return err
	}
	return c.client.GetRawClient().Set(ctx, key, data, ttl).Err()
}

func (c *redisRideCache) GetWatchToken(ctx context.Context, token string) (uuid.UUID, bool, error) {
	key := fmt.Sprintf("ride:watch_token:%s", token)
	val, err := c.client.GetRawClient().Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, false, nil
	}
	var data watchTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return uuid.Nil, false, err
	}
	return data.RideID, true, nil
}

func (c *redisRideCache) DeleteWatchToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("ride:watch_token:%s", token)
	return c.client.GetRawClient().Del(ctx, key).Err()
}

