package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	domain "github.com/hextec/yolyolakay-backend/internal/routing/domain"
)

type Handler struct {
	usecase domain.RoutingUseCase
}

func NewHandler(usecase domain.RoutingUseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// GetRoute godoc
// @Summary Построение маршрута
// @Description Получает дистанцию, время и полилинию маршрута от внешнего картографического движка
// @Tags routing
// @Accept json
// @Produce json
// @Param origin_lat query float64 true "Origin Latitude"
// @Param origin_lon query float64 true "Origin Longitude"
// @Param dest_lat query float64 true "Destination Latitude"
// @Param dest_lon query float64 true "Destination Longitude"
// @Success 200 {object} domain.Route
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/routing/route [get]
func (h *Handler) GetRoute(c *gin.Context) {
	origLatStr := c.Query("origin_lat")
	if origLatStr == "" {
		origLatStr = c.Query("from_lat")
	}
	origLonStr := c.Query("origin_lon")
	if origLonStr == "" {
		origLonStr = c.Query("from_lon")
	}
	destLatStr := c.Query("dest_lat")
	if destLatStr == "" {
		destLatStr = c.Query("to_lat")
	}
	destLonStr := c.Query("dest_lon")
	if destLonStr == "" {
		destLonStr = c.Query("to_lon")
	}

	origLat, err1 := strconv.ParseFloat(origLatStr, 64)
	origLon, err2 := strconv.ParseFloat(origLonStr, 64)
	destLat, err3 := strconv.ParseFloat(destLatStr, 64)
	destLon, err4 := strconv.ParseFloat(destLonStr, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinates parameters (expected origin_lat/origin_lon/dest_lat/dest_lon or from_lat/from_lon/to_lat/to_lon)"})
		return
	}

	points := []domain.Coordinate{
		{Latitude: origLat, Longitude: origLon},
		{Latitude: destLat, Longitude: destLon},
	}

	route, err := h.usecase.CalculateRoute(c.Request.Context(), points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

// Geocode godoc
// @Summary Поиск координат по адресу (Forward Geocoding)
// @Description Возвращает список адресов с координатами по текстовому запросу
// @Tags routing
// @Accept json
// @Produce json
// @Param query query string true "Text address to search"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/routing/geocode [get]
func (h *Handler) Geocode(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		query = c.Query("q")
	}
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query (or q) parameter is required"})
		return
	}

	results, err := h.usecase.GeocodeAddress(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ReverseGeocode godoc
// @Summary Поиск адреса по координатам (Reverse Geocoding)
// @Description Возвращает подробный адрес для заданных координат
// @Tags routing
// @Accept json
// @Produce json
// @Param lat query float64 true "Latitude"
// @Param lon query float64 true "Longitude"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/routing/reverse-geocode [get]
func (h *Handler) ReverseGeocode(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat or lon parameter"})
		return
	}

	coord := domain.Coordinate{Latitude: lat, Longitude: lon}
	result, err := h.usecase.ReverseGeocode(c.Request.Context(), coord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
