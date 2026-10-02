package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/antifraud/domain"
)

type Handler struct {
	usecase domain.AntiFraudUseCase
}

func NewHandler(usecase domain.AntiFraudUseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

type reportMockLocationRequest struct {
	Description string `json:"description"`
}

type TrustScoreResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Score     float64   `json:"score"`
	UpdatedAt string    `json:"updated_at"`
}

// ReportMockLocation godoc
// @Summary Сообщить о подмене GPS
// @Description Штрафует пользователя (обычно водителя) за использование фейковых координат
// @Tags antifraud
// @Accept json
// @Produce json
// @Param request body reportMockLocationRequest true "User Info"
// @Success 200 {object} TrustScoreResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/fraud/report [post]
func (h *Handler) ReportMockLocation(c *gin.Context) {
	var req reportMockLocationRequest
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

	newScore, err := h.usecase.ReportMockLocation(c.Request.Context(), userID, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newScore)
}

// GetTrustScore godoc
// @Summary Получить текущий рейтинг доверия
// @Description Возвращает текущий Trust Score (0-100) пользователя
// @Tags antifraud
// @Accept json
// @Produce json
// @Param user_id query string true "User UUID"
// @Success 200 {object} TrustScoreResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/fraud/score [get]
func (h *Handler) GetTrustScore(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	role := c.GetString("role")
	if (role == "admin" || role == "moderator" || role == "support" || role == "system") && c.Query("user_id") != "" {
		userIDStr = c.Query("user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	score, err := h.usecase.GetTrustScore(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}
