package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Client is a wrapper around the go-redis client.
type Client struct {
	client *redis.Client
}

// GetRawClient returns the underlying go-redis client.
func (c *Client) GetRawClient() *redis.Client {
	return c.client
}

// NewRedisClient creates a new Redis client and checks the connection.
func NewRedisClient(addr string, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set by default
		DB:       db,       // use default DB
	})

	var pingErr error
	for i := 0; i < 15; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		pingErr = rdb.Ping(ctx).Err()
		cancel()

		if pingErr == nil {
			break
		}
		slog.Warn("Redis ping failed, retrying in 2 seconds...", "attempt", i+1, "error", pingErr)
		time.Sleep(2 * time.Second)
	}

	if pingErr != nil {
		return nil, fmt.Errorf("could not connect to redis after 15 attempts: %w", pingErr)
	}

	return &Client{client: rdb}, nil
}

// Set stores a value in Redis with an expiration time. Value is JSON encoded.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}

	return nil
}

// Get retrieves a value from Redis and decodes it into the provided dest pointer.
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key does not exist")
		}
		return fmt.Errorf("failed to get value from redis: %w", err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key from Redis.
func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// GeoAdd adds a geospatial item to the specified key.
func (c *Client) GeoAdd(ctx context.Context, key string, lon, lat float64, member string) error {
	err := c.client.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      member,
		Longitude: lon,
		Latitude:  lat,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to geo add: %w", err)
	}
	return nil
}

// GeoSearch searches for items within a given radius (in km).
func (c *Client) GeoSearch(ctx context.Context, key string, lon, lat, radiusKm float64) ([]string, error) {
	res, err := c.client.GeoSearch(ctx, key, &redis.GeoSearchQuery{
		Longitude:  lon,
		Latitude:   lat,
		Radius:     radiusKm,
		RadiusUnit: "km",
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to geo search: %w", err)
	}
	return res, nil
}

// Close closes the underlying Redis client connection.
func (c *Client) Close() error {
	return c.client.Close()
}
