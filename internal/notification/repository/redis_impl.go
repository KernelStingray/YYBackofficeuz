package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/redis/go-redis/v9"
)

type redisDeviceTokenRepo struct {
	client *redis.Client
}

func NewRedisDeviceTokenRepository(client *redis.Client) domain.DeviceTokenRepository {
	return &redisDeviceTokenRepo{
		client: client,
	}
}

func (r *redisDeviceTokenRepo) SaveDeviceToken(ctx context.Context, token *domain.DeviceToken) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	fieldKey := token.SessionID.String()
	if token.SessionID == uuid.Nil {
		fieldKey = token.FCMToken
	}

	key := fmt.Sprintf("fcm_tokens:%s", token.UserID.String())
	err = r.client.HSet(ctx, key, fieldKey, tokenJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to save FCM token to redis: %w", err)
	}
	return nil
}

func (r *redisDeviceTokenRepo) GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]*domain.DeviceToken, error) {
	key := fmt.Sprintf("fcm_tokens:%s", userID.String())
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tokens []*domain.DeviceToken
	for _, val := range res {
		var token domain.DeviceToken
		if err := json.Unmarshal([]byte(val), &token); err == nil {
			tokens = append(tokens, &token)
		}
	}
	return tokens, nil
}

func (r *redisDeviceTokenRepo) DeleteDeviceToken(ctx context.Context, userID, sessionID uuid.UUID) error {
	key := fmt.Sprintf("fcm_tokens:%s", userID.String())
	return r.client.HDel(ctx, key, sessionID.String()).Err()
}
