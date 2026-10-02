package domain

import (
	"time"

	"github.com/google/uuid"
)

type DriverStatistics struct {
	TotalRides     int     `json:"total_rides"`
	CompletedRides int     `json:"completed_rides"`
	CancelledRides int     `json:"cancelled_rides"`
	TotalEarnings  float64 `json:"total_earnings"`
}

type RideSearchParams struct {
	OriginLat    float64
	OriginLon    float64
	DestLat      float64
	DestLon      float64
	RadiusMeters float64
	MinSeats     int32
	Limit        int32
	Offset       int32
	Gender       string
	PriceMin     *int32
	PriceMax     *int32
	DateFrom         *time.Time
	DateTo           *time.Time
	City             string
	PickupPreference string
	PassengerTier    string
	IncludeTest      bool
}


type RouteSubscription struct {
	ID            uuid.UUID  `json:"id"`
	PassengerID   uuid.UUID  `json:"passenger_id"`
	OriginLat     float64    `json:"origin_lat"`
	OriginLon     float64    `json:"origin_lon"`
	DestLat       float64    `json:"dest_lat"`
	DestLon       float64    `json:"dest_lon"`
	DepartureDate *time.Time `json:"departure_date,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type RideRecommendations struct {
	HistoryBased []RouteStats `json:"from_history"`
	TopGlobal    []RouteStats `json:"top_routes"`
	Unvisited    []CityStats  `json:"not_visited"`
}

type RouteStats struct {
	From       string `json:"from"`
	To         string `json:"to"`
	TripsCount int    `json:"trips_count"`
}

type CityStats struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type DriverProfile struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	Rating      float64   `json:"rating"`
	RatingCount int32     `json:"rating_count"`
	Avatar      string    `json:"avatar"`
	Gender    string    `json:"gender,omitempty"`
	KycStatus string    `json:"kyc_status,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
}



