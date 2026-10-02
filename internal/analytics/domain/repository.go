package domain

import (
	"context"
)

type AnalyticsRepository interface {
	LogEvent(ctx context.Context, event *Event) error
	GetEventStatsFiltered(ctx context.Context, from, to, groupBy, eventType string) ([]EventStats, error)
}
