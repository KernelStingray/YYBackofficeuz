package domain

import (
	"context"

	"github.com/google/uuid"
)

type NotificationUseCase interface {
	SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, rideID *uuid.UUID, replyToID *uuid.UUID, content string) (*Message, error)
	SaveMessageToDB(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, rideID *uuid.UUID) ([]Message, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]ChatListItem, error)
	MarkMessageAsRead(ctx context.Context, messageID, receiverID uuid.UUID) error
	Subscribe(userID uuid.UUID) chan WSEvent
	Unsubscribe(userID uuid.UUID, ch chan WSEvent)
	RefreshOnlineStatus(ctx context.Context, userID uuid.UUID) error
	GetOnlineStatuses(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	PushWSEvent(ctx context.Context, userID uuid.UUID, event string, data interface{}) error
	RegisterDevice(ctx context.Context, userID, sessionID uuid.UUID, fcmToken, platform, language string) error
	GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]*DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, userID, sessionID uuid.UUID) error
	SendSystemPush(ctx context.Context, userID uuid.UUID, title, body, notifType string) error
	SendSystemPushWithData(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error
	SaveSystemNotification(ctx context.Context, userID uuid.UUID, title, body, notifType string) error
	GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]NotificationHistoryItem, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	SubscribeToRide(userID, rideID uuid.UUID) error
	BroadcastLocation(ctx context.Context, rideID uuid.UUID, lat, lon float64) error
	VerifyRideDriver(ctx context.Context, rideID, userID uuid.UUID) bool
	VerifyRideParticipant(ctx context.Context, rideID, userID uuid.UUID) bool
	GetUserLocale(ctx context.Context, userID uuid.UUID) string
}

// EventPublisher abstracts the message broker (e.g., RabbitMQ).
// UseCase depends on this interface instead of a concrete broker client.
type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// PushSender abstracts push notification delivery (e.g., Firebase Cloud Messaging).
// UseCase depends on this interface instead of a concrete FCM client.
type PushSender interface {
	SendPush(ctx context.Context, token string, title, body string, data map[string]string) error
}
