package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/helpdesk/domain"
)

type helpdeskUseCase struct {
	repo domain.HelpdeskRepository
}

func NewHelpdeskUseCase(repo domain.HelpdeskRepository) domain.HelpdeskUseCase {
	return &helpdeskUseCase{
		repo: repo,
	}
}

func (uc *helpdeskUseCase) CreateTicket(ctx context.Context, userID uuid.UUID, subject, description string) (*domain.Ticket, error) {
	if subject == "" {
		return nil, errors.New("ticket subject cannot be empty")
	}
	if description == "" {
		return nil, errors.New("ticket description cannot be empty")
	}

	ticket := &domain.Ticket{
		ID:          uuid.New(),
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Status:      domain.TicketStatusOpen,
	}

	return uc.repo.CreateTicket(ctx, ticket)
}

func (uc *helpdeskUseCase) GetUserTickets(ctx context.Context, userID uuid.UUID) ([]domain.Ticket, error) {
	return uc.repo.GetUserTickets(ctx, userID)
}

func (uc *helpdeskUseCase) ResolveTicket(ctx context.Context, ticketID uuid.UUID, resolution string) (*domain.Ticket, error) {
	return uc.repo.UpdateTicketStatus(ctx, ticketID, domain.TicketStatusResolved, resolution)
}
