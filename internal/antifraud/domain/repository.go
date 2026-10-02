package domain

import (
	"context"

	"github.com/google/uuid"
)

type AntiFraudRepository interface {
	InitTrustScore(ctx context.Context, userID uuid.UUID, initialScore float64) (*TrustScore, error)
	GetTrustScore(ctx context.Context, userID uuid.UUID) (*TrustScore, error)
	UpdateTrustScore(ctx context.Context, userID uuid.UUID, newScore float64) (*TrustScore, error)
	CreateFraudEvent(ctx context.Context, event *FraudEvent) (*FraudEvent, error)
	GetFraudEventsByUser(ctx context.Context, userID uuid.UUID) ([]FraudEvent, error)
}
