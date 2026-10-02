package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/helpdesk/domain"
	"github.com/hextec/yolyolakay-backend/internal/helpdesk/repository/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresHelpdeskRepo struct {
	db      *pgxpool.Pool
	queries postgres.Querier
}

func NewPostgresHelpdeskRepository(db *pgxpool.Pool) domain.HelpdeskRepository {
	return &postgresHelpdeskRepo{
		db:      db,
		queries: postgres.New(db),
	}
}

func (r *postgresHelpdeskRepo) CreateTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Ticket, error) {
	pgID := pgtype.UUID{Bytes: ticket.ID, Valid: true}
	pgUserID := pgtype.UUID{Bytes: ticket.UserID, Valid: true}

	row, err := r.queries.CreateTicket(ctx, postgres.CreateTicketParams{
		ID:          pgID,
		UserID:      pgUserID,
		Subject:     ticket.Subject,
		Description: ticket.Description,
		Status:      string(ticket.Status),
	})
	if err != nil {
		return nil, err
	}

	var resolution *string
	if row.Resolution.Valid {
		resStr := row.Resolution.String
		resolution = &resStr
	}

	return &domain.Ticket{
		ID:          row.ID.Bytes,
		UserID:      row.UserID.Bytes,
		Subject:     row.Subject,
		Description: row.Description,
		Status:      domain.TicketStatus(row.Status),
		Resolution:  resolution,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *postgresHelpdeskRepo) GetUserTickets(ctx context.Context, userID uuid.UUID) ([]domain.Ticket, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}

	rows, err := r.queries.GetUserTickets(ctx, pgUserID)
	if err != nil {
		return nil, err
	}

	var tickets []domain.Ticket
	for _, row := range rows {
		var resolution *string
		if row.Resolution.Valid {
			resStr := row.Resolution.String
			resolution = &resStr
		}

		tickets = append(tickets, domain.Ticket{
			ID:          row.ID.Bytes,
			UserID:      row.UserID.Bytes,
			Subject:     row.Subject,
			Description: row.Description,
			Status:      domain.TicketStatus(row.Status),
			Resolution:  resolution,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}

	return tickets, nil
}

func (r *postgresHelpdeskRepo) UpdateTicketStatus(ctx context.Context, ticketID uuid.UUID, status domain.TicketStatus, resolution string) (*domain.Ticket, error) {
	pgID := pgtype.UUID{Bytes: ticketID, Valid: true}

	pgResolution := pgtype.Text{String: resolution, Valid: resolution != ""}

	row, err := r.queries.UpdateTicketStatus(ctx, postgres.UpdateTicketStatusParams{
		ID:         pgID,
		Status:     string(status),
		Resolution: pgResolution,
	})
	if err != nil {
		return nil, err
	}

	var res *string
	if row.Resolution.Valid {
		resStr := row.Resolution.String
		res = &resStr
	}

	return &domain.Ticket{
		ID:          row.ID.Bytes,
		UserID:      row.UserID.Bytes,
		Subject:     row.Subject,
		Description: row.Description,
		Status:      domain.TicketStatus(row.Status),
		Resolution:  res,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}
