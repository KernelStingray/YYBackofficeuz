package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Session represents a user's active session on a specific device.
type Session struct {
	SessionID    uuid.UUID `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	UserAgent    string    `json:"user_agent"`
	OSType       string    `json:"os_type"`
	RefreshToken string    `json:"refresh_token"`
	DeviceIP     string    `json:"device_ip"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// SessionRepository defines the interface for storing and managing refresh tokens, usually in Redis.
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, refreshToken string) (*Session, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	DeleteByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteByToken(ctx context.Context, refreshToken string) error
	SaveGraceToken(ctx context.Context, oldToken string, response *TokenResponse, ttl time.Duration) error
	GetGraceToken(ctx context.Context, oldToken string) (*TokenResponse, error)
}
