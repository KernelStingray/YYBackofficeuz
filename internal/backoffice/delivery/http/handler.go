package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	backofficeDomain "github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"strings"
	"time"
)

type BackofficeHandler struct {
	usecase    backofficeDomain.BackofficeUseCase
	jwtManager *jwt.JWTManager
}

func NewBackofficeHandler(r *gin.Engine, usecase backofficeDomain.BackofficeUseCase, authMiddleware gin.HandlerFunc, auditMiddleware gin.HandlerFunc, jwtManager *jwt.JWTManager) {
	handler := &BackofficeHandler{
		usecase:    usecase,
		jwtManager: jwtManager,
	}

	api := r.Group("/api/v1/backoffice")
	{
		api.POST("/admins/login", handler.Login)
		api.GET("/configs/:key", handler.GetRemoteConfig)

		protected := api.Group("")
		protected.Use(authMiddleware)
		if auditMiddleware != nil {
			protected.Use(auditMiddleware)
		}
		{
			protected.GET("/admins", handler.GetAdmins)
			protected.PUT("/configs/:key", handler.SetRemoteConfig)
			protected.GET("/logs/stream", handler.StreamLogs)
		}
	}
}

// GetAdmins godoc
// @Summary List admins
// @Description Returns list of all admin users
// @Tags backoffice
// @Produce json
// @Success 200 {array} backofficeDomain.Admin
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/backoffice/admins [get]
func (h *BackofficeHandler) GetAdmins(c *gin.Context) {
	admins, err := h.usecase.GetAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admins"})
		return
	}

	c.JSON(http.StatusOK, admins)
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary Admin login
// @Description Authenticates an admin and returns their profile
// @Tags backoffice
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login Credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/backoffice/admins/login [post]
func (h *BackofficeHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.usecase.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := h.jwtManager.GenerateAccessToken(admin.ID, admin.Role, uuid.New(), 1*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin": admin,
		"token": token,
	})
}

type setRemoteConfigRequest struct {
	Value map[string]interface{} `json:"value" binding:"required"`
}

// SetRemoteConfig godoc
// @Summary Set remote config
// @Description Sets or updates a dynamic JSON configuration for the mobile app
// @Tags backoffice
// @Accept json
// @Produce json
// @Param key path string true "Config Key"
// @Param request body setRemoteConfigRequest true "Config Value"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/backoffice/configs/{key} [put]
func (h *BackofficeHandler) SetRemoteConfig(c *gin.Context) {
	key := c.Param("key")

	var req setRemoteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.SetRemoteConfig(c.Request.Context(), key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetRemoteConfig godoc
// @Summary Get remote config
// @Description Retrieves a dynamic JSON configuration for the mobile app
// @Tags backoffice
// @Produce json
// @Param key path string true "Config Key"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/backoffice/configs/{key} [get]
func (h *BackofficeHandler) GetRemoteConfig(c *gin.Context) {
	key := c.Param("key")

	config, err := h.usecase.GetRemoteConfig(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// StreamLogs godoc
// @Summary Stream infrastructure logs
// @Description Connects to Loki and streams logs via Server-Sent Events (SSE)
// @Tags backoffice
// @Produce text/event-stream
// @Security BearerAuth
// @Success 200
// @Router /api/v1/backoffice/logs/stream [get]
func cleanServiceName(m map[string]string) string {
	name := m["container"]
	if name == "" {
		name = m["compose_service"]
	}
	if name == "" {
		name = m["container_name"]
	}
	if name == "" {
		name = "system"
	}
	name = strings.TrimPrefix(name, "yolyolakay-")
	name = strings.TrimPrefix(name, "yybackend-")
	name = strings.TrimSuffix(name, "-1")
	return name
}

func (h *BackofficeHandler) StreamLogs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Send initial ping to confirm connection
	c.SSEvent("connected", "Streaming logs from Loki...")
	c.Writer.Flush()

	// 1. Fetch recent logs from Loki HTTP query_range so user immediately sees history
	recentURL := fmt.Sprintf("http://loki:3100/loki/api/v1/query_range?query=%s&limit=40", url.QueryEscape(`{container=~".+"}`))
	client := &http.Client{Timeout: 3 * time.Second}
	if respHTTP, errHTTP := client.Get(recentURL); errHTTP == nil && respHTTP.StatusCode == http.StatusOK {
		defer respHTTP.Body.Close()
		var histResp struct {
			Data struct {
				Result []struct {
					Stream map[string]string `json:"stream"`
					Values [][]string        `json:"values"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.NewDecoder(respHTTP.Body).Decode(&histResp); err == nil {
			for _, stream := range histResp.Data.Result {
				serviceName := cleanServiceName(stream.Stream)
				for _, val := range stream.Values {
					if len(val) >= 2 {
						c.SSEvent("log", fmt.Sprintf("[%s] %s", serviceName, val[1]))
					}
				}
			}
			c.Writer.Flush()
		}
	}

	// 2. Connect to Loki websocket for live streaming
	u := url.URL{Scheme: "ws", Host: "loki:3100", Path: "/loki/api/v1/tail"}
	q := u.Query()
	q.Set("query", `{container=~".+"}`)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Failed to connect to Loki: %v", err)
		c.SSEvent("error", "Failed to connect to log server")
		return
	}
	defer conn.Close()

	// Handle client disconnect
	clientGone := c.Writer.CloseNotify()

	type LokiResponse struct {
		Streams []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"streams"`
	}

	for {
		select {
		case <-clientGone:
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Loki WS error: %v", err)
				}
				return
			}

			var resp LokiResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				continue
			}

			for _, stream := range resp.Streams {
				serviceName := cleanServiceName(stream.Stream)

				for _, val := range stream.Values {
					if len(val) >= 2 {
						logLine := fmt.Sprintf("[%s] %s", serviceName, val[1])
						c.SSEvent("log", logLine)
					}
				}
			}
			c.Writer.Flush()
		}
	}
}
