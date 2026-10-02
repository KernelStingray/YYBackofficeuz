package domain

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type RemoteConfig struct {
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type AuditLog struct {
	ID        uuid.UUID              `json:"id"`
	AdminID   uuid.UUID              `json:"admin_id"`
	Action    string                 `json:"action"`
	Entity    string                 `json:"entity"`
	TargetID  *string                `json:"target_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	IPAddress *string                `json:"ip_address,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}
