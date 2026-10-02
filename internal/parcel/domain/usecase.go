package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ParcelUseCase interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	CreateParcel(ctx context.Context, senderID uuid.UUID, origLon, origLat, destLon, destLat float64, receiverPhone, receiverName, dropoffType string, price float64, originAddress, destinationAddress, description, size string, deliverBy *time.Time) (*Parcel, error)
	SearchParcels(ctx context.Context, driverID uuid.UUID, params ParcelSearchParams) ([]Parcel, error)
	AcceptParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID) (*Parcel, error)
	GetAllParcels(ctx context.Context) ([]Parcel, error)
	GetMyParcels(ctx context.Context, userID uuid.UUID, role string) ([]Parcel, error)
	RespondToParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID, offeredPrice float64) (*ParcelResponse, error)
	GetParcelResponses(ctx context.Context, parcelID uuid.UUID, senderID uuid.UUID) ([]ParcelResponse, error)
	GetParcelByID(ctx context.Context, parcelID uuid.UUID) (*Parcel, error)
	AcceptParcelResponse(ctx context.Context, parcelID uuid.UUID, responseID uuid.UUID, senderID uuid.UUID) error
	RejectParcelResponse(ctx context.Context, parcelID uuid.UUID, responseID uuid.UUID, senderID uuid.UUID) error
	CancelParcel(ctx context.Context, parcelID uuid.UUID, senderID uuid.UUID) error
	PickupParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID) error
	DeliverParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID) error
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserParcelStats, error)
}

type EventPublisher interface {
	PublishNotification(ctx context.Context, userID uuid.UUID, title, message, eventType string) error
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}
