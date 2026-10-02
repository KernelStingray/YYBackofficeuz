package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayAuth_PublicAndProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	jwtManager := jwt.NewJWTManager(privateKey, publicKey)

	publicRoutes := []string{
		"/metrics",
		"/health",
		"/api/v1/auth/send-otp",
		"/api/v1/auth/verify-otp",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/backoffice/admins/login",
		"/api/v1/backoffice/configs",
		"/api/v1/config",
		"/api/v1/docs/*",
		"GET:/api/v1/media/*",
		"GET:/api/v1/rides/search",
		"GET:/api/v1/rides/driver/:uuid",
		"GET:/api/v1/rides/:uuid/location",
		"GET:/api/v1/rides/:uuid",
		"GET:/api/v1/users/:uuid/rides",
		"POST:/api/v1/billing/click/webhook",
		"POST:/api/v1/billing/payme/webhook",
	}

	testUserID := uuid.New()
	testSessionID := uuid.New()
	validToken, err := jwtManager.GenerateAccessToken(testUserID, "passenger", testSessionID, time.Hour)
	require.NoError(t, err)

	setupRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(GatewayAuth(jwtManager, publicRoutes))
		r.Any("/*path", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.Request.Header.Get("X-User-Id"),
				"role":    c.Request.Header.Get("X-User-Role"),
			})
		})
		return r
	}

	r := setupRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		token          string
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "Public media GET without token",
			method:         "GET",
			path:           "/api/v1/media/avatar-123.png",
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Protected media POST upload without token",
			method:         "POST",
			path:           "/api/v1/media/upload",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Protected media POST upload with token",
			method:         "POST",
			path:           "/api/v1/media/upload",
			token:          validToken,
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name:           "Public rides search GET",
			method:         "GET",
			path:           "/api/v1/rides/search",
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Public single ride GET by valid UUID",
			method:         "GET",
			path:           "/api/v1/rides/" + uuid.New().String(),
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Admin active rides GET without token should be 401",
			method:         "GET",
			path:           "/api/v1/rides/admin/active",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Admin active rides GET with valid token",
			method:         "GET",
			path:           "/api/v1/rides/admin/active",
			token:          validToken,
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name:           "My rides GET without token should be 401",
			method:         "GET",
			path:           "/api/v1/rides/my",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Ride recommendations GET without token should be 401",
			method:         "GET",
			path:           "/api/v1/rides/recommendations",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Post ride create without token should be 401",
			method:         "POST",
			path:           "/api/v1/rides/",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Public ride location GET without token",
			method:         "GET",
			path:           "/api/v1/rides/" + uuid.New().String() + "/location",
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Public driver rides GET without token",
			method:         "GET",
			path:           "/api/v1/rides/driver/" + uuid.New().String(),
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Public user rides GET without token",
			method:         "GET",
			path:           "/api/v1/users/" + uuid.New().String() + "/rides",
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Protected watch token POST without token should be 401",
			method:         "POST",
			path:           "/api/v1/rides/" + uuid.New().String() + "/watch-token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "Click webhook POST without token",
			method:         "POST",
			path:           "/api/v1/billing/click/webhook",
			token:          "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
