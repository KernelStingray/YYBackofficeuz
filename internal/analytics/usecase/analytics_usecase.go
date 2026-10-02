package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/analytics/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type analyticsUseCase struct {
	repo      domain.AnalyticsRepository
	rmqClient *rabbitmq.Client
}

func NewAnalyticsUseCase(repo domain.AnalyticsRepository, rmqClient *rabbitmq.Client) domain.AnalyticsUseCase {
	return &analyticsUseCase{
		repo:      repo,
		rmqClient: rmqClient,
	}
}

func (uc *analyticsUseCase) LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, payload map[string]interface{}) error {
	if eventType == "" {
		return errors.New("event_type cannot be empty")
	}

	event := &domain.Event{
		ID:        uuid.New(),
		EventType: eventType,
		UserID:    userID,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	// Publish to RabbitMQ for async processing (e.g. DWH / Data Lake insertion)
	eventPayload, _ := json.Marshal(event)
	_ = uc.rmqClient.SetupExchangeQueue("analytics.events", "fanout", "analytics.raw_events", "")
	if err := uc.rmqClient.Publish(ctx, "analytics.events", "", eventPayload); err != nil {
		slog.Error("Failed to publish", "error", err)
	}

	return uc.repo.LogEvent(ctx, event)
}

func (uc *analyticsUseCase) GetEventStatsFiltered(ctx context.Context, from, to, groupBy, eventType string) ([]domain.EventStats, error) {
	return uc.repo.GetEventStatsFiltered(ctx, from, to, groupBy, eventType)
}
