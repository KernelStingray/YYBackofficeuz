package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/auth/domain"
	redisClient "github.com/redis/go-redis/v9"
)

type otpRepo struct {
	client *redisClient.Client
}

func NewOTPRepository(client *redisClient.Client) domain.OTPRepository {
	return &otpRepo{client: client}
}

func hashOTP(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func (r *otpRepo) SaveOTP(ctx context.Context, phone string, code string) error {
	key := fmt.Sprintf("otp:%s", phone)
	hashedCode := hashOTP(code)
	// 2 minutes TTL for OTP
	return r.client.Set(ctx, key, hashedCode, 2*time.Minute).Err()
}

func (r *otpRepo) VerifyOTP(ctx context.Context, phone string, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return false, nil // OTP not found or expired
		}
		return false, err
	}

	hashedCode := hashOTP(code)
	if val == hashedCode {
		// Valid code, delete it to prevent reuse
		r.client.Del(ctx, key)
		return true, nil
	}
	return false, nil
}

func (r *otpRepo) IncrementRateLimit(ctx context.Context, phone string) error {
	// Rate limit window: 5 minutes
	windowStr := time.Now().Truncate(5 * time.Minute).Format("2006-01-02-15:04")
	key := fmt.Sprintf("ratelimit:sms:%s:%s", phone, windowStr)

	pipe := r.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 5*time.Minute) // Expiry to clean up
	_, err := pipe.Exec(ctx)
	return err
}

func (r *otpRepo) CheckRateLimit(ctx context.Context, phone string) (int, error) {
	windowStr := time.Now().Truncate(5 * time.Minute).Format("2006-01-02-15:04")
	key := fmt.Sprintf("ratelimit:sms:%s:%s", phone, windowStr)

	val, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

func (r *otpRepo) IncrementIPRateLimit(ctx context.Context, ip string) error {
	// Rate limit window: 1 hour
	windowStr := time.Now().Truncate(time.Hour).Format("2006-01-02-15")
	key := fmt.Sprintf("ratelimit:ip:%s:%s", ip, windowStr)

	pipe := r.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *otpRepo) CheckIPRateLimit(ctx context.Context, ip string) (int, error) {
	windowStr := time.Now().Truncate(time.Hour).Format("2006-01-02-15")
	key := fmt.Sprintf("ratelimit:ip:%s:%s", ip, windowStr)

	val, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

func (r *otpRepo) IsPhoneLocked(ctx context.Context, phone string) (bool, error) {
	key := fmt.Sprintf("block:phone:%s", phone)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redisClient.Nil) {
			return false, nil // Not locked
		}
		return false, err
	}
	return val == "1", nil
}

func (r *otpRepo) IncrementFailedAttempts(ctx context.Context, phone string) (int, error) {
	key := fmt.Sprintf("fails:phone:%s", phone)

	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int(incr.Val()), nil
}

func (r *otpRepo) LockPhone(ctx context.Context, phone string, duration time.Duration) error {
	key := fmt.Sprintf("block:phone:%s", phone)
	return r.client.Set(ctx, key, "1", duration).Err()
}

func (r *otpRepo) ResetFailedAttempts(ctx context.Context, phone string) error {
	key := fmt.Sprintf("fails:phone:%s", phone)
	return r.client.Del(ctx, key).Err()
}
