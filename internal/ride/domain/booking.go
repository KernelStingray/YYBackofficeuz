package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending            BookingStatus = "pending"
	BookingStatusPendingNegotiation BookingStatus = "pending_negotiation"
	BookingStatusConfirmed          BookingStatus = "confirmed"
	BookingStatusCanceled           BookingStatus = "canceled"
	BookingStatusRejected           BookingStatus = "rejected"
	BookingStatusPassengerNoShow    BookingStatus = "passenger_no_show"
)

type Booking struct {
	ID              uuid.UUID         `json:"id"`
	RideID          uuid.UUID         `json:"ride_id"`
	PassengerID     uuid.UUID         `json:"passenger_id"`
	SeatsBooked     int32             `json:"seats_booked"`
	TotalPrice      float64           `json:"total_price"`
	IsFullCarBuyout bool              `json:"is_full_car_buyout"`
	OfferedPrice    *float64          `json:"offered_price"`
	Status           BookingStatus     `json:"status" swaggertype:"string"`
	PickupPreference string            `json:"pickup_preference"`
	DropoffType      string            `json:"dropoff_type"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Ride            *Ride             `json:"ride,omitempty" swaggertype:"object"`
	Passenger       *PassengerProfile `json:"passenger,omitempty" swaggertype:"object"`
}

type PassengerProfile struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar"`
	Rating      float64   `json:"rating"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Gender      string    `json:"gender,omitempty"`
	KycStatus   string    `json:"kyc_status,omitempty"`
	CreatedAt   string    `json:"created_at,omitempty"`
}


type RejectedBooking struct {
	ID          uuid.UUID
	PassengerID uuid.UUID
	RideID      uuid.UUID
}
