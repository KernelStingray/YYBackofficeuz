package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/hextec/yolyolakay-backend/internal/notification/i18n"
	userPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var (
	wsActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ws_active_connections",
			Help: "Current number of active WebSocket connections in Notification Service",
		},
	)
)

func init() {
	prometheus.MustRegister(wsActiveConnections)
}

type userSubscriptionGroup struct {
	pubsub *redis.PubSub
	chans  map[chan domain.WSEvent]struct{}
}

type notificationUseCase struct {
	repo        domain.NotificationRepository
	tokenRepo   domain.DeviceTokenRepository
	redisClient *redis.Client
	publisher   domain.EventPublisher
	pushSender  domain.PushSender
	userClient  userPb.UserServiceClient

	mu          sync.RWMutex
	connections map[uuid.UUID]*userSubscriptionGroup
	translator  *i18n.Translator
}

func NewNotificationUseCase(
	repo domain.NotificationRepository,
	tokenRepo domain.DeviceTokenRepository,
	redisClient *redis.Client,
	publisher domain.EventPublisher,
	pushSender domain.PushSender,
	userClient userPb.UserServiceClient,
) domain.NotificationUseCase {
	return &notificationUseCase{
		repo:        repo,
		tokenRepo:   tokenRepo,
		redisClient: redisClient,
		publisher:   publisher,
		pushSender:  pushSender,
		userClient:  userClient,
		connections: make(map[uuid.UUID]*userSubscriptionGroup),
		translator:  i18n.NewTranslator(),
	}
}


func (u *notificationUseCase) SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, rideID *uuid.UUID, replyToID *uuid.UUID, content string) (*domain.Message, error) {
	msg := &domain.Message{
		ID:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		RideID:     rideID,
		ReplyToID:  replyToID,
		Content:    content,
		Status:     domain.MessageStatusSent,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// 1. Publish to message broker for async save
	msgJSON, _ := json.Marshal(msg)
	if u.publisher != nil {
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "chat.message.saved", msgJSON); err != nil {
			slog.Error("Failed to publish chat message", "error", err)
		}
	}

	// 2. Publish to Redis Pub/Sub for realtime delivery
	wsEvent := domain.WSEvent{
		Event: "new_message",
		Data:  msg,
	}
	wsEventJSON, _ := json.Marshal(wsEvent)
	channelName := fmt.Sprintf("channel:user:%s", receiverID)
	subscribers := u.redisClient.Publish(ctx, channelName, wsEventJSON).Val()

	// Also publish to sender's channel so their frontend (and other devices) receives the confirmation
	senderChannel := fmt.Sprintf("channel:user:%s", senderID)
	u.redisClient.Publish(ctx, senderChannel, wsEventJSON)

	if subscribers > 0 {
		msg.Status = domain.MessageStatusDelivered
	}

	// 3. Send Push notification to receiver's registered mobile device tokens so that backgrounded/locked devices receive the banner
	tokens, err := u.GetDeviceTokens(ctx, receiverID)
	if err == nil && len(tokens) > 0 {
		var senderName string
		if u.userClient != nil {
			// Try to get sender's name
			senderProfile, err := u.userClient.GetPassengerProfile(ctx, &userPb.GetPassengerProfileRequest{PassengerId: senderID.String()})
			if err == nil && senderProfile != nil {
				senderName = senderProfile.Name
			}
		}
		for _, token := range tokens {
			title := senderName
			if title == "" {
				if token.Language == "uz" {
					title = "Yangi xabar"
				} else if token.Language == "en" {
					title = "New message"
				} else {
					title = "Новое сообщение"
				}
			}

			if u.pushSender != nil {
				data := map[string]string{
					"type":      "new_message",
					"peer_id":   senderID.String(),
					"peer_name": title,
				}
				if msg.RideID != nil {
					data["ride_id"] = msg.RideID.String()
				}
				if sendErr := u.pushSender.SendPush(ctx, token.FCMToken, title, msg.Content, data); sendErr != nil {
					slog.Error("Failed to send push notification", "error", sendErr)
				} else {
					slog.Info(fmt.Sprintf("[PUSH] Sent to %s", token.FCMToken))
				}
			} else {
				slog.Info(fmt.Sprintf("[PUSH MOCK] Sending push notification to token %s: %s", token.FCMToken, msg.Content))
			}
		}
	} else if subscribers == 0 {
		slog.Info(fmt.Sprintf("[SMS MOCK] Skipping SMS to User %s (Offline without FCM)", receiverID))
	}

	return msg, nil
}

