package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"golang.org/x/crypto/bcrypt"
)

type backofficeUseCase struct {
	repo  domain.BackofficeRepository
	cache *cache.Client
}

func NewBackofficeUseCase(repo domain.BackofficeRepository, cacheClient *cache.Client) domain.BackofficeUseCase {
	return &backofficeUseCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (uc *backofficeUseCase) Login(ctx context.Context, login, password string) (*domain.Admin, error) {
	rdb := uc.cache.GetRawClient()
	lockKey := "backoffice:lock:" + login
	failsKey := "backoffice:fails:" + login

	// Check if account is locked
	locked, _ := rdb.Exists(ctx, lockKey).Result()
	if locked > 0 {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	admin, err := uc.repo.GetAdminByLogin(ctx, login)
	if err != nil {
		uc.incrementFailedAttempts(ctx, login)
		return nil, errors.New("invalid credentials") // do not leak 'user not found'
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
	if err != nil {
		uc.incrementFailedAttempts(ctx, login)
		return nil, errors.New("invalid credentials")
	}

	// Reset failed attempts upon successful login
	rdb.Del(ctx, failsKey)

	return admin, nil
}

func (uc *backofficeUseCase) GetAdmins(ctx context.Context) ([]*domain.Admin, error) {
	admins, err := uc.repo.GetAdmins(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range admins {
		a.PasswordHash = "" // never expose password hash
	}
	return admins, nil
}

func (uc *backofficeUseCase) incrementFailedAttempts(ctx context.Context, login string) {
	rdb := uc.cache.GetRawClient()
	failsKey := "backoffice:fails:" + login
	lockKey := "backoffice:lock:" + login

	fails, _ := rdb.Incr(ctx, failsKey).Result()
	if fails == 1 {
		rdb.Expire(ctx, failsKey, 15*time.Minute)
	}

	if fails >= 5 {
		// Lock account for 15 minutes
		rdb.Set(ctx, lockKey, "1", 15*time.Minute)
		rdb.Del(ctx, failsKey)
	}
}

func (uc *backofficeUseCase) SetRemoteConfig(ctx context.Context, key string, value map[string]interface{}) error {
	if key == "" {
		return errors.New("config key cannot be empty")
	}
	if value == nil {
		return errors.New("config value cannot be empty")
	}

	config := &domain.RemoteConfig{
		Key:   key,
		Value: value,
	}

	err := uc.repo.UpsertRemoteConfig(ctx, config)
	if err != nil {
		return err
	}

	// Invalidate cache
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("config:%s", key)
		_ = uc.cache.Delete(ctx, cacheKey)
	}

	return nil
}

func (uc *backofficeUseCase) GetRemoteConfig(ctx context.Context, key string) (*domain.RemoteConfig, error) {
	if key == "" {
		return nil, errors.New("config key cannot be empty")
	}

	cacheKey := fmt.Sprintf("config:%s", key)
	var cachedConfig domain.RemoteConfig

	// Try getting from cache first
	if uc.cache != nil {
		err := uc.cache.Get(ctx, cacheKey, &cachedConfig)
		if err == nil {
			slog.Info("[BackofficeUseCase] RemoteConfig retrieved from cache", "cacheKey", cacheKey)
			return &cachedConfig, nil
		}
	}

	// If cache miss, get from repo
	config, err := uc.repo.GetRemoteConfig(ctx, key)
	if err != nil {
		return nil, err
	}

	// Save to cache
	if uc.cache != nil {
		_ = uc.cache.Set(ctx, cacheKey, config, 1*time.Hour) // Cache config for 1 hour
	}

	return config, nil
}

func (uc *backofficeUseCase) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return errors.New("audit log cannot be nil")
	}

	log.CreatedAt = time.Now()
	return uc.repo.CreateAuditLog(ctx, log)
}
