package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/rating/domain"
	ridepb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/ride"
	userpb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type ratingUseCase struct {
	repo          domain.RatingRepository
	amqpPublisher *rabbitmq.Client
	rideClient    ridepb.RideServiceClient
	userClient    userpb.UserServiceClient
}

func NewRatingUseCase(repo domain.RatingRepository, publisher *rabbitmq.Client, rideClient ridepb.RideServiceClient, userClient userpb.UserServiceClient) domain.RatingUseCase {
	return &ratingUseCase{
		repo:          repo,
		amqpPublisher: publisher,
		rideClient:    rideClient,
		userClient:    userClient,
	}
}


func (uc *ratingUseCase) SubmitFeedback(ctx context.Context, rideID, reviewerID, revieweeID uuid.UUID, score int32, comment string, tags []string) (*domain.Feedback, error) {
	// 1. Validation
	if score < 1 || score > 5 {
		return nil, domain.ErrScoreOutOfRange
	}

	if reviewerID == revieweeID {
		return nil, domain.ErrCannotRateYourself
	}

	if uc.rideClient == nil {
		return nil, errors.New("ride grpc client is not initialized")
	}

	rideReq := &ridepb.GetRideByIDRequest{
		RideId: rideID.String(),
	}

	rideRes, err := uc.rideClient.GetRideByID(ctx, rideReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ride details via gRPC: %w", err)
	}

	var rideData struct {
		Status   string
		DriverID string
	}
	rideData.Status = rideRes.GetStatus()
	rideData.DriverID = rideRes.GetDriverId()

	if rideData.Status != "completed" {
		return nil, domain.ErrCanOnlyRateCompletedRides
	}

	// Simple check: either reviewer is driver and reviewee is someone else,
	// or reviewee is driver and reviewer is someone else.
	// Since we don't have passengers list here, we rely on the fact that driver must be one of them.
	// A more robust check would fetch bookings.
	if reviewerID.String() != rideData.DriverID && revieweeID.String() != rideData.DriverID {
		return nil, domain.ErrNotRideParticipant
	}

	if tags == nil {
		tags = []string{}
	}

	// 2. Create Feedback entity
	feedback := &domain.Feedback{
		ID:         uuid.New(),
		RideID:     rideID,
		ReviewerID: reviewerID,
		RevieweeID: revieweeID,
		Score:      score,
		Comment:    strings.TrimSpace(comment),
		Tags:       tags,
		CreatedAt:  time.Now(),
	}

	// 3. Save to DB
	err = uc.repo.Save(ctx, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to save feedback: %w", err)
	}

	// 4. Publish Event for User Service to consume
	event := map[string]interface{}{
		"target_id": revieweeID,
		"rating":    score,
	}
	payload, err := json.Marshal(event)
	if err == nil && uc.amqpPublisher != nil {
		_ = uc.amqpPublisher.Publish(
			context.Background(),
			"yolyolakay.events", // exchange
			"review.created",    // routing key
			payload,
		)
	}

	return feedback, nil
}

func (uc *ratingUseCase) GetFeedbacks(ctx context.Context, revieweeID uuid.UUID, limit, offset int, withComment bool) (*domain.FeedbackListResponse, error) {
	items, total, avg, err := uc.repo.GetFeedbacksByRevieweeID(ctx, revieweeID, limit, offset, withComment)
	if err != nil {
		return nil, err
	}

	tagSummary := make(map[string]int)
	feedbackItems := make([]domain.FeedbackItem, 0, len(items))
	for _, item := range items {
		for _, t := range item.Tags {
			t = strings.TrimSpace(t)
			if t != "" {
				tagSummary[t]++
			}
		}

		fi := domain.FeedbackItem{
			ID:        item.ID,
			RideID:    item.RideID,
			Score:     item.Score,
			Comment:   item.Comment,
			Tags:      item.Tags,
			CreatedAt: item.CreatedAt,
		}

		if uc.userClient != nil {
			userRes, err := uc.userClient.GetPassengerProfile(ctx, &userpb.GetPassengerProfileRequest{
				PassengerId: item.ReviewerID.String(),
			})
			if err == nil && userRes != nil {
				fi.Reviewer = &domain.ReviewerProfile{
					ID:     item.ReviewerID,
					Name:   userRes.GetName(),
					Avatar: userRes.GetAvatarUrl(),
				}
			}
		}

		if fi.Reviewer == nil {
			fi.Reviewer = &domain.ReviewerProfile{
				ID: item.ReviewerID,
			}
		}

		feedbackItems = append(feedbackItems, fi)
	}

	return &domain.FeedbackListResponse{
		Items:      feedbackItems,
		Total:      total,
		Average:    avg,
		TagSummary: tagSummary,
	}, nil
}

