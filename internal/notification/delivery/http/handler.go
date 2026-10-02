package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	domain "github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 40 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

type Config struct {
	VapidPublicKey   string
	AllowedWSOrigins []string
}

type Handler struct {
	usecase  domain.NotificationUseCase
	upgrader websocket.Upgrader
	config   Config
}

func NewHandler(usecase domain.NotificationUseCase, cfg Config) *Handler {
	allowedOriginsMap := map[string]bool{
		"http://localhost:3000":                 true,
		"http://localhost:3001":                 true,
		"http://localhost:3002":                 true,
		"http://localhost:3003":                 true,
		"http://localhost:8080":                 true,
		"http://localhost:8086":                 true,
		"https://yolyolakay.com":                true,
		"https://www.yolyolakay.com":            true,
		"https://app.yolyolakay.com":            true,
		"https://api.yolyolakay.com":            true,
		"https://yolyolakay.uz":                 true,
		"https://www.yolyolakay.uz":             true,
		"https://yolyolakay.co.uz":              true,
		"https://www.yolyolakay.co.uz":          true,
		"http://yolyolakay.co.uz":               true,
		"http://www.yolyolakay.co.uz":           true,
		"https://yolyolakay-frontend.vercel.app": true,
	}
	for _, o := range cfg.AllowedWSOrigins {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			allowedOriginsMap[trimmed] = true
		}
	}

	return &Handler{
		usecase: usecase,
		config:  cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" || origin == "file://" {
					return true
				}
				return allowedOriginsMap[origin]
			},
		},
	}
}

type registerTokenRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
	Platform string `json:"platform"`
	Language string `json:"language"`
}

// RegisterToken godoc
// @Summary Регистрация FCM токена
// @Description Сохраняет токен устройства для пуш-уведомлений
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body registerTokenRequest true "Данные токена"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/tokens [post]
func (h *Handler) RegisterToken(c *gin.Context) {
	var req registerTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	sessionIDStr := c.GetString("session_id")
	var sessionID uuid.UUID
	if sessionIDStr != "" {
		sessionID, _ = uuid.Parse(sessionIDStr)
	}

	lang := req.Language
	if lang == "" {
		lang = "ru" // Default fallback
	}

	if err := h.usecase.RegisterDevice(c.Request.Context(), userID, sessionID, req.FCMToken, req.Platform, lang); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token registered successfully"})
}

// GetHistory godoc
// @Summary Получить историю переписки
// @Description Возвращает историю сообщений между двумя пользователями
// @Tags notifications
// @Accept json
// @Produce json
// @Param peer_id query string false "Peer ID (ID собеседника, обязателен: peer_id, user_id_2 или user_id_1)"
// @Param user_id_2 query string false "User 2 ID (альтернативный ID собеседника)"
// @Param user_id_1 query string false "User 1 ID (альтернативный ID собеседника)"
// @Param ride_id query string false "Ride ID (опционально)"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/notifications/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	currentUserIDStr := c.GetString("user_id")
	if currentUserIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	currentUserID, err := uuid.Parse(currentUserIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	peerIDStr := c.Query("peer_id")
	u1Str := c.Query("user_id_1")
	u2Str := c.Query("user_id_2")

	// Determine peer ID symmetrically:
	// Find the ID among query params that is not currentUserID
	if peerIDStr == "" || peerIDStr == currentUserIDStr {
		if u1Str != "" && u1Str != currentUserIDStr {
			peerIDStr = u1Str
		} else if u2Str != "" && u2Str != currentUserIDStr {
			peerIDStr = u2Str
		} else if u2Str != "" {
			peerIDStr = u2Str
		} else if u1Str != "" {
			peerIDStr = u1Str
		}
	}

	if peerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "peer_id or user_id_2 query parameter is required to identify chat partner"})
		return
	}

	peerID, err := uuid.Parse(peerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer user ID format"})
		return
	}

	var rideID *uuid.UUID
	rID := c.Query("ride_id")
	if rID != "" {
		if id, err := uuid.Parse(rID); err == nil {
			rideID = &id
		}
	}

	msgs, err := h.usecase.GetChatHistory(c.Request.Context(), currentUserID, peerID, rideID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if msgs == nil {
		msgs = []domain.Message{}
	}

	c.JSON(http.StatusOK, msgs)
}

