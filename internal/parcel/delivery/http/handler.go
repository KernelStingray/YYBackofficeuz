package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/parcel/domain"
	userPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
)

type Parcel struct {
	ID                     uuid.UUID           `json:"id"`
	SenderID               uuid.UUID           `json:"sender_id"`
	DriverID               *uuid.UUID          `json:"driver_id,omitempty"`
	OriginLon              float64             `json:"origin_lon"`
	OriginLat              float64             `json:"origin_lat"`
	DestinationLon         float64             `json:"destination_lon"`
	DestinationLat         float64             `json:"destination_lat"`
	DecryptedReceiverPhone string              `json:"receiver_phone,omitempty"`
	ReceiverName           string              `json:"receiver_name,omitempty"`
	DropoffType            string              `json:"dropoff_type,omitempty"`
	Status                 domain.ParcelStatus `json:"status" swaggertype:"string"`
	Price                  float64             `json:"price"`
	OriginAddress          string              `json:"from_address"`
	DestinationAddress     string              `json:"to_address"`
	Description            string              `json:"description,omitempty"`
	Size                   string              `json:"size,omitempty"`
	DeliverBy              *time.Time          `json:"deliver_by,omitempty"`
	SenderName             string              `json:"sender_name,omitempty"`
	SenderRating           float64             `json:"sender_rating,omitempty"`
	ResponsesCount         int32               `json:"responses_count"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}

type ParcelResponse struct {
	ID           uuid.UUID                   `json:"id"`
	ParcelID     uuid.UUID                   `json:"parcel_id"`
	DriverID     uuid.UUID                   `json:"driver_id"`
	Status       domain.ParcelResponseStatus `json:"status" swaggertype:"string"`
	OfferedPrice float64                     `json:"offered_price"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

type Handler struct {
	usecase    domain.ParcelUseCase
	userClient userPb.UserServiceClient
}

func NewHandler(usecase domain.ParcelUseCase, userClient userPb.UserServiceClient) *Handler {
	return &Handler{
		usecase:    usecase,
		userClient: userClient,
	}
}

type createParcelRequest struct {
	OriginLon          float64 `json:"origin_lon" binding:"required"`
	OriginLat          float64 `json:"origin_lat" binding:"required"`
	DestinationLon     float64 `json:"destination_lon" binding:"required"`
	DestinationLat     float64 `json:"destination_lat" binding:"required"`
	ReceiverPhone      string  `json:"receiver_phone" binding:"required,max=20"`
	ReceiverName       string  `json:"receiver_name" binding:"max=100"`
	DropoffType        string  `json:"dropoff_type" binding:"max=20"`
	Price              float64 `json:"price" binding:"required,gt=0,lte=100000000"`
	OriginAddress      string  `json:"from_address" binding:"required,max=255"`
	DestinationAddress string  `json:"to_address" binding:"required,max=255"`
	Description        string  `json:"description" binding:"max=500"`
	Size               string  `json:"size" binding:"max=50"`
	DeliverBy          string  `json:"deliver_by" binding:"max=50"` // ISO8601 string
}

// CreateParcel godoc
// @Summary Создать посылку
// @Description Создает новую посылку (Слепая логистика скрывает телефон)
// @Tags parcels
// @Accept json
// @Produce json
// @Param request body createParcelRequest true "Данные посылки"
// @Success 201 {object} Parcel
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels [post]
func (h *Handler) CreateParcel(c *gin.Context) {
	var req createParcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	senderIDStr := c.GetString("user_id")
	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	var deliverByTime *time.Time
	if req.DeliverBy != "" {
		t, err := time.Parse(time.RFC3339, req.DeliverBy)
		if err == nil {
			deliverByTime = &t
		}
	}

	parcel, err := h.usecase.CreateParcel(c.Request.Context(), senderID, req.OriginLon, req.OriginLat, req.DestinationLon, req.DestinationLat, req.ReceiverPhone, req.ReceiverName, req.DropoffType, req.Price, req.OriginAddress, req.DestinationAddress, req.Description, req.Size, deliverByTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, parcel)
}

// SearchParcels godoc
// @Summary Поиск посылок рядом
// @Description Ищет свободные посылки в радиусе (в метрах) для водителей. Телефон скрыт.
// @Tags parcels
// @Accept json
// @Produce json
// @Param lon query number true "Долгота водителя"
// @Param lat query number true "Широта водителя"
// @Param radius query number false "Радиус поиска в метрах (по умолчанию 5000)"
// @Param price_min query number false "Минимальная цена"
// @Param price_max query number false "Максимальная цена"
// @Param city query string false "Город"
// @Success 200 {array} Parcel
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/search [get]
func (h *Handler) SearchParcels(c *gin.Context) {
	driverIDStr := c.GetString("user_id")
	lonStr := c.Query("lon")
	if lonStr == "" {
		lonStr = c.Query("origin_lon")
	}
	latStr := c.Query("lat")
	if latStr == "" {
		latStr = c.Query("origin_lat")
	}
	radiusStr := c.Query("radius")
	if radiusStr == "" {
		radiusStr = c.Query("radius_km")
	}

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id in token"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lon (or origin_lon)"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat (or origin_lat)"})
		return
	}

	radius := 5000.0
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			if r <= 100 {
				radius = r * 1000 // convert km to meters
			} else {
				radius = r
			}
		}
	}

	params := domain.ParcelSearchParams{
		Lon:          lon,
		Lat:          lat,
		RadiusMeters: radius,
	}

	destLonStr := c.Query("dest_lon")
	if destLonStr == "" {
		destLonStr = c.Query("destination_lon")
	}
	destLatStr := c.Query("dest_lat")
	if destLatStr == "" {
		destLatStr = c.Query("destination_lat")
	}
	if destLonStr != "" && destLatStr != "" {
		if dLon, err := strconv.ParseFloat(destLonStr, 64); err == nil {
			params.DestLon = &dLon
		}
		if dLat, err := strconv.ParseFloat(destLatStr, 64); err == nil {
			params.DestLat = &dLat
		}
	}

	corridorStr := c.Query("corridor_km")
	if corridorStr == "" {
		corridorStr = c.Query("corridor")
	}
	if corridorStr != "" {
		if cVal, err := strconv.ParseFloat(corridorStr, 64); err == nil {
			var corridorMeters float64
			if cVal <= 100 {
				corridorMeters = cVal * 1000 // convert km to meters
			} else {
				corridorMeters = cVal
			}
			params.CorridorMeters = &corridorMeters
		}
	}

	priceMinStr := c.Query("price_min")
	if priceMinStr != "" {
		if p, err := strconv.ParseFloat(priceMinStr, 64); err == nil {
			params.PriceMin = &p
		}
	}

	priceMaxStr := c.Query("price_max")
	if priceMaxStr != "" {
		if p, err := strconv.ParseFloat(priceMaxStr, 64); err == nil {
			params.PriceMax = &p
		}
	}

	params.City = c.Query("city")

	parcels, err := h.usecase.SearchParcels(c.Request.Context(), driverID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, parcels)
}

