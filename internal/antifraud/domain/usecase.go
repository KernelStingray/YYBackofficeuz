package domain

import (
	"context"

	"github.com/google/uuid"
)

type AntiFraudUseCase interface {
	// ReportMockLocation is called when a mobile app detects mock GPS usage
	ReportMockLocation(ctx context.Context, userID uuid.UUID, description string) (*TrustScore, error)

	// GetTrustScore returns the current score. Initializes to 100.0 if not found.
	GetTrustScore(ctx context.Context, userID uuid.UUID) (*TrustScore, error)
}