func (u *notificationUseCase) SaveMessageToDB(ctx context.Context, msg *domain.Message) error {
	return u.repo.CreateMessage(ctx, msg)
}

func (u *notificationUseCase) GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, rideID *uuid.UUID) ([]domain.Message, error) {
	return u.repo.GetChatHistory(ctx, userID1, userID2, rideID)
}

func (u *notificationUseCase) GetUserChats(ctx context.Context, userID uuid.UUID) ([]domain.ChatListItem, error) {
	return u.repo.GetUserChats(ctx, userID)
}

func (u *notificationUseCase) MarkMessageAsRead(ctx context.Context, messageID, receiverID uuid.UUID) error {
	// Send WS event to sender
	if msg, err := u.repo.GetMessageByID(ctx, messageID); err == nil {
		data := map[string]interface{}{
			"message_ids": []string{messageID.String()},
		}
		_ = u.PushWSEvent(ctx, msg.SenderID, "messages_read", data)
	}
	return u.repo.MarkMessageAsRead(ctx, messageID, receiverID)
}

func (u *notificationUseCase) RegisterDevice(ctx context.Context, userID, sessionID uuid.UUID, fcmToken, platform, language string) error {
	token := &domain.DeviceToken{
		UserID:    userID,
		SessionID: sessionID,
		FCMToken:  fcmToken,
		Platform:  platform,
		Language:  language,
	}

	return u.tokenRepo.SaveDeviceToken(ctx, token)
}

func (u *notificationUseCase) GetDeviceTokens(ctx context.Context, userID uuid.UUID) ([]*domain.DeviceToken, error) {
	return u.tokenRepo.GetDeviceTokens(ctx, userID)
}

func (u *notificationUseCase) DeleteDeviceToken(ctx context.Context, userID, sessionID uuid.UUID) error {
	return u.tokenRepo.DeleteDeviceToken(ctx, userID, sessionID)
}

func (u *notificationUseCase) Subscribe(userID uuid.UUID) chan domain.WSEvent {
	u.mu.Lock()
	defer u.mu.Unlock()

	wsActiveConnections.Inc()
	ch := make(chan domain.WSEvent, 100) // Buffered channel for this connection

	group, exists := u.connections[userID]
	if !exists {
		channelName := fmt.Sprintf("channel:user:%s", userID)
		pubsub := u.redisClient.Subscribe(context.Background(), channelName)
		group = &userSubscriptionGroup{
			pubsub: pubsub,
			chans:  make(map[chan domain.WSEvent]struct{}),
		}
		u.connections[userID] = group

		// Set online status in Redis with a TTL of 120 seconds
		u.redisClient.SetEx(context.Background(), fmt.Sprintf("user:online:%s", userID), "online", 120*time.Second)

		// Single Redis listener for this user, broadcasting to ALL active connection channels
		go func(g *userSubscriptionGroup) {
			defer g.pubsub.Close()
			chRedis := g.pubsub.Channel()
			for msg := range chRedis {
				var wsEvent domain.WSEvent
				if err := json.Unmarshal([]byte(msg.Payload), &wsEvent); err == nil {
					u.mu.RLock()
					for c := range g.chans {
						select {
						case c <- wsEvent:
						default:
						}
					}
					u.mu.RUnlock()
				}
			}
		}(group)
	}

	group.chans[ch] = struct{}{}
	return ch
}

func (u *notificationUseCase) Unsubscribe(userID uuid.UUID, ch chan domain.WSEvent) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if group, exists := u.connections[userID]; exists {
		if _, ok := group.chans[ch]; ok {
			wsActiveConnections.Dec()
			delete(group.chans, ch)
			close(ch)
			if len(group.chans) == 0 {
				_ = group.pubsub.Close()
				delete(u.connections, userID)
				// Remove online status
				u.redisClient.Del(context.Background(), fmt.Sprintf("user:online:%s", userID))
			}
		}
	}
}

