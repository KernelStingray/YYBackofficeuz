package repository

import (
	"context"
	"encoding/json"

	"github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresBackofficeRepo struct {
	db      *pgxpool.Pool
	queries postgres.Querier
}

func NewPostgresBackofficeRepository(db *pgxpool.Pool) domain.BackofficeRepository {
	return &postgresBackofficeRepo{
		db:      db,
		queries: postgres.New(db),
	}
}

func (r *postgresBackofficeRepo) CreateAdmin(ctx context.Context, admin *domain.Admin) error {
	pgID := database.UUIDToPg(admin.ID)

	pgAdmin, err := r.queries.CreateAdmin(ctx, postgres.CreateAdminParams{
		ID:           pgID,
		Login:        admin.Login,
		PasswordHash: admin.PasswordHash,
		Role:         admin.Role,
	})
	if err != nil {
		return err
	}

	admin.CreatedAt = pgAdmin.CreatedAt.Time
	return nil
}

func (r *postgresBackofficeRepo) GetAdminByLogin(ctx context.Context, login string) (*domain.Admin, error) {
	pgAdmin, err := r.queries.GetAdminByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	return &domain.Admin{
		ID:           pgAdmin.ID.Bytes,
		Login:        pgAdmin.Login,
		PasswordHash: pgAdmin.PasswordHash,
		Role:         pgAdmin.Role,
		CreatedAt:    pgAdmin.CreatedAt.Time,
	}, nil
}

func (r *postgresBackofficeRepo) GetAdmins(ctx context.Context) ([]*domain.Admin, error) {
	pgAdmins, err := r.queries.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}

	admins := make([]*domain.Admin, len(pgAdmins))
	for i, pgAdmin := range pgAdmins {
		admins[i] = &domain.Admin{
			ID:           pgAdmin.ID.Bytes,
			Login:        pgAdmin.Login,
			PasswordHash: pgAdmin.PasswordHash,
			Role:         pgAdmin.Role,
			CreatedAt:    pgAdmin.CreatedAt.Time,
		}
	}
	return admins, nil
}

func (r *postgresBackofficeRepo) UpsertRemoteConfig(ctx context.Context, config *domain.RemoteConfig) error {
	payloadBytes, err := json.Marshal(config.Value)
	if err != nil {
		return err
	}

	pgConfig, err := r.queries.UpsertRemoteConfig(ctx, postgres.UpsertRemoteConfigParams{
		Key:   config.Key,
		Value: payloadBytes,
	})
	if err != nil {
		return err
	}

	config.UpdatedAt = pgConfig.UpdatedAt.Time
	return nil
}

func (r *postgresBackofficeRepo) GetRemoteConfig(ctx context.Context, key string) (*domain.RemoteConfig, error) {
	pgConfig, err := r.queries.GetRemoteConfig(ctx, key)
	if err != nil {
		return nil, err
	}

	var value map[string]interface{}
	if err := json.Unmarshal(pgConfig.Value, &value); err != nil {
		return nil, err
	}

	return &domain.RemoteConfig{
		Key:       pgConfig.Key,
		Value:     value,
		UpdatedAt: pgConfig.UpdatedAt.Time,
	}, nil
}

func (r *postgresBackofficeRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	var payloadBytes []byte
	var err error
	if log.Payload != nil {
		payloadBytes, err = json.Marshal(log.Payload)
		if err != nil {
			return err
		}
	}

	var targetID pgtype.Text
	if log.TargetID != nil {
		targetID = database.StringToPgText(*log.TargetID)
	}

	var ipAddr pgtype.Text
	if log.IPAddress != nil {
		ipAddr = database.StringToPgText(*log.IPAddress)
	}

	err = r.queries.InsertAuditLog(ctx, postgres.InsertAuditLogParams{
		ID:        database.UUIDToPg(log.ID),
		AdminID:   database.UUIDToPg(log.AdminID),
		Action:    log.Action,
		Entity:    log.Entity,
		TargetID:  targetID,
		Payload:   payloadBytes,
		IpAddress: ipAddr,
	})
	if err != nil {
		return err
	}
	return nil
}