// AcceptParcel godoc
// @Summary Принять посылку водителем
// @Description Закрепляет посылку за водителем (Mock проверки премиум статуса). Дешифрует телефон.
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Param driver_id query string true "Driver ID"
// @Success 200 {object} Parcel
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string "Только для премиум водителей"
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id}/accept [post]
func (h *Handler) AcceptParcel(c *gin.Context) {
	parcelIDStr := c.Param("id")
	driverIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id format"})
		return
	}

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id format"})
		return
	}

	parcel, err := h.usecase.AcceptParcel(c.Request.Context(), parcelID, driverID)
	if err != nil {
		// Mock logic returns error if not premium or not pending
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, parcel)
}

// GetAdminParcels godoc
// @Summary Get All Parcels (Admin)
// @Description Gets all parcels for the admin panel. Requires Admin role.
// @Tags parcels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Parcel
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/parcels/admin/list [get]
func (h *Handler) GetAdminParcels(c *gin.Context) {
	role := c.GetString("role")
	statusFilter := c.Query("status")

	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	parcels, err := h.usecase.GetAllParcels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	var filtered []domain.Parcel
	for _, p := range parcels {
		if statusFilter == "" || statusFilter == "active" {
			if p.Status != "canceled" && p.Status != "delivered" {
				filtered = append(filtered, p)
			}
		} else if statusFilter == string(p.Status) {
			filtered = append(filtered, p)
		}
	}

	if filtered == nil {
		filtered = []domain.Parcel{}
	}

	c.JSON(http.StatusOK, filtered)
}