func (u *notificationUseCase) RefreshOnlineStatus(ctx context.Context, userID uuid.UUID) error {
	return u.redisClient.Expire(ctx, fmt.Sprintf("user:online:%s", userID), 120*time.Second).Err()
}

func (u *notificationUseCase) GetOnlineStatuses(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]bool{}, nil
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = fmt.Sprintf("user:online:%s", id)
	}

	// MGet is O(N) where N is number of keys, but much faster than N separate GETs
	vals, err := u.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]bool, len(userIDs))
	for i, val := range vals {
		if val != nil {
			result[userIDs[i]] = true
		} else {
			result[userIDs[i]] = false
		}
	}

	return result, nil
}

func (u *notificationUseCase) PushWSEvent(ctx context.Context, userID uuid.UUID, event string, data interface{}) error {
	wsEvent := domain.WSEvent{
		Event: event,
		Data:  data,
	}
	wsEventJSON, err := json.Marshal(wsEvent)
	if err != nil {
		return err
	}
	channelName := fmt.Sprintf("channel:user:%s", userID)
	return u.redisClient.Publish(ctx, channelName, wsEventJSON).Err()
}

func (u *notificationUseCase) SubscribeToRide(userID, rideID uuid.UUID) error {
	u.mu.RLock()
	group, exists := u.connections[userID]
	u.mu.RUnlock()

	if !exists || group == nil {
		return fmt.Errorf("user not connected to websockets")
	}

	channelName := fmt.Sprintf("channel:ride:%s:location", rideID)
	return group.pubsub.Subscribe(context.Background(), channelName)
}

func (u *notificationUseCase) BroadcastLocation(ctx context.Context, rideID uuid.UUID, lat, lon float64) error {
	wsEvent := domain.WSEvent{
		Event: "driver_location",
		Data: map[string]interface{}{
			"ride_id": rideID.String(),
			"lat":     lat,
			"lon":     lon,
		},
	}
	wsEventJSON, err := json.Marshal(wsEvent)
	if err != nil {
		return err
	}
	channelName := fmt.Sprintf("channel:ride:%s:location", rideID)
	return u.redisClient.Publish(ctx, channelName, wsEventJSON).Err()
}

func (u *notificationUseCase) GetUserLocale(ctx context.Context, userID uuid.UUID) string {
	if u.userClient != nil {
		req := &userPb.GetUserLocaleRequest{UserId: userID.String()}
		if res, err := u.userClient.GetUserLocale(ctx, req); err == nil && res.Locale != "" {
			return res.Locale
		}
	}

	if u.tokenRepo != nil {
		if tokens, err := u.tokenRepo.GetDeviceTokens(ctx, userID); err == nil && len(tokens) > 0 {
			for _, t := range tokens {
				if t.Language != "" {
					return t.Language
				}
			}
		}
	}

	return "uz"
}

