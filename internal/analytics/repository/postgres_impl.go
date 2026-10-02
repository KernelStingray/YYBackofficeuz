package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hextec/yolyolakay-backend/internal/analytics/domain"
	"github.com/hextec/yolyolakay-backend/internal/analytics/repository/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAnalyticsRepo struct {
	db      *pgxpool.Pool
	queries postgres.Querier
}

func NewPostgresAnalyticsRepository(db *pgxpool.Pool) domain.AnalyticsRepository {
	return &postgresAnalyticsRepo{
		db:      db,
		queries: postgres.New(db),
	}
}

func (r *postgresAnalyticsRepo) LogEvent(ctx context.Context, event *domain.Event) error {
	pgID := pgtype.UUID{Bytes: event.ID, Valid: true}

	var pgUserID pgtype.UUID
	if event.UserID != nil {
		pgUserID = pgtype.UUID{Bytes: *event.UserID, Valid: true}
	}

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	_, err = r.queries.CreateEvent(ctx, postgres.CreateEventParams{
		ID:        pgID,
		EventType: event.EventType,
		UserID:    pgUserID,
		Payload:   payloadBytes,
	})
	return err
}

func (r *postgresAnalyticsRepo) GetEventStatsFiltered(ctx context.Context, fromStr, toStr, groupBy, eventType string) ([]domain.EventStats, error) {
	query := `SELECT event_type, COUNT(*) as event_count`
	if groupBy == "day" {
		query = `SELECT event_type, TO_CHAR(created_at, 'YYYY-MM-DD') as event_date, COUNT(*) as event_count`
	}

	query += ` FROM events WHERE 1=1`
	var args []interface{}
	argID := 1

	if fromStr != "" {
		query += ` AND created_at >= $` + fmt.Sprint(argID)
		args = append(args, fromStr)
		argID++
	}
	if toStr != "" {
		query += ` AND created_at <= $` + fmt.Sprint(argID)
		args = append(args, toStr)
		argID++
	}
	if eventType != "" {
		query += ` AND event_type = $` + fmt.Sprint(argID)
		args = append(args, eventType)
		argID++
	}

	query += ` GROUP BY event_type`
	if groupBy == "day" {
		query += `, TO_CHAR(created_at, 'YYYY-MM-DD') ORDER BY event_date ASC, event_type`
	} else {
		query += ` ORDER BY event_type`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []domain.EventStats{}
	for rows.Next() {
		var stat domain.EventStats
		if groupBy == "day" {
			if err := rows.Scan(&stat.EventType, &stat.Date, &stat.EventCount); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&stat.EventType, &stat.EventCount); err != nil {
				return nil, err
			}
		}
		stats = append(stats, stat)
	}
	return stats, nil
}
