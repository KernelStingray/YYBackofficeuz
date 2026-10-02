package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ratingDomain "github.com/hextec/yolyolakay-backend/internal/rating/domain"
)

type Handler struct {
	useCase ratingDomain.RatingUseCase
}

func NewHandler(useCase ratingDomain.RatingUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) InitRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/api/v1/rating")
	{
		group.GET("", h.GetFeedbacks)
		group.POST("", authMiddleware, h.SubmitFeedback)
	}
}


type SubmitFeedbackRequest struct {
	RideID     string   `json:"ride_id" binding:"required"`
	RevieweeID string   `json:"reviewee_id" binding:"required"`
	Score      int32    `json:"score" binding:"required,min=1,max=5"`
	Comment    string   `json:"comment"`
	Tags       []string `json:"tags"`
}

// SubmitFeedback godoc
// @Summary Submit a new feedback rating
// @Tags rating
// @Accept json
// @Produce json
// @Param body body SubmitFeedbackRequest true "Feedback Details"
// @Success 201 {object} ratingDomain.Feedback
// @Router /api/v1/rating [post]
// @Security BearerAuth
func (h *Handler) SubmitFeedback(c *gin.Context) {
	var req SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rideID, err := uuid.Parse(req.RideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride_id format"})
		return
	}

	revieweeID, err := uuid.Parse(req.RevieweeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reviewee_id format"})
		return
	}

	// Extract reviewer ID from context (injected by internal auth middleware)
	reviewerIDStr := c.GetString("user_id")
	reviewerID, err := uuid.Parse(reviewerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing X-User-Id header"})
		return
	}

	feedback, err := h.useCase.SubmitFeedback(c.Request.Context(), rideID, reviewerID, revieweeID, req.Score, req.Comment, req.Tags)
	if err != nil {
		if errors.Is(err, ratingDomain.ErrCanOnlyRateCompletedRides) ||
			errors.Is(err, ratingDomain.ErrCannotRateYourself) ||
			errors.Is(err, ratingDomain.ErrScoreOutOfRange) ||
			errors.Is(err, ratingDomain.ErrNotRideParticipant) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}

// GetFeedbacks godoc
type ReviewerProfileDTO struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar string    `json:"avatar"`
}

type FeedbackItemDTO struct {
	ID        uuid.UUID           `json:"id"`
	RideID    uuid.UUID           `json:"ride_id"`
	Reviewer  *ReviewerProfileDTO `json:"reviewer,omitempty"`
	Score     int32               `json:"score"`
	Comment   string              `json:"comment"`
	Tags      []string            `json:"tags"`
	CreatedAt time.Time           `json:"created_at"`
}

type FeedbackListResponse struct {
	Items      []FeedbackItemDTO `json:"items"`
	Total      int64             `json:"total"`
	Average    float64           `json:"average"`
	TagSummary map[string]int    `json:"tag_summary"`
}

// @Summary Get user feedbacks and ratings
// @Description Fetches paginated reviews for a user (driver or passenger)
// @Tags rating
// @Accept json
// @Produce json
// @Param reviewee_id query string true "Reviewee User ID (UUID)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Param with_comment query bool false "Only return reviews with non-empty comments"
// @Success 200 {object} FeedbackListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rating [get]
func (h *Handler) GetFeedbacks(c *gin.Context) {
	revieweeIDStr := c.Query("reviewee_id")
	if revieweeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reviewee_id query parameter is required"})
		return
	}

	revieweeID, err := uuid.Parse(revieweeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reviewee_id format"})
		return
	}

	limit := 20
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := c.Query("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	withComment := false
	if wcStr := c.Query("with_comment"); wcStr != "" {
		withComment, _ = strconv.ParseBool(wcStr)
	}

	resp, err := h.useCase.GetFeedbacks(c.Request.Context(), revieweeID, limit, offset, withComment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

