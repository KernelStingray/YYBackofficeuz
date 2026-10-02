package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/usecase"
	"github.com/stretchr/testify/assert"
)

type mockBackofficeRepo struct {
	admins []*domain.Admin
}

func (m *mockBackofficeRepo) CreateAdmin(ctx context.Context, admin *domain.Admin) error {
	m.admins = append(m.admins, admin)
	return nil
}

func (m *mockBackofficeRepo) GetAdminByLogin(ctx context.Context, login string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Login == login {
			return a, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockBackofficeRepo) GetAdmins(ctx context.Context) ([]*domain.Admin, error) {
	result := make([]*domain.Admin, len(m.admins))
	for i, a := range m.admins {
		copyA := *a
		result[i] = &copyA
	}
	return result, nil
}

func (m *mockBackofficeRepo) UpsertRemoteConfig(ctx context.Context, config *domain.RemoteConfig) error {
	return nil
}

func (m *mockBackofficeRepo) GetRemoteConfig(ctx context.Context, key string) (*domain.RemoteConfig, error) {
	return nil, nil
}

func (m *mockBackofficeRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return nil
}

func TestBackofficeUseCase_GetAdmins(t *testing.T) {
	adminID := uuid.New()
	mockRepo := &mockBackofficeRepo{
		admins: []*domain.Admin{
			{
				ID:           adminID,
				Login:        "KernelStingray",
				PasswordHash: "$2a$10$secretHashThatMustBeStripped",
				Role:         "admin",
				CreatedAt:    time.Now(),
			},
		},
	}

	uc := usecase.NewBackofficeUseCase(mockRepo, nil)
	admins, err := uc.GetAdmins(context.Background())

	assert.NoError(t, err)
	assert.Len(t, admins, 1)
	assert.Equal(t, "KernelStingray", admins[0].Login)
	assert.Equal(t, "", admins[0].PasswordHash, "PasswordHash must be cleared")
}
