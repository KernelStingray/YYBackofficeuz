package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type BillingEventsConsumer struct {
	rmq     *rabbitmq.Client
	repo    domain.BillingRepository
	useCase domain.BillingUseCase
}

func NewBillingEventsConsumer(rmq *rabbitmq.Client, repo domain.BillingRepository, useCase domain.BillingUseCase) *BillingEventsConsumer {
	return &BillingEventsConsumer{
		rmq:     rmq,
		repo:    repo,
		useCase: useCase,
	}
}

func (c *BillingEventsConsumer) Start(ctx context.Context) error {
	queueName := "billing.events.queue"

	// Bind to ride events
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "ride.created"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for ride.created: %w", err)
	}
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "ride.cancelled"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for ride.cancelled: %w", err)
	}
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "ride.booking.confirmed"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for ride.booking.confirmed: %w", err)
	}
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "ride.completed"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for ride.completed: %w", err)
	}

	// Bind to parcel events
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "parcel.cancelled"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for parcel.cancelled: %w", err)
	}

	// Bind to user events (for trial creation)
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "user.created"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for user.created: %w", err)
	}
	if err := c.rmq.SetupExchangeQueue("yolyolakay.events", "topic", queueName, "user.verified"); err != nil {
		return fmt.Errorf("failed to setup exchange and queue for user.verified: %w", err)
	}

	msgs, err := c.rmq.Consume(queueName)
	if err != nil {
		return fmt.Errorf("failed to start consuming billing events: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				c.handleMessage(ctx, d)
			}
		}
	}()

	return nil
}

func (c *BillingEventsConsumer) handleMessage(ctx context.Context, d amqp.Delivery) {
	switch d.RoutingKey {
	case "ride.created":
		c.handleRideCreated(ctx, d)
	case "ride.booking.confirmed":
		c.handleBookingConfirmed(ctx, d)
	case "ride.completed":
		c.handleRideCompleted(ctx, d)
	case "ride.cancelled":
		c.handleRideCancelled(ctx, d)
	case "parcel.cancelled":
		c.handleParcelCancelled(ctx, d)
	case "user.created":
		c.handleUserCreated(ctx, d)
	case "user.verified":
		c.handleUserVerified(ctx, d)
	default:
		d.Ack(false)
	}
}

func (c *BillingEventsConsumer) handleUserCreated(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal user.created payload", "error", err)
		d.Nack(false, false)
		return
	}

	// Only drivers receive driver_trial subscription.
	// Passengers use the platform 100% free and do not have driver subscriptions.
	if strings.ToLower(payload.Role) == "driver" {
		userID, err := uuid.Parse(payload.UserID)
		if err == nil {
			err = c.useCase.StartTrialSubscription(ctx, userID, "ride")
			if err != nil {
				slog.Error("Failed to start trial subscription for user", "user_id", userID, "error", err)
				d.Nack(false, true)
				return
			}
		}
	}

	d.Ack(false)
}

func (c *BillingEventsConsumer) handleUserVerified(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal user.verified payload", "error", err)
		d.Nack(false, false)
		return
	}

	if strings.ToLower(payload.Role) == "driver" {
		userID, err := uuid.Parse(payload.UserID)
		if err == nil {
			existingSub, _ := c.repo.GetLatestSubscription(ctx, userID)
			if existingSub == nil {
				if err := c.useCase.StartTrialSubscription(ctx, userID, "ride"); err != nil {
					slog.Error("Failed to start trial subscription for verified driver", "user_id", userID, "error", err)
					d.Nack(false, true)
					return
				}
			}
		}
	}

	d.Ack(false)
}

func (c *BillingEventsConsumer) handleRideCreated(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		DriverID string `json:"driver_id"`
		RideID   string `json:"ride_id"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal ride.created payload", "error", err)
		d.Nack(false, false)
		return
	}

	driverUUID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		slog.Error("Invalid driver UUID in ride.created", "error", err)
		d.Nack(false, false)
		return
	}

	if err := c.repo.DeductRideLimit(ctx, driverUUID); err != nil {
		if errors.Is(err, domain.ErrLimitExceeded) {
			slog.Warn("Limit exceeded for driver", "driver_id", driverUUID)
			limitPayload, _ := json.Marshal(map[string]interface{}{
				"ride_id": payload.RideID,
				"user_id": payload.DriverID,
				"reason":  "limit exceeded",
			})
			if err := c.rmq.Publish(ctx, "yolyolakay.events", "billing.limit_exceeded", limitPayload); err != nil {
				slog.Error("Failed to publish", "error", err)
			}
			d.Ack(false)
			return
		}
		slog.Error("Failed to deduct ride limit for driver", "driver_id", driverUUID, "error", err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (c *BillingEventsConsumer) handleBookingConfirmed(ctx context.Context, d amqp.Delivery) {
	// Passengers do not have ride publication limits. Simply acknowledge the event.
	d.Ack(false)
}

func (c *BillingEventsConsumer) handleRideCompleted(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		RideID   string `json:"ride_id"`
		DriverID string `json:"driver_id"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal ride.completed payload", "error", err)
		d.Nack(false, false)
		return
	}

	driverUUID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		slog.Error("Invalid driver UUID in ride.completed", "error", err)
		d.Nack(false, false)
		return
	}

	if err := c.useCase.ExpireTrialSubscription(ctx, driverUUID); err != nil {
		slog.Error("Failed to expire trial subscription for driver", "driver_id", driverUUID, "error", err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (c *BillingEventsConsumer) handleRideCancelled(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		DriverID     string `json:"driver_id"`
		RideID       string `json:"ride_id"`
		Reason       string `json:"reason"`
		WasActivated bool   `json:"was_activated"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal ride.cancelled payload", "error", err)
		d.Nack(false, false)
		return
	}

	// Only refund quota if the ride was actually activated and deducted.
	// Draft rides never deducted quota, so refunding them would create free rides out of thin air.
	if !payload.WasActivated {
		slog.Debug("Ride was never activated, skipping quota refund", "ride_id", payload.RideID)
		d.Ack(false)
		return
	}

	driverUUID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		slog.Error("Invalid driver UUID in ride.cancelled", "error", err)
		d.Nack(false, false)
		return
	}

	if err := c.repo.RefundRideLimit(ctx, driverUUID); err != nil {
		slog.Error("Failed to refund ride limit for driver", "driver_id", driverUUID, "error", err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (c *BillingEventsConsumer) handleParcelCancelled(ctx context.Context, d amqp.Delivery) {
	var payload struct {
		DriverID string `json:"driver_id"`
		ParcelID string `json:"parcel_id"`
		Reason   string `json:"reason"`
	}

	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal parcel.cancelled payload", "error", err)
		d.Nack(false, false)
		return
	}

	driverUUID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		slog.Error("Invalid driver UUID in parcel.cancelled", "error", err)
		d.Nack(false, false)
		return
	}

	if err := c.repo.RefundParcelLimit(ctx, driverUUID); err != nil {
		slog.Error("Failed to refund parcel limit for driver", "driver_id", driverUUID, "error", err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}
