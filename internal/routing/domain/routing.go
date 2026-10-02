package domain

import (
	"context"
)

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Route struct {
	Distance float64      `json:"distance"`                      // in meters
	Duration float64      `json:"duration"`                      // in seconds
	Geometry []Coordinate `json:"geometry" swaggerignore:"true"` // the polyline decoded into an array of coords
}

type RoutingProvider interface {
	// GetRoute takes a slice of Coordinates (origin, waypoints..., destination)
	GetRoute(ctx context.Context, points []Coordinate) (*Route, error)
}

type GeocodedAddress struct {
	DisplayName string     `json:"display_name"`
	Coordinate  Coordinate `json:"coordinate"`
}

type GeocodingProvider interface {
	GeocodeAddress(ctx context.Context, query string) ([]GeocodedAddress, error)
	ReverseGeocode(ctx context.Context, coord Coordinate) (*GeocodedAddress, error)
}