// GetMyParcels godoc
// @Summary Мои посылки
// @Description Возвращает посылки текущего пользователя (отправителя или водителя). Поддерживает query параметр ?role=sender (по-умолчанию) или ?role=driver
// @Tags parcels
// @Accept json
// @Produce json
// @Param role query string false "Роль пользователя: sender или driver"
// @Success 200 {array} Parcel
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/parcels/my [get]
func (h *Handler) GetMyParcels(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	role := c.Query("role")
	if role != "driver" {
		role = "sender"
	}

	parcels, err := h.usecase.GetMyParcels(c.Request.Context(), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if parcels == nil {
		parcels = []domain.Parcel{}
	}
	c.JSON(http.StatusOK, parcels)
}

// RespondToParcel godoc
// @Summary Откликнуться на посылку
// @Description Водитель предлагает свою цену для доставки посылки
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Param request body map[string]float64 true "Offered Price"
// @Success 200 {object} ParcelResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id}/respond [post]
func (h *Handler) RespondToParcel(c *gin.Context) {
	parcelIDStr := c.Param("id")
	driverIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid driver id"})
		return
	}

	var req struct {
		OfferedPrice float64 `json:"offered_price" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.RespondToParcel(c.Request.Context(), parcelID, driverID, req.OfferedPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type ParcelResponseDTO struct {
	ID           string  `json:"id"`
	ParcelID     string  `json:"parcel_id"`
	DriverID     string  `json:"driver_id"`
	DriverName   string  `json:"driver_name"`
	DriverAvatar string  `json:"driver_avatar"`
	DriverRating float64 `json:"driver_rating"`
	Status       string  `json:"status"`
	OfferedPrice float64 `json:"offered_price"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// GetParcelResponses godoc
// @Summary Получить отклики на посылку
// @Description Отправитель смотрит отклики от водителей на свою посылку
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Success 200 {array} ParcelResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id}/responses [get]
func (h *Handler) GetParcelResponses(c *gin.Context) {
	parcelIDStr := c.Param("id")
	senderIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	responses, err := h.usecase.GetParcelResponses(c.Request.Context(), parcelID, senderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dtos []ParcelResponseDTO
	for _, r := range responses {
		dto := ParcelResponseDTO{
			ID:           r.ID.String(),
			ParcelID:     r.ParcelID.String(),
			DriverID:     r.DriverID.String(),
			Status:       string(r.Status),
			OfferedPrice: r.OfferedPrice,
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
		}

		if h.userClient != nil {
			profileResp, err := h.userClient.GetDriverProfile(c.Request.Context(), &userPb.GetDriverProfileRequest{
				DriverId: r.DriverID.String(),
			})
			if err == nil && profileResp != nil {
				dto.DriverName = profileResp.Name
				dto.DriverAvatar = profileResp.AvatarUrl
				dto.DriverRating = profileResp.Rating
			}
		}

		dtos = append(dtos, dto)
	}

	if dtos == nil {
		dtos = []ParcelResponseDTO{}
	}
	c.JSON(http.StatusOK, dtos)
}

// GetParcelByID godoc
// @Summary Получить посылку по ID
// @Description Получает детальную информацию о посылке
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Success 200 {object} Parcel
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id} [get]
func (h *Handler) GetParcelByID(c *gin.Context) {
	parcelIDStr := c.Param("id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	parcel, err := h.usecase.GetParcelByID(c.Request.Context(), parcelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, parcel)
}

// AcceptParcelResponse godoc
// @Summary Принять отклик водителя
// @Description Отправитель принимает предложенную цену от водителя
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Param response_id path string true "Response ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id}/responses/{response_id}/accept [post]
func (h *Handler) AcceptParcelResponse(c *gin.Context) {
	parcelIDStr := c.Param("id")
	responseIDStr := c.Param("response_id")
	senderIDStr := c.GetString("user_id")

	responseID, err := uuid.Parse(responseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid response id"})
		return
	}

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	err = h.usecase.AcceptParcelResponse(c.Request.Context(), parcelID, responseID, senderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

// RejectParcelResponse godoc
// @Summary Отклонить отклик водителя
// @Description Отправитель отклоняет предложенную цену от водителя
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Param response_id path string true "Response ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/parcels/{id}/responses/{response_id}/reject [post]
func (h *Handler) RejectParcelResponse(c *gin.Context) {
	parcelIDStr := c.Param("id")
	responseIDStr := c.Param("response_id")
	senderIDStr := c.GetString("user_id")

	responseID, err := uuid.Parse(responseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid response id"})
		return
	}

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	err = h.usecase.RejectParcelResponse(c.Request.Context(), parcelID, responseID, senderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

// CancelParcel godoc
// @Summary Отменить посылку
// @Description Отправитель отменяет свою посылку (только если она в статусе pending)
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/parcels/{id}/cancel [post]
func (h *Handler) CancelParcel(c *gin.Context) {
	parcelIDStr := c.Param("id")
	senderIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	err = h.usecase.CancelParcel(c.Request.Context(), parcelID, senderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// PickupParcel godoc
// @Summary Забрать посылку
// @Description Водитель отмечает, что забрал посылку
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/parcels/{id}/pickup [post]
func (h *Handler) PickupParcel(c *gin.Context) {
	parcelIDStr := c.Param("id")
	driverIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	err = h.usecase.PickupParcel(c.Request.Context(), parcelID, driverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "picked_up"})
}

// DeliverParcel godoc
// @Summary Доставить посылку
// @Description Водитель отмечает, что доставил посылку
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/parcels/{id}/deliver [post]
func (h *Handler) DeliverParcel(c *gin.Context) {
	parcelIDStr := c.Param("id")
	driverIDStr := c.GetString("user_id")

	parcelID, err := uuid.Parse(parcelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parcel id"})
		return
	}

	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	err = h.usecase.DeliverParcel(c.Request.Context(), parcelID, driverID)
	if err != nil {
		if strings.Contains(err.Error(), "only the assigned driver") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "delivered"})
}

// ConfirmDelivery godoc
// @Summary Подтверждение доставки посылки (водителем или отправителем)
// @Description Позволяет водителю или отправителю подтвердить доставку посылки
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/parcels/{id}/confirm-delivery [post]
func (h *Handler) ConfirmDelivery(c *gin.Context) {
	h.DeliverParcel(c)
}

// GetAdminStats godoc
// @Summary Fetch parcel statistics
// @Description Fetch aggregated statistics for parcels for admin dashboard
// @Tags parcels
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.AdminStats
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parcels/admin/stats [get]

func (h *Handler) GetAdminStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	stats, err := h.usecase.GetAdminStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserStats godoc
// @Summary Получить статистику посылок пользователя (Админ)
// @Description Получает количество посылок отправленных и доставленных пользователем
// @Tags parcels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} domain.UserParcelStats
// @Router /api/v1/parcels/admin/user/{id}/stats [get]
func (h *Handler) GetUserStats(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.usecase.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
