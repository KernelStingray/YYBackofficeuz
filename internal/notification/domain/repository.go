package domain

import (
	"context"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	CreateMessage(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, rideID *uuid.UUID) ([]Message, error)
	MarkMessageAsRead(ctx context.Context, messageID, receiverID uuid.UUID) error
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*Message, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]ChatListItem, error)
	SaveSystemNotification(ctx context.Context, notif *NotificationHistoryItem) error
	GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]NotificationHistoryItem, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
}

type DeviceTokenRepository interface {
	SaveDeviceToken(ctx context.Context, token *DeviceToken) error
	GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]*DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, userID, sessionID uuid.UUID) error
}
