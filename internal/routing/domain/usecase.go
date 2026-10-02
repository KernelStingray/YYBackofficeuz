package domain

import (
	"context"
)

type RoutingUseCase interface {
	// CalculateRoute returns the optimal route given a list of ordered points
	CalculateRoute(ctx context.Context, points []Coordinate) (*Route, error)
	GeocodeAddress(ctx context.Context, query string) ([]GeocodedAddress, error)
	ReverseGeocode(ctx context.Context, coord Coordinate) (*GeocodedAddress, error)
}
