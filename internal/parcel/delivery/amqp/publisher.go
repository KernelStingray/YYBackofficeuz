package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/parcel/domain"
	rmqPkg "github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type amqpPublisher struct {
	client *rmqPkg.Client
}

func NewEventPublisher(client *rmqPkg.Client) domain.EventPublisher {
	return &amqpPublisher{client: client}
}

func (p *amqpPublisher) PublishNotification(ctx context.Context, userID uuid.UUID, title, message, eventType string) error {
	// not used in parcel service yet, returning nil or implement basic publisher
	return nil
}

func (p *amqpPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return p.client.Publish(ctx, exchange, routingKey, body)
}
