package domain

import (
	"context"
)

type BackofficeUseCase interface {
	Login(ctx context.Context, login, password string) (*Admin, error)
	GetAdmins(ctx context.Context) ([]*Admin, error)
	SetRemoteConfig(ctx context.Context, key string, value map[string]interface{}) error
	GetRemoteConfig(ctx context.Context, key string) (*RemoteConfig, error)
	CreateAuditLog(ctx context.Context, log *AuditLog) error
}
