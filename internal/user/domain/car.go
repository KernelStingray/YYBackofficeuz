package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Car struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	GovNumber          string     `json:"gov_number"`
	Brand              string     `json:"brand"`
	Model              string     `json:"model"`
	Color              string     `json:"color"`
	Year               int32      `json:"year"`
	TechPassportPhoto   *string    `json:"tech_passport_photo,omitempty"`
	DriverLicensePhoto  *string    `json:"driver_license_photo,omitempty"`
	TechPassportFront  *string    `json:"tech_passport_front,omitempty"`
	TechPassportBack   *string    `json:"tech_passport_back,omitempty"`
	DriverLicenseFront *string    `json:"driver_license_front,omitempty"`
	DriverLicenseBack  *string    `json:"driver_license_back,omitempty"`
	CarFrontPhoto      *string    `json:"car_front_photo,omitempty"`
	CarBackPhoto       *string    `json:"car_back_photo,omitempty"`
	CarSidePhoto       *string    `json:"car_side_photo,omitempty"`
	ModerationStatus   string     `json:"moderation_status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type CarRepository interface {
	Create(ctx context.Context, car *Car) (*Car, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Car, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Car, error)
	Update(ctx context.Context, car *Car) (*Car, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
