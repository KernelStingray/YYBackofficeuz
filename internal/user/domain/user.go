package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRole string

const (
	RolePassenger UserRole = "passenger"
	RoleDriver    UserRole = "driver"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
)

type User struct {
	ID             uuid.UUID  `json:"id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Gender         *string    `json:"gender,omitempty"`
	AvatarURL      *string    `json:"avatar_url,omitempty"`
	PhoneNumber    string     `json:"phone_number,omitempty"`
	Role               UserRole   `json:"role" swaggertype:"string"`
	KycStatus          string     `json:"kyc_status"`
	KycRejectionReason *string    `json:"kyc_rejection_reason,omitempty"`
	Status             string     `json:"status"`
	Locale             string     `json:"locale"`
	PasswordHash       *string    `json:"-"`
	TrustedDevices     []string   `json:"-"`
	Rating             float64    `json:"rating"`
	RatingCount        int32      `json:"rating_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type AdminStats struct {
	TotalUsers      int64 `json:"total_users"`
	DriversCount    int64 `json:"drivers_count"`
	PassengersCount int64 `json:"passengers_count"`
	ChartData       []ChartDataPoint `json:"chart_data"`
}

type ChartDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type UserRepository interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	SearchAdminUsers(ctx context.Context, page, limit int32, search, role, status, kycStatus string) ([]*User, int64, error)
	Create(ctx context.Context, firstName, lastName, phoneNumber string, role UserRole, gender *string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateKYC(ctx context.Context, id uuid.UUID, kycStatus string, firstName, lastName string) (*User, error)
	UpdateKYCStatusAndRole(ctx context.Context, id uuid.UUID, kycStatus string, role UserRole, reason *string) (*User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string, gender *string, locale *string, avatarURL *string) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	AddTrustedDevice(ctx context.Context, id uuid.UUID, deviceID string) error
	UpdateRating(ctx context.Context, id uuid.UUID, newRating float64, newCount int32) error
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status string) error
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error)
	CreateReport(ctx context.Context, reportID, reporterID, reportedID uuid.UUID, reason string) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateCarsModerationStatusByUserID(ctx context.Context, userID uuid.UUID, status string) error
	SystemBlockUser(ctx context.Context, id uuid.UUID) error
}
