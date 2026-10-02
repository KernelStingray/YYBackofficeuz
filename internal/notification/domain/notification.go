package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type Message struct {
	ID         uuid.UUID     `json:"id"`
	SenderID   uuid.UUID     `json:"sender_id"`
	ReceiverID uuid.UUID     `json:"receiver_id"`
	RideID     *uuid.UUID    `json:"ride_id,omitempty"`
	ReplyToID  *uuid.UUID    `json:"reply_to_id,omitempty"`
	Content    string        `json:"content"`
	Status     MessageStatus `json:"status" swaggertype:"string"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type DeviceToken struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	FCMToken  string    `json:"fcm_token"`
	Platform  string    `json:"platform"`
	Language  string    `json:"language"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationHistoryItem struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string            `json:"type"`
	IsRead    bool              `json:"is_read"`
	RideID    *uuid.UUID        `json:"ride_id,omitempty"`
	BookingID *uuid.UUID        `json:"booking_id,omitempty"`
	ParcelID  *uuid.UUID        `json:"parcel_id,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type WSEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type ChatListItem struct {
	PartnerID   uuid.UUID `json:"partner_id"`
	LastMessage *Message  `json:"last_message"`
	UnreadCount int32     `json:"unread_count"`
}
