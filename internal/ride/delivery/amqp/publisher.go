package amqp

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type EventPublisher struct {
	client *rabbitmq.Client
}

func NewEventPublisher(client *rabbitmq.Client) *EventPublisher {
	return &EventPublisher{
		client: client,
	}
}

// NotificationPayload matches the typical notification structure
type NotificationPayload struct {
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	EventType string `json:"event_type"`
}

func (p *EventPublisher) PublishNotification(ctx context.Context, userID uuid.UUID, title, message, eventType string) error {
	payload := NotificationPayload{
		UserID:    userID.String(),
		Title:     title,
		Message:   message,
		EventType: eventType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "notification_exchange", "notification.send", body)
}

func (p *EventPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return p.client.Publish(ctx, exchange, routingKey, body)
}
