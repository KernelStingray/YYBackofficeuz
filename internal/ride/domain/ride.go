package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type RideStatus string

const (
	RideStatusDraft      RideStatus = "draft"
	RideStatusActive     RideStatus = "active"
	RideStatusInProgress RideStatus = "in_progress"
	RideStatusCompleted  RideStatus = "completed"
	RideStatusCancelled  RideStatus = "cancelled"
)

var (
	ErrRideNotEditable        = errors.New("ride is not in draft status, cannot be edited")
	ErrSeatLimitExceeded      = errors.New("requested seats exceed available seats")
	ErrBookingAlreadyAccepted = errors.New("booking is already confirmed or pending")
	ErrInsufficientRideLimits = errors.New("insufficient ride limits")
	ErrInvalidRideData        = errors.New("invalid ride data provided")
	ErrDriverNotFound         = errors.New("driver not found or invalid")
	ErrRideDeparted           = errors.New("cannot book ride: ride has already departed")
	ErrRideNotActive          = errors.New("cannot book ride: ride is not active")
	ErrPastDepartureTime      = errors.New("departure time cannot be in the past")
)

type UserRideStats struct {
	TotalRidesAsDriver    int64 `json:"total_rides_as_driver"`
	TotalRidesAsPassenger int64 `json:"total_rides_as_passenger"`
}



type Ride struct {
	ID                 uuid.UUID      `json:"id"`
	DriverID           uuid.UUID      `json:"driver_id"`
	VehicleID          uuid.UUID      `json:"vehicle_id"`
	OriginLat          float64        `json:"origin_lat"`
	OriginLon          float64        `json:"origin_lon"`
	DestinationLat     float64        `json:"destination_lat"`
	DestinationLon     float64        `json:"destination_lon"`
	OriginAddress      string         `json:"origin_address"`
	DestinationAddress string         `json:"destination_address"`
	DepartureTime      time.Time      `json:"departure_time"`
	TotalSeats         int32          `json:"total_seats"`
	AvailableSeats     int32          `json:"available_seats"`
	PricePerSeat       float64        `json:"price_per_seat"`
	Status             RideStatus     `json:"status" swaggertype:"string"`
	IsNegotiable       bool           `json:"is_negotiable"`
	HasFairPriceBadge  bool           `json:"has_fair_price_badge"`
	GenderRestriction  string         `json:"gender_restriction"`
	HasAc              bool           `json:"has_ac"`
	HasLuggage         bool           `json:"has_luggage"`
	TakesParcels       bool           `json:"takes_parcels"`
	NoSmoking          bool           `json:"no_smoking"`
	NoPets             bool           `json:"no_pets"`
	DrivingSpeed       string         `json:"driving_speed"`
	MaxTwoBack         bool           `json:"max_two_back"`
	PickupPoint        string         `json:"pickup_point"`
	PickupLat          *float64       `json:"pickup_lat,omitempty"`
	PickupLon          *float64       `json:"pickup_lon,omitempty"`
	MusicTaste         *string        `json:"music_taste,omitempty"`
	AllowsCityDetour   bool           `json:"allows_city_detour"`
	CityDetourFee      float64        `json:"city_detour_fee"`
	CancelReason       string         `json:"cancel_reason,omitempty"`
	CancelledAt        *time.Time     `json:"cancelled_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	Driver             *DriverProfile `json:"driver,omitempty"`
	VehicleData        *Vehicle       `json:"car,omitempty"`
	Stops              []RideStop     `json:"stops,omitempty"`
	Bookings           []Booking      `json:"bookings,omitempty"`
}

type Vehicle struct {
	ID           uuid.UUID `json:"id"`
	Brand        string    `json:"brand"`
	Model        string    `json:"model"`
	Color            string    `json:"color"`
	LicensePlate     string    `json:"license_plate"`
	Year             int32     `json:"year"`
	ModerationStatus string    `json:"moderation_status,omitempty"`
}

type RideStop struct {
	ID         uuid.UUID `json:"id"`
	RideID     uuid.UUID `json:"ride_id"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	Address    string    `json:"address"`
	OrderIndex int32     `json:"order_index"`
}

type AdminStats struct {
	TotalRides    int64            `json:"total_rides"`
	ActiveRides   int64            `json:"active_rides"`
	PopularRoutes []PopularRoute   `json:"popular_routes"`
	ChartData     []ChartDataPoint `json:"chart_data"`
}

type ChartDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type PopularRoute struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Count       int64  `json:"count"`
}

type DriverLocation struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WatchTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WatchLocationResponse struct {
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"lon"`
	Heading   float64   `json:"heading"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DriverPublicRide struct {
	ID                 uuid.UUID `json:"id"`
	DriverID           uuid.UUID `json:"driver_id"`
	OriginAddress      string    `json:"origin_address"`
	DestinationAddress string    `json:"destination_address"`
	OriginLat          float64   `json:"origin_lat"`
	OriginLon          float64   `json:"origin_lon"`
	DestLat            float64   `json:"dest_lat"`
	DestLon            float64   `json:"dest_lon"`
	DepartureTime      time.Time `json:"departure_time"`
	PricePerSeat       float64   `json:"price_per_seat"`
	Status             string    `json:"status"`
}
