package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/rating/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)


type PostgresRatingRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRatingRepository(db *pgxpool.Pool) domain.RatingRepository {
	return &PostgresRatingRepository{db: db}
}

func (r *PostgresRatingRepository) Save(ctx context.Context, feedback *domain.Feedback) error {
	query := `
		INSERT INTO feedbacks (id, ride_id, reviewer_id, reviewee_id, score, comment, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	tags := feedback.Tags
	if tags == nil {
		tags = []string{}
	}

	_, err := r.db.Exec(ctx, query,
		feedback.ID,
		feedback.RideID,
		feedback.ReviewerID,
		feedback.RevieweeID,
		feedback.Score,
		feedback.Comment,
		tags,
		feedback.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	return nil
}

func (r *PostgresRatingRepository) GetFeedbacksByRevieweeID(ctx context.Context, revieweeID uuid.UUID, limit, offset int, withComment bool) ([]domain.Feedback, int64, float64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	commentFilter := ""
	if withComment {
		commentFilter = " AND comment IS NOT NULL AND TRIM(comment) != ''"
	}

	// 1. Get total and average
	aggQuery := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(AVG(score), 0.0)
		FROM feedbacks
		WHERE reviewee_id = $1%s
	`, commentFilter)

	var total int64
	var avg float64
	err := r.db.QueryRow(ctx, aggQuery, revieweeID).Scan(&total, &avg)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to aggregate feedbacks: %w", err)
	}

	if total == 0 {
		return []domain.Feedback{}, 0, 0, nil
	}

	// 2. Get paginated items
	query := fmt.Sprintf(`
		SELECT id, ride_id, reviewer_id, reviewee_id, score, COALESCE(comment, ''), COALESCE(tags, '{}'), created_at
		FROM feedbacks
		WHERE reviewee_id = $1%s
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, commentFilter)

	rows, err := r.db.Query(ctx, query, revieweeID, limit, offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to query feedbacks: %w", err)
	}
	defer rows.Close()

	var items []domain.Feedback
	for rows.Next() {
		var f domain.Feedback
		var tags []string
		if err := rows.Scan(&f.ID, &f.RideID, &f.ReviewerID, &f.RevieweeID, &f.Score, &f.Comment, &tags, &f.CreatedAt); err != nil {
			return nil, 0, 0, fmt.Errorf("failed to scan feedback: %w", err)
		}
		if tags == nil {
			tags = []string{}
		}
		f.Tags = tags
		items = append(items, f)
	}

	if items == nil {
		items = []domain.Feedback{}
	}

	return items, total, avg, nil
}

