package http

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/ride/domain"
)

type PublishRideRequest struct {
	VehicleID          uuid.UUID         `json:"vehicle_id" binding:"required"`
	OriginLat          float64           `json:"origin_lat" binding:"required"`
	OriginLon          float64           `json:"origin_lon" binding:"required"`
	DestinationLat     float64           `json:"destination_lat" binding:"required"`
	DestinationLon     float64           `json:"destination_lon" binding:"required"`
	OriginAddress      string            `json:"origin_address" binding:"required,max=255"`
	DestinationAddress string            `json:"destination_address" binding:"required,max=255"`
	DepartureTime      string            `json:"departure_time" binding:"required,max=50"` // RFC3339 format
	TotalSeats         int32             `json:"total_seats" binding:"required,gt=0,lte=8"`
	PricePerSeat       float64           `json:"price_per_seat" binding:"required,gt=0,lte=10000000"`
	IsNegotiable       bool              `json:"is_negotiable"`
	GenderRestriction  string            `json:"gender_restriction" binding:"max=20"`
	HasAc              bool              `json:"has_ac"`
	HasLuggage         bool              `json:"has_luggage"`
	TakesParcels       bool              `json:"takes_parcels"`
	NoSmoking          bool              `json:"no_smoking"`
	NoPets             bool              `json:"no_pets"`
	DrivingSpeed       string            `json:"driving_speed" binding:"max=20"`
	MaxTwoBack         bool              `json:"max_two_back"`
	PickupPoint        string            `json:"pickup_point" binding:"max=255"`
	PickupLat          *float64          `json:"pickup_lat"`
	PickupLon          *float64          `json:"pickup_lon"`
	PickupPointLat     *float64          `json:"pickup_point_lat"`
	PickupPointLon     *float64          `json:"pickup_point_lon"`
	MusicTaste         string            `json:"music_taste" binding:"max=100"`
	AllowsCityDetour   bool              `json:"allows_city_detour"`
	CityDetourFee      float64           `json:"city_detour_fee"`
	Stops              []RideStopRequest `json:"stops"`
}

type RideStopRequest struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lon     float64 `json:"lon" binding:"required"`
	Address string  `json:"address" binding:"required,max=255"`
}

type SearchRidesRequest struct {
	OriginLat         float64    `form:"origin_lat" binding:"required"`
	OriginLon         float64    `form:"origin_lon" binding:"required"`
	DestLat           float64    `form:"dest_lat" binding:"required"`
	DestLon           float64    `form:"dest_lon" binding:"required"`
	RadiusKm          float64    `form:"radius_km"` // Optional, defaults to 10.0
	MinSeats          int32      `form:"min_seats"`
	Gender            string     `form:"gender"`
	GenderRestriction string     `form:"gender_restriction"` // alias if frontend sends this
	PriceMin          *int32     `form:"price_min"`
	PriceMax          *int32     `form:"price_max"`
	DateFrom          *string    `form:"date_from"`
	DateTo            *string    `form:"date_to"`
	City              string     `form:"city"`
	PickupPreference  string     `form:"pickup_preference"`
	PassengerTier     string     `form:"passenger_tier"`
	IncludeTest       bool       `form:"include_test"`
}

type BookRideRequest struct {
	Seats            int32    `json:"seats"`
	IsFullCarBuyout  bool     `json:"is_full_car_buyout"`
	OfferedPrice     *float64 `json:"offered_price,omitempty"`
	PickupPreference string   `json:"pickup_preference" binding:"max=250"`
	DropoffType      string   `json:"dropoff_type" binding:"max=20"`
}

