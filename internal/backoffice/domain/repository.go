package domain

import (
	"context"
)

type BackofficeRepository interface {
	CreateAdmin(ctx context.Context, admin *Admin) error
	GetAdminByLogin(ctx context.Context, login string) (*Admin, error)
	GetAdmins(ctx context.Context) ([]*Admin, error)
	UpsertRemoteConfig(ctx context.Context, config *RemoteConfig) error
	GetRemoteConfig(ctx context.Context, key string) (*RemoteConfig, error)
	CreateAuditLog(ctx context.Context, log *AuditLog) error
}
