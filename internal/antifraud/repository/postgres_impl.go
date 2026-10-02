package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/antifraud/domain"
	"github.com/hextec/yolyolakay-backend/internal/antifraud/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAntiFraudRepo struct {
	db      *pgxpool.Pool
	queries postgres.Querier
}

func NewPostgresAntiFraudRepository(db *pgxpool.Pool) domain.AntiFraudRepository {
	return &postgresAntiFraudRepo{
		db:      db,
		queries: postgres.New(db),
	}
}

func (r *postgresAntiFraudRepo) InitTrustScore(ctx context.Context, userID uuid.UUID, initialScore float64) (*domain.TrustScore, error) {
	pgUserID := database.UUIDToPg(userID)
	pgScore := database.Float64ToNumeric(initialScore)

	row, err := r.queries.InitTrustScore(ctx, postgres.InitTrustScoreParams{
		UserID: pgUserID,
		Score:  pgScore,
	})
	if err != nil {
		return nil, err
	}

	scoreFloat, _ := row.Score.Float64Value()
	return &domain.TrustScore{
		UserID:    userID,
		Score:     scoreFloat.Float64,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *postgresAntiFraudRepo) GetTrustScore(ctx context.Context, userID uuid.UUID) (*domain.TrustScore, error) {
	pgUserID := database.UUIDToPg(userID)
	row, err := r.queries.GetTrustScore(ctx, pgUserID)
	if err != nil {
		return nil, err
	}

	scoreFloat, _ := row.Score.Float64Value()
	return &domain.TrustScore{
		UserID:    userID,
		Score:     scoreFloat.Float64,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *postgresAntiFraudRepo) UpdateTrustScore(ctx context.Context, userID uuid.UUID, newScore float64) (*domain.TrustScore, error) {
	pgUserID := database.UUIDToPg(userID)
	pgScore := database.Float64ToNumeric(newScore)

	row, err := r.queries.UpdateTrustScore(ctx, postgres.UpdateTrustScoreParams{
		UserID: pgUserID,
		Score:  pgScore,
	})
	if err != nil {
		return nil, err
	}

	scoreFloat, _ := row.Score.Float64Value()
	return &domain.TrustScore{
		UserID:    userID,
		Score:     scoreFloat.Float64,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *postgresAntiFraudRepo) CreateFraudEvent(ctx context.Context, event *domain.FraudEvent) (*domain.FraudEvent, error) {
	pgID := database.UUIDToPg(event.ID)
	pgUserID := database.UUIDToPg(event.UserID)

	desc := database.StringToPgText(event.Description)
	pgPenalty := database.Float64ToNumeric(event.PenaltyApplied)

	row, err := r.queries.CreateFraudEvent(ctx, postgres.CreateFraudEventParams{
		ID:             pgID,
		UserID:         pgUserID,
		EventType:      event.EventType,
		Description:    desc,
		PenaltyApplied: pgPenalty,
	})
	if err != nil {
		return nil, err
	}

	penaltyFloat, _ := row.PenaltyApplied.Float64Value()

	return &domain.FraudEvent{
		ID:             event.ID,
		UserID:         event.UserID,
		EventType:      row.EventType,
		Description:    row.Description.String,
		PenaltyApplied: penaltyFloat.Float64,
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}

func (r *postgresAntiFraudRepo) GetFraudEventsByUser(ctx context.Context, userID uuid.UUID) ([]domain.FraudEvent, error) {
	pgUserID := database.UUIDToPg(userID)
	rows, err := r.queries.GetFraudEventsByUser(ctx, pgUserID)
	if err != nil {
		return nil, err
	}

	var events []domain.FraudEvent
	for _, row := range rows {
		penaltyFloat, _ := row.PenaltyApplied.Float64Value()
		events = append(events, domain.FraudEvent{
			ID:             row.ID.Bytes,
			UserID:         row.UserID.Bytes,
			EventType:      row.EventType,
			Description:    row.Description.String,
			PenaltyApplied: penaltyFloat.Float64,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return events, nil
}
