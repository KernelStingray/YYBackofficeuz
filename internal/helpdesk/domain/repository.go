package domain

import (
	"context"

	"github.com/google/uuid"
)

type HelpdeskRepository interface {
	CreateTicket(ctx context.Context, ticket *Ticket) (*Ticket, error)
	GetUserTickets(ctx context.Context, userID uuid.UUID) ([]Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID uuid.UUID, status TicketStatus, resolution string) (*Ticket, error)
}
