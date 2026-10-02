package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParcelStatus string

const (
	ParcelStatusPending   ParcelStatus = "pending"
	ParcelStatusAccepted  ParcelStatus = "accepted"
	ParcelStatusPickedUp  ParcelStatus = "picked_up"
	ParcelStatusDelivered ParcelStatus = "delivered"
	ParcelStatusCanceled  ParcelStatus = "canceled"
)

type UserParcelStats struct {
	TotalSent      int64 `json:"total_sent"`
	TotalDelivered int64 `json:"total_delivered"` // driver role
}

type Parcel struct {
	ID                     uuid.UUID    `json:"id"`
	SenderID               uuid.UUID    `json:"sender_id"`
	DriverID               *uuid.UUID   `json:"driver_id,omitempty"`
	OriginLon              float64      `json:"origin_lon"`
	OriginLat              float64      `json:"origin_lat"`
	DestinationLon         float64      `json:"destination_lon"`
	DestinationLat         float64      `json:"destination_lat"`
	ReceiverPhoneEncrypted string       `json:"-"` // DO NOT return in standard JSON
	DecryptedReceiverPhone string       `json:"receiver_phone,omitempty"`
	ReceiverName           string       `json:"receiver_name,omitempty"`
	DropoffType            string       `json:"dropoff_type,omitempty"`
	Status                 ParcelStatus `json:"status" swaggertype:"string"`
	Price                  float64      `json:"price"`
	OriginAddress          string       `json:"from_address"`
	DestinationAddress     string       `json:"to_address"`
	Description            string       `json:"description,omitempty"`
	Size                   string       `json:"size,omitempty"`
	DeliverBy              *time.Time   `json:"deliver_by,omitempty"`
	SenderName             string       `json:"sender_name,omitempty"`
	SenderRating           float64      `json:"sender_rating,omitempty"`
	ResponsesCount         int32        `json:"responses_count"`
	CreatedAt              time.Time    `json:"created_at"`
	UpdatedAt              time.Time    `json:"updated_at"`
}

type ParcelSearchParams struct {
	Lon            float64
	Lat            float64
	RadiusMeters   float64
	DestLon        *float64
	DestLat        *float64
	CorridorMeters *float64
	PriceMin       *float64
	PriceMax       *float64
	City           string
}

type ParcelResponseStatus string

const (
	ParcelResponseStatusPending  ParcelResponseStatus = "pending"
	ParcelResponseStatusAccepted ParcelResponseStatus = "accepted"
	ParcelResponseStatusRejected ParcelResponseStatus = "rejected"
)

type ParcelResponse struct {
	ID           uuid.UUID            `json:"id"`
	ParcelID     uuid.UUID            `json:"parcel_id"`
	DriverID     uuid.UUID            `json:"driver_id"`
	Status       ParcelResponseStatus `json:"status" swaggertype:"string"`
	OfferedPrice float64              `json:"offered_price"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type AdminStats struct {
	TotalParcels     int64            `json:"total_parcels"`
	ActiveParcels    int64            `json:"active_parcels"`
	DeliveredParcels int64            `json:"delivered_parcels"`
	ChartData        []ChartDataPoint `json:"chart_data"`
}

type ChartDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
