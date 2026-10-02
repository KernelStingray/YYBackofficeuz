package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	helpdeskDomain "github.com/hextec/yolyolakay-backend/internal/helpdesk/domain"
)

type Handler struct {
	usecase helpdeskDomain.HelpdeskUseCase
}

func NewHandler(usecase helpdeskDomain.HelpdeskUseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

type createTicketRequest struct {
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// CreateTicket godoc
// @Summary Создать новый тикет
// @Description Создает обращение в службу поддержки (жалоба, ДТП, забытые вещи)
// @Tags helpdesk
// @Accept json
// @Produce json
// @Param request body createTicketRequest true "Ticket Data"
// @Success 201 {object} helpdeskDomain.Ticket
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/helpdesk/tickets [post]
func (h *Handler) CreateTicket(c *gin.Context) {
	var req createTicketRequest
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

	ticket, err := h.usecase.CreateTicket(c.Request.Context(), userID, req.Subject, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// GetUserTickets godoc
// @Summary Получить тикеты текущего пользователя
// @Description Возвращает список всех обращений текущего авторизованного пользователя из JWT токена
// @Tags helpdesk
// @Accept json
// @Produce json
// @Success 200 {array} helpdeskDomain.Ticket
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/helpdesk/tickets [get]
func (h *Handler) GetUserTickets(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	tickets, err := h.usecase.GetUserTickets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

type resolveTicketRequest struct {
	TicketID   string `json:"ticket_id" binding:"required"`
	Resolution string `json:"resolution" binding:"required"`
}

// ResolveTicket godoc
// @Summary Решить тикет (для оператора)
// @Description Переводит тикет в статус RESOLVED и сохраняет текст ответа оператора
// @Tags helpdesk
// @Accept json
// @Produce json
// @Param request body resolveTicketRequest true "Ticket Resolution Data"
// @Success 200 {object} helpdeskDomain.Ticket
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/helpdesk/tickets/resolve [post]
func (h *Handler) ResolveTicket(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "support" && role != "operator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: operator or admin permissions required"})
		return
	}

	var req resolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketID, err := uuid.Parse(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket_id format"})
		return
	}

	ticket, err := h.usecase.ResolveTicket(c.Request.Context(), ticketID, req.Resolution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}
