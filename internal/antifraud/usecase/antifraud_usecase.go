package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/antifraud/domain"
	pgx "github.com/jackc/pgx/v5"
)

type antiFraudUseCase struct {
	repo domain.AntiFraudRepository
}

func NewAntiFraudUseCase(repo domain.AntiFraudRepository) domain.AntiFraudUseCase {
	return &antiFraudUseCase{
		repo: repo,
	}
}

// Penalty for using a fake GPS location
const MockLocationPenalty = 10.0

func (u *antiFraudUseCase) GetTrustScore(ctx context.Context, userID uuid.UUID) (*domain.TrustScore, error) {
	score, err := u.repo.GetTrustScore(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Initialize with 100.0 score
			return u.repo.InitTrustScore(ctx, userID, 100.0)
		}
		return nil, err
	}
	return score, nil
}

func (u *antiFraudUseCase) ReportMockLocation(ctx context.Context, userID uuid.UUID, description string) (*domain.TrustScore, error) {
	// 1. Get current score (or init if not exists)
	score, err := u.GetTrustScore(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Log the event
	event := &domain.FraudEvent{
		ID:             uuid.New(),
		UserID:         userID,
		EventType:      domain.FraudEventTypeMockLocation,
		Description:    description,
		PenaltyApplied: MockLocationPenalty,
	}
	_, err = u.repo.CreateFraudEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	// 3. Apply penalty and update score
	newScore := score.Score - MockLocationPenalty
	// Ensure score doesn't drop below 0.0
	newScore = math.Max(0.0, newScore)

	updatedScore, err := u.repo.UpdateTrustScore(ctx, userID, newScore)
	if err != nil {
		return nil, err
	}

	return updatedScore, nil
}
