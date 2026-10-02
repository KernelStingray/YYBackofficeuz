package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/auth/domain"
	redisClient "github.com/redis/go-redis/v9"
)

type sessionRepo struct {
	client *redisClient.Client
}

func NewSessionRepository(client *redisClient.Client) domain.SessionRepository {
	return &sessionRepo{client: client}
}

func getSessionKey(userID uuid.UUID, deviceID string) string {
	return fmt.Sprintf("session:%s:%s", userID.String(), deviceID)
}

func getUserSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_sessions:%s", userID.String())
}

func getTokenKey(refreshToken string) string {
	return fmt.Sprintf("token:%s", refreshToken)
}

func (r *sessionRepo) Save(ctx context.Context, session *domain.Session) error {
	sessionKey := getSessionKey(session.UserID, session.DeviceID)
	userSessionsKey := getUserSessionsKey(session.UserID)
	tokenKey := getTokenKey(session.RefreshToken)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session already expired")
	}

	pipe := r.client.TxPipeline()
	// Store session data
	pipe.Set(ctx, sessionKey, data, ttl)
	// Add to user sessions set
	pipe.SAdd(ctx, userSessionsKey, session.DeviceID)
	pipe.Expire(ctx, userSessionsKey, ttl) // Extend set TTL, though individual device expiry is trickier, it's fine
	// Store token mapping
	pipe.Set(ctx, tokenKey, sessionKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *sessionRepo) GetByToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	tokenKey := getTokenKey(refreshToken)
	sessionKey, err := r.client.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return nil, nil // Not found
		}
		return nil, err
	}

	val, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepo) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	userSessionsKey := getUserSessionsKey(userID)
	deviceIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, err
	}

	var sessions []*domain.Session
	for _, deviceID := range deviceIDs {
		sessionKey := getSessionKey(userID, deviceID)
		val, err := r.client.Get(ctx, sessionKey).Result()
		if err == nil {
			var session domain.Session
			if err := json.Unmarshal([]byte(val), &session); err == nil {
				sessions = append(sessions, &session)
			}
		} else if errors.Is(err, redisClient.Nil) {
			// Session expired but deviceID still in Set, cleanup
			r.client.SRem(ctx, userSessionsKey, deviceID)
		}
	}

	return sessions, nil
}

func (r *sessionRepo) DeleteByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	sessionKey := getSessionKey(userID, deviceID)
	userSessionsKey := getUserSessionsKey(userID)

	// First get session to delete token mapping
	val, err := r.client.Get(ctx, sessionKey).Result()
	if err == nil {
		var session domain.Session
		if err := json.Unmarshal([]byte(val), &session); err == nil {
			tokenKey := getTokenKey(session.RefreshToken)
			r.client.Del(ctx, tokenKey)
		}
	}

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, deviceID)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *sessionRepo) DeleteByToken(ctx context.Context, refreshToken string) error {
	session, err := r.GetByToken(ctx, refreshToken)
	if err != nil || session == nil {
		return nil
	}
	return r.DeleteByDeviceID(ctx, session.UserID, session.DeviceID)
}

func (r *sessionRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	sessions, err := r.GetAllByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		_ = r.DeleteByDeviceID(ctx, userID, session.DeviceID)
	}
	return nil
}

func getGraceTokenKey(refreshToken string) string {
	return fmt.Sprintf("grace:token:%s", refreshToken)
}

func (r *sessionRepo) SaveGraceToken(ctx context.Context, oldToken string, response *domain.TokenResponse, ttl time.Duration) error {
	key := getGraceTokenKey(oldToken)
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *sessionRepo) GetGraceToken(ctx context.Context, oldToken string) (*domain.TokenResponse, error) {
	key := getGraceTokenKey(oldToken)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return nil, nil // Not in grace period
		}
		return nil, err
	}

	var res domain.TokenResponse
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return nil, err
	}
	return &res, nil
}
