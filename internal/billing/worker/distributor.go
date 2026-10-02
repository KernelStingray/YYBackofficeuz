package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TaskGiftInject                = "task:gift_inject"
	TaskProcessDueSubscriptions   = "task:process_due_subscriptions"
	TaskCancelStaleTransactions   = "task:cancel_stale_transactions"
	TaskResetExpiredSubscriptions = "task:reset_expired_subscriptions"
)

type GiftInjectPayload struct {
	Quota int32 `json:"quota"`
}

type TaskDistributor interface {
	DistributeTaskGiftInject(ctx context.Context, payload *GiftInjectPayload, opts ...asynq.Option) error
	RegisterCronTasks(scheduler *asynq.Scheduler) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

func (distributor *RedisTaskDistributor) DistributeTaskGiftInject(ctx context.Context, payload *GiftInjectPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskGiftInject, jsonPayload, opts...)
	_, err = distributor.client.EnqueueContext(ctx, task)
	return err
}

func (distributor *RedisTaskDistributor) RegisterCronTasks(scheduler *asynq.Scheduler) error {
	// Run every minute
	if _, err := scheduler.Register("* * * * *", asynq.NewTask(TaskProcessDueSubscriptions, nil)); err != nil {
		return err
	}

	// Run every minute
	if _, err := scheduler.Register("* * * * *", asynq.NewTask(TaskCancelStaleTransactions, nil)); err != nil {
		return err
	}

	// Run every day at midnight
	if _, err := scheduler.Register("0 0 * * *", asynq.NewTask(TaskResetExpiredSubscriptions, nil)); err != nil {
		return err
	}

	return nil
}
