package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	deliveryHttp "github.com/hextec/yolyolakay-backend/internal/notification/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	notifMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/notification"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetHistory_Symmetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := notifMocks.NewMockNotificationUseCase(ctrl)
	handler := deliveryHttp.NewHandler(mockUC, deliveryHttp.Config{})

	userA := uuid.New()
	userB := uuid.New()

	expectedMsgs := []domain.Message{
		{
			ID:         uuid.New(),
			SenderID:   userA,
			ReceiverID: userB,
			Content:    "Salom!",
			Status:     domain.MessageStatusDelivered,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			SenderID:   userB,
			ReceiverID: userA,
			Content:    "Va alaykum salom!",
			Status:     domain.MessageStatusRead,
			CreatedAt:  time.Now(),
		},
	}

	setupRouter := func(callerID uuid.UUID) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", callerID.String())
			c.Next()
		})
		r.GET("/api/v1/notifications/history", handler.GetHistory)
		r.GET("/api/v1/chats/:peer_id/messages", handler.GetChatMessages)
		return r
	}

	t.Run("Caller A: user_id_1=A and user_id_2=B", func(t *testing.T) {
		mockUC.EXPECT().GetChatHistory(gomock.Any(), userA, userB, nil).Return(expectedMsgs, nil).Times(1)

		r := setupRouter(userA)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/notifications/history?user_id_1="+userA.String()+"&user_id_2="+userB.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var msgs []domain.Message
		_ = json.Unmarshal(w.Body.Bytes(), &msgs)
		assert.Len(t, msgs, 2)
	})

	t.Run("Caller A: user_id_1=B and user_id_2=A (reversed order)", func(t *testing.T) {
		mockUC.EXPECT().GetChatHistory(gomock.Any(), userA, userB, nil).Return(expectedMsgs, nil).Times(1)

		r := setupRouter(userA)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/notifications/history?user_id_1="+userB.String()+"&user_id_2="+userA.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var msgs []domain.Message
		_ = json.Unmarshal(w.Body.Bytes(), &msgs)
		assert.Len(t, msgs, 2)
	})

	t.Run("Caller A: peer_id=B", func(t *testing.T) {
		mockUC.EXPECT().GetChatHistory(gomock.Any(), userA, userB, nil).Return(expectedMsgs, nil).Times(1)

		r := setupRouter(userA)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/notifications/history?peer_id="+userB.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var msgs []domain.Message
		_ = json.Unmarshal(w.Body.Bytes(), &msgs)
		assert.Len(t, msgs, 2)
	})

	t.Run("Caller A via /chats/:peer_id/messages", func(t *testing.T) {
		mockUC.EXPECT().GetChatHistory(gomock.Any(), userA, userB, nil).Return(expectedMsgs, nil).Times(1)

		r := setupRouter(userA)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/chats/"+userB.String()+"/messages", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var msgs []domain.Message
		_ = json.Unmarshal(w.Body.Bytes(), &msgs)
		assert.Len(t, msgs, 2)
	})
}
