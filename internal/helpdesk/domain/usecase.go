package domain

import (
	"context"

	"github.com/google/uuid"
)

type HelpdeskUseCase interface {
	CreateTicket(ctx context.Context, userID uuid.UUID, subject, description string) (*Ticket, error)
	GetUserTickets(ctx context.Context, userID uuid.UUID) ([]Ticket, error)
	ResolveTicket(ctx context.Context, ticketID uuid.UUID, resolution string) (*Ticket, error)
}