type BookingResponse struct {
	ID               uuid.UUID                `json:"id"`
	RideID           uuid.UUID                `json:"ride_id"`
	PassengerID      uuid.UUID                `json:"passenger_id"`
	SeatsBooked      int32                    `json:"seats_booked"`
	TotalPrice       float64                  `json:"total_price"`
	IsFullCarBuyout  bool                     `json:"is_full_car_buyout"`
	OfferedPrice     *float64                 `json:"offered_price"`
	Status           string                   `json:"status"`
	PickupPreference string                   `json:"pickup_preference"`
	DropoffType      string                   `json:"dropoff_type,omitempty"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
	Ride             *RideSearchResponse      `json:"ride,omitempty"`
	Passenger        *domain.PassengerProfile `json:"passenger,omitempty"`
}

type DriverStatisticsResponse struct {
	TotalRides     int     `json:"total_rides"`
	CompletedRides int     `json:"completed_rides"`
	CancelledRides int     `json:"cancelled_rides"`
	TotalEarnings  float64 `json:"total_earnings"`
}

type RideResponse struct {
	ID                 uuid.UUID                   `json:"id"`
	DriverID           uuid.UUID                   `json:"driver_id"`
	VehicleID          uuid.UUID                   `json:"vehicle_id"`
	OriginLat          float64                     `json:"origin_lat"`
	OriginLon          float64                     `json:"origin_lon"`
	DestinationLat     float64                     `json:"destination_lat"`
	DestinationLon     float64                     `json:"destination_lon"`
	OriginAddress      string                      `json:"origin_address"`
	DestinationAddress string                      `json:"destination_address"`
	DepartureTime      string                      `json:"departure_time"`
	TotalSeats         int32                       `json:"total_seats"`
	AvailableSeats     int32                       `json:"available_seats"`
	PricePerSeat       float64                     `json:"price_per_seat"`
	Status             string                      `json:"status"`
	IsNegotiable       bool                        `json:"is_negotiable"`
	HasFairPriceBadge  bool                        `json:"has_fair_price_badge"`
	GenderRestriction  string                      `json:"gender_restriction"`
	HasAc              bool                        `json:"has_ac"`
	HasLuggage         bool                        `json:"has_luggage"`
	TakesParcels       bool                        `json:"takes_parcels"`
	NoSmoking          bool                        `json:"no_smoking"`
	NoPets             bool                        `json:"no_pets"`
	DrivingSpeed       string                      `json:"driving_speed"`
	MaxTwoBack         bool                        `json:"max_two_back"`
	PickupPoint        string                      `json:"pickup_point"`
	PickupLat          *float64                    `json:"pickup_lat,omitempty"`
	PickupLon          *float64                    `json:"pickup_lon,omitempty"`
	MusicTaste         string                      `json:"music_taste,omitempty"`
	AllowsCityDetour   bool                        `json:"allows_city_detour"`
	CityDetourFee      float64                     `json:"city_detour_fee"`
	CreatedAt          string                      `json:"created_at"`
	UpdatedAt          string                      `json:"updated_at"`
	Driver             *DriverProfile              `json:"driver,omitempty"`
	Stops              []RideStop                  `json:"stops,omitempty"`
	Car                *Vehicle                    `json:"car,omitempty"`
	Passengers         []domain.PassengerProfile   `json:"passengers"`
}

type DriverProfile struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Rating      float64   `json:"rating"`
	RatingCount int32     `json:"rating_count"`
	Avatar      string    `json:"avatar"`
	Gender      string    `json:"gender,omitempty"`
	KycStatus   string    `json:"kyc_status,omitempty"`
	CreatedAt   string    `json:"created_at,omitempty"`
}


type RideStop struct {
	ID         uuid.UUID `json:"id"`
	RideID     uuid.UUID `json:"ride_id"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	Address    string    `json:"address"`
	OrderIndex int32     `json:"order_index"`
}

type Vehicle struct {
	ID               uuid.UUID `json:"id"`
	Brand            string    `json:"brand"`
	Model            string    `json:"model"`
	Color            string    `json:"color"`
	LicensePlate     string    `json:"license_plate"`
	Year             int32     `json:"year"`
	ModerationStatus string    `json:"moderation_status,omitempty"`
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

type RideSearchResponse struct {
	ID                 uuid.UUID      `json:"id"`
	OriginLat          float64        `json:"origin_lat"`
	OriginLon          float64        `json:"origin_lon"`
	DestinationLat     float64        `json:"destination_lat"`
	DestinationLon     float64        `json:"destination_lon"`
	OriginAddress      string         `json:"origin_address"`
	DestinationAddress string         `json:"destination_address"`
	DepartureTime      string         `json:"departure_time"`
	TotalSeats         int32          `json:"total_seats"`
	AvailableSeats     int32          `json:"available_seats"`
	PricePerSeat       float64        `json:"price_per_seat"`
	Status             string         `json:"status"`
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
	MusicTaste         string         `json:"music_taste,omitempty"`
	AllowsCityDetour   bool           `json:"allows_city_detour"`
	CityDetourFee      float64        `json:"city_detour_fee"`
	Stops              []RideStop     `json:"stops"`
	Driver             *DriverProfile `json:"driver"`
	Car                *Vehicle       `json:"car,omitempty"`
}

type RecommendationsResponse struct {
	ByYourRoutes  []RouteStats `json:"by_your_routes"`
	Popular       []RouteStats `json:"popular"`
	NewDirections []CityStats  `json:"new_directions"`
}

type CounterBookingRequest struct {
	NewPrice float64 `json:"new_price" binding:"required,gt=0"`
}

type SearchRidesResponse struct {
	Rides []RideSearchResponse `json:"rides"`
}

type UserRideStatsResponse struct {
	TotalRidesAsDriver    int32   `json:"total_rides_as_driver"`
	TotalRidesAsPassenger int32   `json:"total_rides_as_passenger"`
	RatingAsDriver        float64 `json:"rating_as_driver"`
	RatingAsPassenger     float64 `json:"rating_as_passenger"`
}

type AdminStatsResponse struct {
	TotalRides     int64   `json:"total_rides"`
	ActiveRides    int64   `json:"active_rides"`
	CompletedRides int64   `json:"completed_rides"`
	CancelledRides int64   `json:"cancelled_rides"`
	TotalBookings  int64   `json:"total_bookings"`
	TotalRevenue   float64 `json:"total_revenue"`
}

type RouteSubscriptionResponse struct {
	ID            uuid.UUID  `json:"id"`
	PassengerID   uuid.UUID  `json:"passenger_id"`
	OriginLat     float64    `json:"origin_lat"`
	OriginLon     float64    `json:"origin_lon"`
	DestLat       float64    `json:"dest_lat"`
	DestLon       float64    `json:"dest_lon"`
	DepartureDate *time.Time `json:"departure_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}


