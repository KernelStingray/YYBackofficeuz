package domain

import (
	"context"

	"github.com/google/uuid"
)

// RatingUseCase defines the business logic for submitting ratings
type RatingUseCase interface {
	SubmitFeedback(ctx context.Context, rideID, reviewerID, revieweeID uuid.UUID, score int32, comment string, tags []string) (*Feedback, error)
	GetFeedbacks(ctx context.Context, revieweeID uuid.UUID, limit, offset int, withComment bool) (*FeedbackListResponse, error)
}

