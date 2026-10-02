package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserUseCase defines the contract for user business logic.
// All delivery layers (HTTP, gRPC, AMQP) should depend on this interface,
// not on the concrete usecase implementation.
type UserUseCase interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	SearchAdminUsers(ctx context.Context, page, limit int32, search, role, status, kycStatus string) ([]*User, int64, error)
	RegisterPassenger(ctx context.Context, firstName, lastName, phone string, gender *string) (*User, error)
	UpgradeToDriver(ctx context.Context, userID uuid.UUID, car *Car, avatarURL *string) error
	ApproveDriverKYC(ctx context.Context, userID uuid.UUID) error
	RejectDriverKYC(ctx context.Context, userID uuid.UUID, reason string) error
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*User, []Car, error)
	GetCarByID(ctx context.Context, id uuid.UUID) (*Car, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string, gender, locale, avatarURL *string) (*User, error)
	CheckCarApproved(ctx context.Context, userID uuid.UUID) (bool, error)
	GetActiveCar(ctx context.Context, userID uuid.UUID) (*Car, error)
	CreateCar(ctx context.Context, userID uuid.UUID, car *Car) (*Car, error)
	UpdateCar(ctx context.Context, userID, carID uuid.UUID, car *Car) (*Car, error)
	DeleteCar(ctx context.Context, userID, carID uuid.UUID) error
	GetCars(ctx context.Context, userID uuid.UUID) ([]Car, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]User, error)
	CreateReport(ctx context.Context, reporterID, reportedID uuid.UUID, reason string) error
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}
