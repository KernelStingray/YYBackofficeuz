package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	telematicsDomain "github.com/hextec/yolyolakay-backend/internal/telematics/domain"
)

type Handler struct {
	usecase telematicsDomain.TelematicsUseCase
}

func NewHandler(usecase telematicsDomain.TelematicsUseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

type pingRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
}

// Ping godoc
// @Summary Отправка GPS пинга
// @Description Высоконагруженный эндпоинт для приема координат от водителей. Данные мгновенно оседают в Redis и фоново батчатся в PostgreSQL.
// @Tags telematics
// @Accept json
// @Produce json
// @Param request body pingRequest true "Текущие координаты водителя"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/ping [post]
func (h *Handler) Ping(c *gin.Context) {
	var req pingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	loc := &telematicsDomain.Location{
		DriverID:  driverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Speed:     req.Speed,
		Heading:   req.Heading,
		Timestamp: time.Now(),
	}

	if err := h.usecase.ProcessPing(c.Request.Context(), loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 202 Accepted because persistence is asynchronous
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// GetNearby godoc
// @Summary Поиск водителей рядом
// @Description Ищет водителей в заданном радиусе (в км) с помощью Redis GEORADIUS
// @Tags telematics
// @Accept json
// @Produce json
// @Param lat query float64 true "Latitude"
// @Param lon query float64 true "Longitude"
// @Param radius query float64 true "Radius (km)"
// @Success 200 {array} telematicsDomain.NearbyDriver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/nearby [get]
func (h *Handler) GetNearby(c *gin.Context) {
	role := c.GetString("role")
	if role != "driver" && role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied: nearby driver telemetry is restricted to drivers. Passengers must use /api/v1/rides/{id}/location"})
		return
	}

	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius")

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	radius, err3 := strconv.ParseFloat(radiusStr, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	drivers, err := h.usecase.GetNearbyDrivers(c.Request.Context(), lat, lon, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

type LocationResponse struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Timestamp time.Time `json:"timestamp"`
}

// GetDriverLocation godoc
// @Summary Получение текущих координат конкретного водителя
// @Tags telematics
// @Produce json
// @Param id path string false "Driver ID"
// @Param driver_id query string false "Driver ID (query parameter alternative)"
// @Success 200 {object} LocationResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/drivers/{id}/location [get]
func (h *Handler) GetDriverLocation(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("driver_id")
	}
	driverID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id format"})
		return
	}

	callerIDStr := c.GetString("user_id")
	callerID, _ := uuid.Parse(callerIDStr)
	role := c.GetString("role")

	if callerID != driverID && role != "admin" && role != "moderator" {
		isAuth, err := h.usecase.IsAuthorizedPassenger(c.Request.Context(), driverID, callerID)
		if err != nil || !isAuth {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied: driver location is restricted to the driver and confirmed passengers of active rides"})
			return
		}
	}

	loc, err := h.usecase.GetDriverLocation(c.Request.Context(), driverID)
	if err != nil || loc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "driver location not found or driver offline"})
		return
	}

	c.JSON(http.StatusOK, loc)
}

// GetLocationByDriverID godoc
// @Summary Получение координат водителя по query-параметру driver_id
// @Description Возвращает координаты водителя для участника активной поездки
// @Tags telematics
// @Produce json
// @Param driver_id query string true "Driver ID"
// @Success 200 {object} LocationResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/location [get]
func (h *Handler) GetLocationByDriverID(c *gin.Context) {
	h.GetDriverLocation(c)
}

type ETAResponse struct {
	EtaSeconds int `json:"eta_seconds"`
	EtaMinutes int `json:"eta_minutes"`
}

// CalculateETA godoc
// @Summary Расчет ETA
// @Description Рассчитывает время прибытия заданного водителя к заданной точке (через формулу Гаверсинуса)
// @Tags telematics
// @Accept json
// @Produce json
// @Param driver_id query string true "Driver ID"
// @Param target_lat query float64 true "Target Latitude"
// @Param target_lon query float64 true "Target Longitude"
// @Success 200 {object} ETAResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/eta [get]
func (h *Handler) CalculateETA(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id query parameter is required"})
		return
	}
	latStr := c.Query("target_lat")
	lonStr := c.Query("target_lon")

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	targetLat, err1 := strconv.ParseFloat(latStr, 64)
	targetLon, err2 := strconv.ParseFloat(lonStr, 64)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target coordinates"})
		return
	}

	callerIDStr := c.GetString("user_id")
	callerID, _ := uuid.Parse(callerIDStr)
	role := c.GetString("role")

	if callerID != driverID && role != "admin" && role != "moderator" {
		isAuth, err := h.usecase.IsAuthorizedPassenger(c.Request.Context(), driverID, callerID)
		if err != nil || !isAuth {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied: driver ETA is restricted to the driver and confirmed passengers of active rides"})
			return
		}
	}

	etaDuration, err := h.usecase.CalculateETA(c.Request.Context(), driverID, targetLat, targetLon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Returns ETA in seconds and minutes for convenience
	c.JSON(http.StatusOK, gin.H{
		"eta_seconds": int(etaDuration.Seconds()),
		"eta_minutes": int(etaDuration.Minutes()),
	})
}

// GetAdminLocations godoc
// @Summary Получение позиций всех водителей
// @Description Возвращает текущее местоположение всех активных водителей из Redis (для админки)
// @Tags telematics
// @Produce json
// @Success 200 {array} telematicsDomain.Location
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/telematics/admin/locations [get]
func (h *Handler) GetAdminLocations(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	locations, err := h.usecase.GetAllActiveLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}
