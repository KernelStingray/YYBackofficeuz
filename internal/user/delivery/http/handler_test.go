package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	delivery "github.com/hextec/yolyolakay-backend/internal/user/delivery/http"
	mocks "github.com/hextec/yolyolakay-backend/pkg/mocks/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUseCase := mocks.NewMockUserUseCase(ctrl)

	setupRouter := func(role string) *gin.Engine {
		r := gin.New()
		authMiddleware := func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Set("role", role)
			c.Next()
		}
		api := r.Group("/api/v1")
		delivery.NewUserHandler(api, userUseCase, authMiddleware)
		return r
	}

	targetID := uuid.New()

	t.Run("Admin can delete user", func(t *testing.T) {
		r := setupRouter("admin")
		userUseCase.EXPECT().DeleteUser(gomock.Any(), targetID).Return(nil).Times(1)

		req := httptest.NewRequest("DELETE", "/api/v1/users/"+targetID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user deleted successfully")
	})

	t.Run("Moderator can delete user", func(t *testing.T) {
		r := setupRouter("moderator")
		userUseCase.EXPECT().DeleteUser(gomock.Any(), targetID).Return(nil).Times(1)

		req := httptest.NewRequest("DELETE", "/api/v1/users/"+targetID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user deleted successfully")
	})

	t.Run("Passenger is forbidden to delete another user", func(t *testing.T) {
		r := setupRouter("passenger")

		req := httptest.NewRequest("DELETE", "/api/v1/users/"+targetID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
