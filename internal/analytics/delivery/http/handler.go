package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/analytics/domain"
)

type AnalyticsHandler struct {
	usecase domain.AnalyticsUseCase
}

func NewAnalyticsHandler(r *gin.Engine, usecase domain.AnalyticsUseCase, authMiddleware gin.HandlerFunc) {
	handler := &AnalyticsHandler{
		usecase: usecase,
	}

	api := r.Group("/api/v1/analytics")
	api.Use(authMiddleware)
	{
		api.POST("/events", handler.LogEvent)
		api.GET("/stats", handler.GetEventStats)
	}
}

type logEventRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	UserID    *string                `json:"user_id"`
	Payload   map[string]interface{} `json:"payload"`
}

type EventStatsResponse struct {
	EventType  string `json:"event_type"`
	EventCount int64  `json:"event_count"`
}

// LogEvent godoc
// @Summary Log a new analytics event
// @Description Logs telemetry data/event into the Data Lake
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body logEventRequest true "Event Details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/events [post]
func (h *AnalyticsHandler) LogEvent(c *gin.Context) {
	var req logEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	var uid *uuid.UUID

	// Check context for user ID first
	userIDStr := c.GetString("user_id")
	if userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err == nil {
			uid = &parsed
		}
	} else if req.UserID != nil && *req.UserID != "" {
		parsed, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
			return
		}
		uid = &parsed
	}

	if req.Payload == nil {
		req.Payload = make(map[string]interface{})
	}

	err := h.usecase.LogEvent(c.Request.Context(), req.EventType, uid, req.Payload)
	if err != nil {
		slog.Error("failed to log analytics event", "error", err, "event_type", req.EventType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log analytics event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetEventStats godoc
// @Summary Get event statistics
// @Description Retrieves aggregated event counts for the BI Dashboard
// @Tags analytics
// @Produce json
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Param group_by query string false "Group by (e.g. day, month)"
// @Param event_type query string false "Event type filter (e.g. app_open, ride_create)"
// @Success 200 {array} EventStatsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/stats [get]
func (h *AnalyticsHandler) GetEventStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "analyst" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions for analytics stats"})
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	groupBy := c.Query("group_by")
	eventType := c.Query("event_type")

	stats, err := h.usecase.GetEventStatsFiltered(c.Request.Context(), fromStr, toStr, groupBy, eventType)
	if err != nil {
		slog.Error("failed to retrieve event statistics", "error", err, "group_by", groupBy, "event_type", eventType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve event statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
