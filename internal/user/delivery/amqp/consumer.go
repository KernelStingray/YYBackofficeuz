package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/user/domain"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

type UserEventsConsumer struct {
	client *rabbitmq.Client
	repo   domain.UserRepository
	cache  *cache.Client
}

func NewUserEventsConsumer(client *rabbitmq.Client, repo domain.UserRepository, cacheClient *cache.Client) *UserEventsConsumer {
	return &UserEventsConsumer{
		client: client,
		repo:   repo,
		cache:  cacheClient,
	}
}

// Start listens to the user.rating.update queue and user.blocked queue with retry logic
func (c *UserEventsConsumer) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := c.initAndConsume(ctx)
			if err == nil {
				return
			}

			slog.Warn("UserEventsConsumer: failed to initialize RabbitMQ queues, retrying in 3 seconds...", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
		}
	}()
	return nil
}

func (c *UserEventsConsumer) initAndConsume(ctx context.Context) error {
	// Setup review.created
	err := c.client.SetupExchangeQueue(
		"yolyolakay.events",
		"topic",
		"review.created",
		"review.created",
	)
	if err != nil {
		return err
	}

	// Setup user.block_requested
	err = c.client.SetupExchangeQueue(
		"yolyolakay.events",
		"topic",
		"user.block_requested",
		"user.block_requested",
	)
	if err != nil {
		return err
	}

	msgsRating, err := c.client.Consume("review.created")
	if err != nil {
		return err
	}

	msgsBlocked, err := c.client.Consume("user.block_requested")
	if err != nil {
		return err
	}

	slog.Info("UserEventsConsumer: successfully connected to RabbitMQ and listening for events")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgsRating:
			if !ok {
				return fmt.Errorf("rating messages channel closed")
			}
			c.handleRatingMessage(msg)
		case msg, ok := <-msgsBlocked:
			if !ok {
				return fmt.Errorf("blocked messages channel closed")
			}
			c.handleBlockedMessage(msg)
		}
	}
}

type ReviewCreatedEvent struct {
	TargetID uuid.UUID `json:"target_id"`
	Rating   int32     `json:"rating"`
}

func (c *UserEventsConsumer) handleRatingMessage(msg amqp091.Delivery) {
	var event ReviewCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		slog.Error("UserEventsConsumer: failed to unmarshal rating event", "error", err)
		msg.Nack(false, false)
		return
	}

	// Fetch old rating and count
	user, err := c.repo.GetByID(context.Background(), event.TargetID)
	if err != nil {
		slog.Error("UserEventsConsumer: user not found", "user", event.TargetID, "error", err)
		msg.Nack(false, false) // discard if user doesn't exist
		return
	}

	oldCount := user.RatingCount
	oldRating := user.Rating

	newCount := oldCount + 1
	newRating := ((oldRating * float64(oldCount)) + float64(event.Rating)) / float64(newCount)

	err = c.repo.UpdateRating(context.Background(), event.TargetID, newRating, newCount)
	if err != nil {
		slog.Error("UserEventsConsumer: failed to update rating", "user", event.TargetID, "error", err)
		msg.Nack(false, true)
	} else {
		slog.Info("UserEventsConsumer: successfully updated rating", "user", event.TargetID, "new_rating", newRating, "new_count", newCount)

		// Invalidate profile cache
		if c.cache != nil {
			profileKey := "driver:profile:" + event.TargetID.String()
			_ = c.cache.Delete(context.Background(), profileKey)
		}

		msg.Ack(false)
	}
}

type BlockEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Reason string    `json:"reason"`
}

func (c *UserEventsConsumer) handleBlockedMessage(msg amqp091.Delivery) {
	var event BlockEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		slog.Error("UserEventsConsumer: failed to unmarshal blocked event", "error", err)
		msg.Nack(false, false)
		return
	}

	err := c.repo.SystemBlockUser(context.Background(), event.UserID)
	if err != nil {
		slog.Error("UserEventsConsumer: failed to block user", "user", event.UserID, "error", err)
		msg.Nack(false, true)
	} else {
		slog.Info("UserEventsConsumer: successfully blocked user", "user", event.UserID, "reason", event.Reason)
		msg.Ack(false)
	}
}
