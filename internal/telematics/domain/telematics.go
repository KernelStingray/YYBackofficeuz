package domain

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

type NearbyDriver struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Distance  float64   `json:"distance"` // in meters
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
