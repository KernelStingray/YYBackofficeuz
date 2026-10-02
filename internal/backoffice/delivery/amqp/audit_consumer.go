package amqp

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AuditConsumer struct {
	rmqClient *rabbitmq.Client
	usecase   domain.BackofficeUseCase
}

func NewAuditConsumer(rmqClient *rabbitmq.Client, usecase domain.BackofficeUseCase) *AuditConsumer {
	return &AuditConsumer{
		rmqClient: rmqClient,
		usecase:   usecase,
	}
}

type auditPayload struct {
	AdminID   string                 `json:"admin_id"`
	Action    string                 `json:"action"`
	Entity    string                 `json:"entity"`
	IPAddress string                 `json:"ip_address"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

func (c *AuditConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := c.rmqClient.SetupExchangeQueue("audit.events", "topic", "audit.logs.backoffice", "audit.log.created")
			if err != nil {
				slog.Warn("Failed to setup queue for audit logs, retrying in 5s...", "error", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			msgs, err := c.rmqClient.Consume("audit.logs.backoffice")
			if err != nil {
				slog.Warn("Failed to consume audit logs, retrying in 5s...", "error", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			slog.Info("Audit consumer successfully subscribed to audit.logs.backoffice")
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-msgs:
					if !ok {
						slog.Warn("Audit logs message channel closed, reconnecting consumer in 2s...")
						time.Sleep(2 * time.Second)
						goto reconnect
					}
					c.handleMessage(ctx, d)
				}
			}
		reconnect:
		}
	}()
}

func (c *AuditConsumer) handleMessage(ctx context.Context, d amqp.Delivery) {
	var payload auditPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		slog.Error("Failed to unmarshal audit log event", "error", err)
		d.Nack(false, false)
		return
	}

	adminID, err := uuid.Parse(payload.AdminID)
	if err != nil {
		adminID = uuid.Nil
	}

	logEntry := &domain.AuditLog{
		ID:      uuid.New(),
		AdminID: adminID,
		Action:  payload.Action,
		Entity:  payload.Entity,
		Payload: payload.Payload,
	}

	if payload.IPAddress != "" {
		logEntry.IPAddress = &payload.IPAddress
	}

	err = c.usecase.CreateAuditLog(ctx, logEntry)
	if err != nil {
		slog.Error("Failed to save audit log", "error", err)
		d.Nack(false, true) // Requeue on failure
		return
	}

	d.Ack(false)
}
