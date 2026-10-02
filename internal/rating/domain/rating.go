package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrScoreOutOfRange           = errors.New("score must be between 1 and 5")
	ErrCannotRateYourself        = errors.New("cannot rate yourself")
	ErrCanOnlyRateCompletedRides = errors.New("can only rate completed rides")
	ErrNotRideParticipant        = errors.New("at least one of the participants must be the driver of the ride")
)

// Feedback represents a user rating and review for a specific ride.
type Feedback struct {
	ID         uuid.UUID `json:"id"`
	RideID     uuid.UUID `json:"ride_id"`
	ReviewerID uuid.UUID `json:"reviewer_id"` // User who gives the rating
	RevieweeID uuid.UUID `json:"reviewee_id"` // User who receives the rating
	Score      int32     `json:"score"`       // 1 to 5
	Comment    string    `json:"comment"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReviewerProfile contains basic public info of the review author
type ReviewerProfile struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar string    `json:"avatar"`
}

// FeedbackItem represents a single review in the public review list
type FeedbackItem struct {
	ID        uuid.UUID        `json:"id"`
	RideID    uuid.UUID        `json:"ride_id"`
	Reviewer  *ReviewerProfile `json:"reviewer,omitempty"`
	Score     int32            `json:"score"`
	Comment   string           `json:"comment"`
	Tags      []string         `json:"tags"`
	CreatedAt time.Time        `json:"created_at"`
}

// FeedbackListResponse represents the paginated reviews list with aggregates
type FeedbackListResponse struct {
	Items      []FeedbackItem `json:"items"`
	Total      int64          `json:"total"`
	Average    float64        `json:"average"`
	TagSummary map[string]int `json:"tag_summary"`
}

// FeedbackEvent is sent to the message broker when a new rating is created
type FeedbackEvent struct {
	RevieweeID uuid.UUID `json:"reviewee_id"`
	Score      int32     `json:"score"`
}

