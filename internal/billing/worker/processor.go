package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskGiftInject(ctx context.Context, task *asynq.Task) error
	ProcessTaskProcessDueSubscriptions(ctx context.Context, task *asynq.Task) error
	ProcessTaskCancelStaleTransactions(ctx context.Context, task *asynq.Task) error
	ProcessTaskResetExpiredSubscriptions(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	repo   domain.BillingRepository
	mq     *rabbitmq.Client
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, repo domain.BillingRepository, mq *rabbitmq.Client) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				"default": 10,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("process task failed", "error", err)
			}),
		},
	)

	return &RedisTaskProcessor{
		server: server,
		repo:   repo,
		mq:     mq,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskGiftInject, processor.ProcessTaskGiftInject)
	mux.HandleFunc(TaskProcessDueSubscriptions, processor.ProcessTaskProcessDueSubscriptions)
	mux.HandleFunc(TaskCancelStaleTransactions, processor.ProcessTaskCancelStaleTransactions)
	mux.HandleFunc(TaskResetExpiredSubscriptions, processor.ProcessTaskResetExpiredSubscriptions)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) ProcessTaskGiftInject(ctx context.Context, task *asynq.Task) error {
	var payload GiftInjectPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Add free parcels
	err := processor.repo.AddFreeParcelsToPremium(ctx, payload.Quota)
	if err != nil {
		return fmt.Errorf("failed to add free parcels to premium: %w", err)
	}

	// 2. Send fan-out push notification via rabbitmq
	msg := map[string]interface{}{
		"action": "gifts_distributed",
		"quota":  payload.Quota,
	}

	msgBytes, _ := json.Marshal(msg)
	err = processor.mq.Publish(ctx, "yolyolakay.events", "marketing.gifts_distributed", msgBytes)
	if err != nil {
		slog.Error("failed to publish fanout notification", "error", err)
		// We don't return error here because the core job (db update) is done.
	}

	slog.Info("Successfully injected free parcels to premium drivers", "quota", payload.Quota)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskCancelStaleTransactions(ctx context.Context, task *asynq.Task) error {
	err := processor.repo.CancelStaleTransactions(ctx)
	if err != nil {
		slog.Error("Failed to cancel stale transactions", "error", err)
		return err
	}
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskProcessDueSubscriptions(ctx context.Context, task *asynq.Task) error {
	slog.Info("Checking for due subscriptions...")

	subs, err := processor.repo.GetSubscriptionsForBilling(ctx)
	if err != nil {
		slog.Error("Error fetching subscriptions for billing", "error", err)
		return err
	}

	if len(subs) == 0 {
		return nil
	}

	slog.Info("Found %d subscriptions to process", "count", len(subs))

	for _, sub := range subs {
		processor.processSubscription(ctx, sub)
	}

	return nil
}

func (processor *RedisTaskProcessor) processSubscription(ctx context.Context, sub domain.Subscription) {
	var price float64
	switch sub.PlanName {
	case "plus":
		price = 5000
	case "pro":
		price = 7000
	case "premium":
		price = 10000
	default:
		price = 5000
	}

	tx := &domain.Transaction{
		ID:       uuid.New(),
		UserID:   sub.UserID,
		Amount:   price,
		Currency: "UZS",
		Type:     "subscription",
		Status:   domain.TransactionStatusPending,
		Provider: "system_mock",
	}

	err := processor.repo.CreateTransaction(ctx, tx)
	if err != nil {
		slog.Error("Failed to create tx for sub", "sub_id", sub.ID, "error", err)
		return
	}

	// Mock payment success
	paymentSuccess := true

	if paymentSuccess {
		err = processor.repo.UpdateTransactionStatus(ctx, tx.ID, domain.TransactionStatusSuccess, nil)
		if err == nil {
			newEndDate := sub.EndDate.AddDate(0, 1, 0)
			if newEndDate.Before(time.Now()) {
				newEndDate = time.Now().AddDate(0, 1, 0)
			}

			processor.repo.UpdateSubscriptionStatus(ctx, sub.ID, domain.SubscriptionStatusActive, newEndDate)
			slog.Info("Successfully renewed sub", "sub_id", sub.ID)
		}
	} else {
		err = processor.repo.UpdateTransactionStatus(ctx, tx.ID, domain.TransactionStatusFailed, nil)
		if err == nil {
			if sub.Status == domain.SubscriptionStatusPastDue && time.Since(sub.EndDate) > 7*24*time.Hour {
				processor.repo.UpdateSubscriptionStatus(ctx, sub.ID, domain.SubscriptionStatusCanceled, sub.EndDate)
				slog.Error("Canceled sub due to failed payment", "sub_id", sub.ID)
			} else {
				processor.repo.UpdateSubscriptionStatus(ctx, sub.ID, domain.SubscriptionStatusPastDue, sub.EndDate)
				slog.Info("Marked sub as past_due", "sub_id", sub.ID)
			}
		}
	}
}

func (processor *RedisTaskProcessor) ProcessTaskResetExpiredSubscriptions(ctx context.Context, task *asynq.Task) error {
	count, err := processor.repo.ResetExpiredSubscriptions(ctx)
	if err != nil {
		slog.Error("Failed to reset expired subscriptions", "error", err)
		return err
	}
	slog.Info("Reset expired subscriptions successfully", "count", count)
	return nil
}
