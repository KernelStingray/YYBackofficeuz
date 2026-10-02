package domain

import (
	"context"
	"time"
)

// OTPRepository defines the interface for managing OTP requests and rate limiting in Redis.
type OTPRepository interface {
	// SaveOTP saves a 5-digit code associated with a phone number and a TTL
	SaveOTP(ctx context.Context, phone string, code string) error

	// VerifyOTP checks if the provided code matches the one stored in Redis.
	// Returns true if valid, and usually deletes it to prevent reuse.
	VerifyOTP(ctx context.Context, phone string, code string) (bool, error)

	// IncrementRateLimit increments the request count for a phone number in the current 5-minute window.
	IncrementRateLimit(ctx context.Context, phone string) error

	// CheckRateLimit returns the current request count for the phone number.
	CheckRateLimit(ctx context.Context, phone string) (int, error)

	// IncrementIPRateLimit increments the request count for an IP in the current 1-hour window.
	IncrementIPRateLimit(ctx context.Context, ip string) error

	// CheckIPRateLimit returns the current request count for the IP.
	CheckIPRateLimit(ctx context.Context, ip string) (int, error)

	// IsPhoneLocked checks if the phone is currently locked due to too many failed attempts
	IsPhoneLocked(ctx context.Context, phone string) (bool, error)

	// IncrementFailedAttempts increments the failed OTP attempt counter
	IncrementFailedAttempts(ctx context.Context, phone string) (int, error)

	// LockPhone locks the phone for a specified duration
	LockPhone(ctx context.Context, phone string, duration time.Duration) error

	// ResetFailedAttempts resets the failed attempt counter
	ResetFailedAttempts(ctx context.Context, phone string) error
}
