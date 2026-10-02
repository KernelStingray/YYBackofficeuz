package domain

import (
	"context"

	"github.com/google/uuid"
)

// TokenResponse represents the JWT token pair returned after authentication.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}

// AuthUseCase defines the contract for authentication business logic.
// All delivery layers (HTTP, AMQP) should depend on this interface,
// not on the concrete usecase implementation.
type AuthUseCase interface {
	SendOTP(ctx context.Context, phone, ip, reqType string) error
	VerifyOTP(ctx context.Context, phone, code, deviceIP, deviceID, userAgent, osType string, password ...string) (*TokenResponse, error)
	SetupPassword(ctx context.Context, userID uuid.UUID, password string) error
	LoginWithPassword(ctx context.Context, phone, password, deviceIP, deviceID, userAgent, osType string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken, deviceIP string) (*TokenResponse, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	TerminateAllUserSessions(ctx context.Context, userID uuid.UUID) error
	TerminateSession(ctx context.Context, userID uuid.UUID, deviceID string) error
	SaveMockOTP(ctx context.Context, phone, code string) error
}
