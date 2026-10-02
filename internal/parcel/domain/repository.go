package domain

import (
	"context"

	"github.com/google/uuid"
)

type ParcelRepository interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	CreateParcel(ctx context.Context, parcel *Parcel) error
	GetParcelByID(ctx context.Context, id uuid.UUID) (*Parcel, error)
	UpdateParcelStatus(ctx context.Context, id uuid.UUID, status ParcelStatus, driverID *uuid.UUID) error
	SearchParcels(ctx context.Context, params ParcelSearchParams) ([]Parcel, error)
	GetAllParcels(ctx context.Context) ([]Parcel, error)
	GetParcelsBySenderID(ctx context.Context, senderID uuid.UUID) ([]Parcel, error)
	GetParcelsByDriverID(ctx context.Context, driverID uuid.UUID) ([]Parcel, error)
	CreateParcelResponse(ctx context.Context, response *ParcelResponse) error
	GetParcelResponsesByParcelID(ctx context.Context, parcelID uuid.UUID) ([]ParcelResponse, error)
	UpdateParcelResponseStatus(ctx context.Context, responseID uuid.UUID, status ParcelResponseStatus) error
	RejectOtherParcelResponses(ctx context.Context, parcelID uuid.UUID, acceptedResponseID uuid.UUID) error
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserParcelStats, error)
}
