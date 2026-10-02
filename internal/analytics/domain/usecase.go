package domain

import (
	"context"

	"github.com/google/uuid"
)

type AnalyticsUseCase interface {
	LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, payload map[string]interface{}) error
	GetEventStatsFiltered(ctx context.Context, from, to, groupBy, eventType string) ([]EventStats, error)
}