func (u *notificationUseCase) SendSystemPush(ctx context.Context, userID uuid.UUID, title, body, notifType string) error {
	// 1. Save to DB
	if err := u.SaveSystemNotification(ctx, userID, title, body, notifType); err != nil {
		slog.Error("Failed to save system notification", "error", err)
		// we continue to push anyway
	}

	// 2. Prepare payload for WS
	payloadMsg := &domain.Message{
		ID:         uuid.New(),
		SenderID:   uuid.Nil,
		ReceiverID: userID,
		Content:    body,
		Status:     domain.MessageStatusSent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 3. Try WS
	u.mu.RLock()
	group, ok := u.connections[userID]
	u.mu.RUnlock()

	if ok && group != nil {
		wsEvent := domain.WSEvent{
			Event: "system_push",
			Data:  payloadMsg,
		}
		u.mu.RLock()
		for c := range group.chans {
			select {
			case c <- wsEvent:
			default:
			}
		}
		u.mu.RUnlock()
		return nil
	} else {
		// Fallback to push notification
		tokens, err := u.GetDeviceTokens(ctx, userID)
		if err == nil && len(tokens) > 0 {
			for _, token := range tokens {
				if u.pushSender != nil {
					data := map[string]string{
						"type": notifType,
					}
					if err := u.pushSender.SendPush(ctx, token.FCMToken, title, body, data); err != nil {
						slog.Error("Failed to send push", "token", token.FCMToken, "error", err)
					} else {
						slog.Info("Successfully sent push", "token", token.FCMToken, "type", notifType)
					}
				} else {
					slog.Info(fmt.Sprintf("[PUSH MOCK] Sending system push to token %s: %s - %s (type: %s)", token.FCMToken, title, body, notifType))
				}
			}
		} else {
			slog.Info(fmt.Sprintf("[SMS MOCK] Sending SMS to User %s: %s - %s", userID, title, body))
		}
	}
	return nil
}

func (u *notificationUseCase) SendSystemPushWithData(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	notifType := data["type"]
	if notifType == "" {
		notifType = "system"
	}

	var rideID, bookingID, parcelID *uuid.UUID
	if rIDStr, ok := data["ride_id"]; ok && rIDStr != "" {
		if rID, err := uuid.Parse(rIDStr); err == nil {
			rideID = &rID
		}
	}
	if bIDStr, ok := data["booking_id"]; ok && bIDStr != "" {
		if bID, err := uuid.Parse(bIDStr); err == nil {
			bookingID = &bID
		}
	}
	if pIDStr, ok := data["parcel_id"]; ok && pIDStr != "" {
		if pID, err := uuid.Parse(pIDStr); err == nil {
			parcelID = &pID
		}
	}

	notif := &domain.NotificationHistoryItem{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		Type:      notifType,
		RideID:    rideID,
		BookingID: bookingID,
		ParcelID:  parcelID,
		Data:      data,
	}
	if err := u.repo.SaveSystemNotification(ctx, notif); err != nil {
		slog.Error("Failed to save system notification to database", "error", err, "user_id", userID, "type", notifType)
	}

	tokens, err := u.GetDeviceTokens(ctx, userID)
	if err != nil {
		slog.Error("Failed to get device tokens", "error", err, "user_id", userID)
		return err
	}
	if len(tokens) == 0 {
		slog.Warn("No device tokens found for user push", "user_id", userID, "type", notifType)
		return nil
	}
	for _, token := range tokens {
		if u.pushSender != nil {
			if err := u.pushSender.SendPush(ctx, token.FCMToken, title, body, data); err != nil {
				slog.Error("Failed to send push with data", "token", token.FCMToken, "error", err)
			} else {
				slog.Info("Successfully sent push with data", "token", token.FCMToken, "type", notifType)
			}
		} else {
			slog.Info(fmt.Sprintf("[PUSH MOCK] Sending system push with data to token %s: %s - %s (type: %s)", token.FCMToken, title, body, notifType))
		}
	}
	return nil
}

func (u *notificationUseCase) SaveSystemNotification(ctx context.Context, userID uuid.UUID, title, body, notifType string) error {
	notif := &domain.NotificationHistoryItem{
		ID:     uuid.New(),
		UserID: userID,
		Title:  title,
		Body:   body,
		Type:   notifType,
	}
	return u.repo.SaveSystemNotification(ctx, notif)
}

func (u *notificationUseCase) GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.NotificationHistoryItem, error) {
	return u.repo.GetNotificationHistory(ctx, userID, limit, offset)
}

func (u *notificationUseCase) MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return u.repo.MarkNotificationAsRead(ctx, notificationID, userID)
}

func (u *notificationUseCase) VerifyRideDriver(ctx context.Context, rideID, userID uuid.UUID) bool {
	key := fmt.Sprintf("ride:%s:driver", rideID.String())
	driverIDStr, err := u.redisClient.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return driverIDStr == userID.String()
}

func (u *notificationUseCase) VerifyRideParticipant(ctx context.Context, rideID, userID uuid.UUID) bool {
	if u.VerifyRideDriver(ctx, rideID, userID) {
		return true
	}
	key := fmt.Sprintf("ride:%s:passengers", rideID.String())
	isPassenger, err := u.redisClient.SIsMember(ctx, key, userID.String()).Result()
	if err != nil {
		return false
	}
	return isPassenger
}
