package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/routing/domain"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
)

type routingUseCase struct {
	provider          domain.RoutingProvider
	geocodingProvider domain.GeocodingProvider
	cache             *cache.Client
}

func NewRoutingUseCase(provider domain.RoutingProvider, geocodingProvider domain.GeocodingProvider, cacheClient *cache.Client) domain.RoutingUseCase {
	return &routingUseCase{
		provider:          provider,
		geocodingProvider: geocodingProvider,
		cache:             cacheClient,
	}
}

func (u *routingUseCase) CalculateRoute(ctx context.Context, points []domain.Coordinate) (*domain.Route, error) {
	if len(points) < 2 {
		return nil, errors.New("at least 2 points required")
	}

	// Generate cache key
	cacheKey := "route"
	for _, p := range points {
		cacheKey += fmt.Sprintf(":%f:%f", p.Latitude, p.Longitude)
	}

	var cachedRoute domain.Route
	// Try to get from cache
	if u.cache != nil {
		err := u.cache.Get(ctx, cacheKey, &cachedRoute)
		if err == nil {
			slog.Info("[RoutingUseCase] Route retrieved from cache", "cacheKey", cacheKey)
			return &cachedRoute, nil
		}
	}

	start := time.Now()
	route, err := u.provider.GetRoute(ctx, points)
	if err != nil {
		slog.Error("[RoutingUseCase] Failed to get route", "error", err)
		return nil, err
	}

	slog.Info("[RoutingUseCase] Route calculated successfully", "duration", time.Since(start), "distance_m", route.Distance, "duration_s", route.Duration)

	// Save to cache for 24 hours
	if u.cache != nil {
		err = u.cache.Set(ctx, cacheKey, route, 24*time.Hour)
		if err != nil {
			slog.Error("[RoutingUseCase] Failed to cache route", "error", err)
		} else {
			slog.Info("[RoutingUseCase] Route cached successfully", "cacheKey", cacheKey)
		}
	}

	return route, nil
}

func (u *routingUseCase) GeocodeAddress(ctx context.Context, query string) ([]domain.GeocodedAddress, error) {
	if u.geocodingProvider == nil {
		return nil, errors.New("geocoding provider is not configured")
	}
	return u.geocodingProvider.GeocodeAddress(ctx, query)
}

func (u *routingUseCase) ReverseGeocode(ctx context.Context, coord domain.Coordinate) (*domain.GeocodedAddress, error) {
	if u.geocodingProvider == nil {
		return nil, errors.New("geocoding provider is not configured")
	}
	return u.geocodingProvider.ReverseGeocode(ctx, coord)
}
