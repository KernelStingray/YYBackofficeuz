package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"` // nullable for anonymous events
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

type EventStats struct {
	Date       string `json:"date,omitempty"`
	EventType  string `json:"event_type"`
	EventCount int64  `json:"event_count"`
}
