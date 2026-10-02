package domain

import (
	"time"

	"github.com/google/uuid"
)

// TrustScore represents a user's trust level in the platform.
// Ranges from 0.00 to 100.00
type TrustScore struct {
	UserID    uuid.UUID `json:"user_id"`
	Score     float64   `json:"score"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	FraudEventTypeMockLocation = "mock_location"
	FraudEventTypeFakeBooking  = "fake_booking"
)

type FraudEvent struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	EventType      string    `json:"event_type"`
	Description    string    `json:"description"`
	PenaltyApplied float64   `json:"penalty_applied"`
	CreatedAt      time.Time `json:"created_at"`
}
