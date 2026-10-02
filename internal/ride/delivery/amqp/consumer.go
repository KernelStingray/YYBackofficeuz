package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RideConsumer struct {
	rmq     *rabbitmq.Client
	repo    domain.RideRepository
	cache   domain.RideCache
	useCase domain.RideUseCase
}

func NewRideConsumer(rmq *rabbitmq.Client, repo domain.RideRepository, cache domain.RideCache, useCase domain.RideUseCase) *RideConsumer {
	return &RideConsumer{
		rmq:     rmq,
		repo:    repo,
		cache:   cache,
		useCase: useCase,
	}
}

func (c *RideConsumer) Start(ctx context.Context) error {
	// 1. Listen for user.deleted
	queueDeleted := "ride.user.deleted.queue"
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueDeleted, "user.deleted"); err != nil {
		return fmt.Errorf("failed to setup queue %s: %w", queueDeleted, err)
	}
	msgsDeleted, err := c.rmq.Consume(queueDeleted)
	if err != nil {
		return fmt.Errorf("failed to consume user.deleted: %w", err)
	}

	// 2. Listen for user.rating.update
	queueRating := "ride.user.rating.queue"
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueRating, "user.rating.update"); err != nil {
		return fmt.Errorf("failed to setup queue %s: %w", queueRating, err)
	}
	msgsRating, err := c.rmq.Consume(queueRating)
	if err != nil {
		return fmt.Errorf("failed to consume user.rating.update: %w", err)
	}

	// 3. Listen for billing.limit_exceeded
	queueLimit := "ride.billing.limit_exceeded.queue"
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueLimit, "billing.limit_exceeded"); err != nil {
		return fmt.Errorf("failed to setup queue %s: %w", queueLimit, err)
	}
	msgsLimit, err := c.rmq.Consume(queueLimit)
	if err != nil {
		return fmt.Errorf("failed to consume billing.limit_exceeded: %w", err)
	}

	// 4. Listen for user.verified
	queueVerified := "ride.user.verified.queue"
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueVerified, "user.verified"); err != nil {
		return fmt.Errorf("failed to setup queue %s: %w", queueVerified, err)
	}
	msgsVerified, err := c.rmq.Consume(queueVerified)
	if err != nil {
		return fmt.Errorf("failed to consume user.verified: %w", err)
	}

	go c.handleUserDeleted(ctx, msgsDeleted)
	go c.handleUserRating(ctx, msgsRating)
	go c.handleLimitExceeded(ctx, msgsLimit)
	go c.handleUserProfileUpdated(ctx, msgsVerified)

	return nil
}

func (c *RideConsumer) handleUserDeleted(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			var payload struct {
				UserID string `json:"user_id"`
			}
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				slog.Error("Failed to unmarshal user.deleted", "error", err)
				d.Nack(false, false)
				continue
			}

			userUUID, err := uuid.Parse(payload.UserID)
			if err != nil {
				slog.Error("Invalid user UUID", "error", err)
				d.Nack(false, false)
				continue
			}

			if err := c.repo.DeleteRidesByUserID(ctx, userUUID); err != nil {
				slog.Error(fmt.Sprintf("Failed to delete rides for user %s", userUUID), "error", err)
				d.Nack(false, true)
				continue
			}
			d.Ack(false)
		}
	}
}

func (c *RideConsumer) handleUserRating(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			var payload struct {
				RevieweeID uuid.UUID `json:"target_id"`
				Score      int32     `json:"rating"`
			}
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				slog.Error("Failed to unmarshal user.rating.update", "error", err)
				d.Nack(false, false)
				continue
			}

			// Invalidate driver profile cache so it gets fetched with new rating on next read
			if err := c.cache.InvalidateDriverProfile(ctx, payload.RevieweeID); err != nil {
				slog.Error(fmt.Sprintf("Failed to invalidate cache for driver %s", payload.RevieweeID), "error", err)
			}
			d.Ack(false)
		}
	}
}

func (c *RideConsumer) handleLimitExceeded(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			var payload struct {
				RideID    string  `json:"ride_id"`
				UserID    string  `json:"user_id"`
				BookingID *string `json:"booking_id"`
			}
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				slog.Error("Failed to unmarshal billing.limit_exceeded", "error", err)
				d.Nack(false, false)
				continue
			}

			rideUUID, err := uuid.Parse(payload.RideID)
			if err != nil {
				slog.Error("Invalid ride UUID in billing.limit_exceeded", "error", err)
				d.Nack(false, false)
				continue
			}

			userUUID, err := uuid.Parse(payload.UserID)
			if err != nil {
				slog.Error("Invalid user UUID in billing.limit_exceeded", "error", err)
				d.Nack(false, false)
				continue
			}

			var bookingUUIDPtr *uuid.UUID
			if payload.BookingID != nil && *payload.BookingID != "" {
				parsed, err := uuid.Parse(*payload.BookingID)
				if err == nil {
					bookingUUIDPtr = &parsed
				}
			}

			if err := c.useCase.HandleLimitExceeded(ctx, rideUUID, bookingUUIDPtr, userUUID); err != nil {
				slog.Error("Failed to handle limit exceeded", "error", err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}
}

func (c *RideConsumer) handleUserProfileUpdated(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			var payload struct {
				UserID string `json:"user_id"`
			}
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				slog.Error("Failed to unmarshal user profile update", "error", err)
				d.Nack(false, false)
				continue
			}

			userUUID, err := uuid.Parse(payload.UserID)
			if err != nil {
				slog.Error("Invalid user UUID in profile update", "error", err)
				d.Nack(false, false)
				continue
			}

			// Invalidate driver profile cache
			if err := c.cache.InvalidateDriverProfile(ctx, userUUID); err != nil {
				slog.Error(fmt.Sprintf("Failed to invalidate cache for driver %s", payload.UserID), "error", err)
			}
			d.Ack(false)
		}
	}
}
