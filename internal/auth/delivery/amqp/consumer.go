package amqp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/auth/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

type AuthEventsConsumer struct {
	client  *rabbitmq.Client
	useCase domain.AuthUseCase
}

func NewAuthEventsConsumer(client *rabbitmq.Client, useCase domain.AuthUseCase) *AuthEventsConsumer {
	return &AuthEventsConsumer{
		client:  client,
		useCase: useCase,
	}
}

// Start listens to the user.verified queue
func (c *AuthEventsConsumer) Start(ctx context.Context) error {
	err := c.client.SetupExchangeQueue(
		"yolyolakay.events",
		"topic",
		"auth.user.verified",
		"user.verified",
	)
	if err != nil {
		return err
	}

	msgs, err := c.client.Consume("auth.user.verified")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.handleUserVerified(ctx, msg)
			}
		}
	}()

	return nil
}

type UserVerifiedEvent struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (c *AuthEventsConsumer) handleUserVerified(ctx context.Context, msg amqp091.Delivery) {
	var event UserVerifiedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		slog.Error("AuthEventsConsumer: failed to unmarshal event", "error", err)
		msg.Nack(false, false)
		return
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		slog.Error("AuthEventsConsumer: invalid user_id", "error", err)
		msg.Nack(false, false)
		return
	}

	err = c.useCase.TerminateAllUserSessions(ctx, userID)
	if err != nil {
		slog.Error("AuthEventsConsumer: failed to terminate sessions", "user_id", event.UserID, "error", err)
		msg.Nack(false, true)
	} else {
		slog.Info("AuthEventsConsumer: successfully terminated sessions for verified user", "user_id", event.UserID)
		msg.Ack(false)
	}
}
