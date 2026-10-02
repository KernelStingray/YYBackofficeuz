package domain

import (
	"time"

	"github.com/google/uuid"
)

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "OPEN"
	TicketStatusInProgress TicketStatus = "IN_PROGRESS"
	TicketStatusResolved   TicketStatus = "RESOLVED"
	TicketStatusClosed     TicketStatus = "CLOSED"
)

type Ticket struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Subject     string       `json:"subject"`
	Description string       `json:"description"`
	Status      TicketStatus `json:"status" swaggertype:"string"`
	Resolution  *string      `json:"resolution,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