// GetChatMessages godoc
// @Summary Получить историю сообщений конкретного чата
// @Tags chats
// @Produce json
// @Param peer_id path string true "Peer User ID"
// @Param ride_id query string false "Ride ID (опционально)"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/chats/{peer_id}/messages [get]
func (h *Handler) GetChatMessages(c *gin.Context) {
	currentUserIDStr := c.GetString("user_id")
	if currentUserIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	currentUserID, err := uuid.Parse(currentUserIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	peerIDStr := c.Param("peer_id")
	peerID, err := uuid.Parse(peerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer_id format"})
		return
	}

	var rideID *uuid.UUID
	rID := c.Query("ride_id")
	if rID != "" {
		if id, err := uuid.Parse(rID); err == nil {
			rideID = &id
		}
	}

	msgs, err := h.usecase.GetChatHistory(c.Request.Context(), currentUserID, peerID, rideID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if msgs == nil {
		msgs = []domain.Message{}
	}

	c.JSON(http.StatusOK, msgs)
}

type chatResponse struct {
	PartnerID   uuid.UUID       `json:"partner_id"`
	LastMessage *domain.Message `json:"last_message"`
	UnreadCount int32           `json:"unread_count"`
}

// GetChats godoc
// @Summary Получить список диалогов чата с последним сообщением
// @Tags notifications, chats
// @Produce json
// @Success 200 {array} chatResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/notifications/chats [get]
// @Router /api/v1/chats [get]
func (h *Handler) GetChats(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	chats, err := h.usecase.GetUserChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if chats == nil {
		chats = []domain.ChatListItem{}
	}

	c.JSON(http.StatusOK, chats)
}

// GetAdminOnlineStatuses godoc
// @Summary Получить онлайн статусы (для админки)
// @Description Возвращает онлайн статус для переданных user_ids
// @Tags notifications
// @Produce json
// @Param user_ids query string true "Comma-separated user IDs"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/notifications/admin/online-status [get]
func (h *Handler) GetAdminOnlineStatuses(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	userIdsQuery := c.Query("user_ids")
	if userIdsQuery == "" {
		c.JSON(http.StatusOK, map[string]bool{})
		return
	}

	idStrings := strings.Split(userIdsQuery, ",")
	var userIDs []uuid.UUID
	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		if id, err := uuid.Parse(idStr); err == nil {
			userIDs = append(userIDs, id)
		}
	}

	statuses, err := h.usecase.GetOnlineStatuses(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

type wsRequest struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type sendMessageData struct {
	ReceiverID string `json:"receiver_id"`
	RideID     string `json:"ride_id,omitempty"`
	ReplyToID  string `json:"reply_to_id,omitempty"`
	Content    string `json:"content"`
}

type markAsReadData struct {
	MessageIDs []string `json:"message_ids"`
}

// WSHandler godoc
// @Summary WebSocket Endpoint (не тестируется через Swagger напрямую)
// @Description Открыть WebSocket соединение. Отправка сообщений и получение пушей.
// @Tags notifications
// @Param token query string false "JWT токен (альтернатива заголовку Authorization)"
// @Router /ws [get]
func (h *Handler) WSHandler(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Failed to upgrade to websocket", "error", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		// Refresh online status ttl on each pong (heartbeat)
		_ = h.usecase.RefreshOnlineStatus(context.Background(), userID)
		return nil
	})

	ch := h.usecase.Subscribe(userID)
	defer h.usecase.Unsubscribe(userID, ch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Rate limiter (5 req/sec, burst 5)
	limiter := rate.NewLimiter(5, 5)

	// Goroutine for writing messages back to the client
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return // Channel closed
				}

				if err := conn.WriteJSON(msg); err != nil {
					slog.Error("Error writing to ws", "error", err)
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Read messages from the client
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Info("Unexpected ws close: %v", "error", err)
			}
			break
		}

		if !limiter.Allow() {
			slog.Warn("Rate limit exceeded for user", "user_id", userIDStr)
			// Instead of closing connection, just drop the packet to avoid disconnecting
			continue
		}

		var incoming wsRequest
		if err := json.Unmarshal(payload, &incoming); err != nil {
			slog.Error("Failed to parse ws message", "error", err)
			continue
		}

		switch incoming.Event {
		case "send_message":
			var data sendMessageData
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				slog.Error("Failed to parse send_message data", "error", err, "data", string(incoming.Data))
				continue
			}

			receiverID, err := uuid.Parse(data.ReceiverID)
			if err != nil {
				slog.Error("Failed to parse receiverID", "error", err, "receiverID", data.ReceiverID, "raw_data", string(incoming.Data))
				continue
			}

			var rideID *uuid.UUID
			if data.RideID != "" {
				if id, err := uuid.Parse(data.RideID); err == nil {
					rideID = &id
				} else {
					slog.Error("Failed to parse rideID", "error", err, "rideID", data.RideID)
				}
			}

			var replyToID *uuid.UUID
			if data.ReplyToID != "" {
				if id, err := uuid.Parse(data.ReplyToID); err == nil {
					replyToID = &id
				} else {
					slog.Error("Failed to parse replyToID", "error", err, "replyToID", data.ReplyToID)
				}
			}

			go func(userID, receiverID uuid.UUID, rideID, replyToID *uuid.UUID, content string) {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("Panic recovered in SendMessage goroutine", "panic", r)
					}
				}()
				_, err = h.usecase.SendMessage(context.Background(), userID, receiverID, rideID, replyToID, content)
				if err != nil {
					slog.Error("Failed to process send message", "error", err)
				}
			}(userID, receiverID, rideID, replyToID, data.Content)

		case "subscribe_ride":
			var data struct {
				RideID string `json:"ride_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				continue
			}
			if rideID, err := uuid.Parse(data.RideID); err == nil {
				if !h.usecase.VerifyRideParticipant(context.Background(), rideID, userID) {
					slog.Warn("User is not a participant of this ride, dropping subscribe request", "userID", userID, "rideID", rideID)
					continue
				}

				if err := h.usecase.SubscribeToRide(userID, rideID); err != nil {
					slog.Error("Failed to subscribe to ride", "error", err)
				}
			}

		case "location_update":
			var data struct {
				RideID string  `json:"ride_id"`
				Lat    float64 `json:"lat"`
				Lon    float64 `json:"lon"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				continue
			}
			if rideID, err := uuid.Parse(data.RideID); err == nil {
				if !h.usecase.VerifyRideDriver(context.Background(), rideID, userID) {
					slog.Warn("User is not the driver of this ride, dropping location update", "userID", userID, "rideID", rideID)
					continue
				}

				if err := h.usecase.BroadcastLocation(context.Background(), rideID, data.Lat, data.Lon); err != nil {
					slog.Error("Failed to broadcast location", "error", err)
				}
			}

		case "mark_as_read":
			var data markAsReadData
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				continue
			}
			for _, msgIDStr := range data.MessageIDs {
				if msgID, err := uuid.Parse(msgIDStr); err == nil {
					_ = h.usecase.MarkMessageAsRead(context.Background(), msgID, userID)
				}
			}

		case "typing":
			var data struct {
				ReceiverID string `json:"receiver_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				continue
			}
			if recID, err := uuid.Parse(data.ReceiverID); err == nil {
				forwardData := map[string]interface{}{
					"sender_id": userID.String(),
				}
				_ = h.usecase.PushWSEvent(context.Background(), recID, "typing", forwardData)
			}
		}
	}
}

// GetSystemNotifications godoc
// @Summary История системных уведомлений
// @Description Возвращает системные уведомления для текущего пользователя
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications [get]
func (h *Handler) GetSystemNotifications(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifs, err := h.usecase.GetNotificationHistory(c.Request.Context(), userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifs)
}

// MarkSystemNotificationAsRead godoc
// @Summary Пометить уведомление как прочитанное
// @Description Помечает системное уведомление как прочитанное
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/{id}/read [put]
func (h *Handler) MarkSystemNotificationAsRead(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifIDStr := c.Param("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	err = h.usecase.MarkNotificationAsRead(c.Request.Context(), notifID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetVapidKey godoc
// @Summary Получить публичный VAPID ключ
// @Description Возвращает VAPID ключ для настройки Web Push на клиенте
// @Tags notifications
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/notifications/vapid-key [get]
func (h *Handler) GetVapidKey(c *gin.Context) {
	key := h.config.VapidPublicKey
	if key == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "VAPID_PUBLIC_KEY is not configured"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vapid_key": key})
}

