package domain

import (
	"context"

	"github.com/google/uuid"
)

// RatingRepository defines the data access layer for feedbacks
type RatingRepository interface {
	Save(ctx context.Context, feedback *Feedback) error
	GetFeedbacksByRevieweeID(ctx context.Context, revieweeID uuid.UUID, limit, offset int, withComment bool) ([]Feedback, int64, float64, error)
}

