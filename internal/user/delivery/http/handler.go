package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/user/domain"
)

type User struct {
	ID          uuid.UUID  `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Gender      *string    `json:"gender,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	PhoneNumber        string     `json:"phone_number,omitempty"`
	Role               string     `json:"role"`
	KycStatus          string     `json:"kyc_status"`
	KycRejectionReason *string    `json:"kyc_rejection_reason,omitempty"`
	Status             string     `json:"status"`
	Locale      string     `json:"locale"`
	Rating      float64    `json:"rating"`
	RatingCount int32      `json:"rating_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type PublicUserProfile struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Gender      *string   `json:"gender,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Rating      float64   `json:"rating"`
	RatingCount int32     `json:"rating_count"`
	Role               string    `json:"role"`
	KycStatus          string    `json:"kyc_status"`
	KycRejectionReason *string   `json:"kyc_rejection_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}


type PublicCarProfile struct {
	ID        uuid.UUID `json:"id"`
	Brand     string    `json:"brand"`
	Model     string    `json:"model"`
	Color     string    `json:"color"`
	GovNumber string    `json:"gov_number"`
	Year      int32     `json:"year"`
	CarPhotos []string  `json:"car_photos,omitempty"`
}

type UserProfileResponse struct {
	User PublicUserProfile  `json:"user"`
	Cars []PublicCarProfile `json:"cars"`
}

type Car struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	GovNumber          string     `json:"gov_number"`
	Brand              string     `json:"brand"`
	Model              string     `json:"model"`
	TechPassportPhoto  *string    `json:"tech_passport_photo,omitempty"`
	DriverLicensePhoto *string    `json:"driver_license_photo,omitempty"`
	ModerationStatus   string     `json:"moderation_status"`
	CreatedAt          time.Time  `json:"created_at"`

	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type AdminStats struct {
	TotalUsers      int64            `json:"total_users"`
	DriversCount    int64            `json:"drivers_count"`
	PassengersCount int64            `json:"passengers_count"`
	ChartData       []ChartDataPoint `json:"chart_data"`
}

type ChartDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}


type UserHandler struct {
	useCase domain.UserUseCase
}

func NewUserHandler(r *gin.RouterGroup, uc domain.UserUseCase, authMiddleware gin.HandlerFunc) {
	handler := &UserHandler{
		useCase: uc,
	}

	publicUsers := r.Group("/users")
	{
		publicUsers.POST("", handler.RegisterPassenger)
	}

	publicVehicles := r.Group("/vehicles")
	{
		publicVehicles.GET("/:id", handler.GetVehicleByID)
	}

	protectedUsers := r.Group("/users")
	protectedUsers.Use(authMiddleware)
	{
		// Admin Routes
		protectedUsers.GET("/admin/stats", handler.GetAdminStats)
		protectedUsers.GET("/admin/list", handler.GetAdminUsers)
		protectedUsers.GET("/admin/search", handler.SearchAdminUsers)
		protectedUsers.POST("/admin/:id/status", handler.UpdateUserStatus)
		protectedUsers.POST("/admin/kyc/:id/approve", handler.ApproveDriverKYC)
		protectedUsers.POST("/admin/kyc/:id/reject", handler.RejectDriverKYC)

		// User Routes
		protectedUsers.GET("/me", handler.GetMe)
		protectedUsers.PATCH("/me", handler.UpdateProfile) // Alias for /profile
		protectedUsers.DELETE("/me", handler.DeleteAccount)
		protectedUsers.GET("/me/cars", handler.GetCars)
		protectedUsers.POST("/me/cars", handler.CreateCar)
		protectedUsers.PUT("/me/cars/:id", handler.UpdateCar)
		protectedUsers.PATCH("/me/cars/:id", handler.UpdateCar)
		protectedUsers.DELETE("/me/cars/:id", handler.DeleteCar)
		protectedUsers.GET("/blocked", handler.GetBlockedUsers)
		protectedUsers.GET("/:id", handler.GetUserProfile)
		protectedUsers.DELETE("/:id", handler.DeleteUser)
		protectedUsers.PUT("/profile", handler.UpdateProfile)
		protectedUsers.POST("/upgrade", handler.UpgradeToDriver)
		protectedUsers.POST("/:id/block", handler.BlockUser)
		protectedUsers.DELETE("/blocked/:id", handler.UnblockUser)
		protectedUsers.POST("/:id/report", handler.CreateReport)
	}
}

// RegisterPassenger handles the creation of a new passenger profile
// @Summary Register Passenger
// @Description Creates a new user profile with passenger role
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterPassengerRequest true "Registration data"
// @Success 201 {object} User
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/users [post]
func (h *UserHandler) RegisterPassenger(c *gin.Context) {
	var req RegisterPassengerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	user, err := h.useCase.RegisterPassenger(c.Request.Context(), req.FirstName, req.LastName, req.PhoneNumber, req.Gender)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpgradeToDriver handles the request to become a driver (adding car)
// @Summary Upgrade to Driver
// @Description Upgrades a passenger to a driver by adding car and waiting for KYC approval
// @Tags users
// @Accept json
// @Produce json
// @Param request body UpgradeToDriverRequest true "Upgrade Request Body"
// @Success 200 {object} map[string]string
// @Failure 422 {object} map[string]interface{}
// @Router /api/v1/users/upgrade [post]
func (h *UserHandler) UpgradeToDriver(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	var req UpgradeToDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	err = h.useCase.UpgradeToDriver(c.Request.Context(), userID, req.ToCarDomain(), req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully submitted application to upgrade to driver"})
}

// GetUserProfile retrieves the user and their associated cars
// @Summary Get User Profile
// @Description Gets user profile details along with associated cars
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} UserProfileResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, cars, err := h.useCase.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if cars == nil {
		cars = []domain.Car{}
	}

	// Security: Do not expose private fields if the requester is not the user itself (unless admin)
	requesterID := c.GetString("user_id")
	role := c.GetString("role")
	if requesterID != userIDStr && role != "admin" && role != "moderator" && role != "system" {
		user.PhoneNumber = ""
		user.DeletedAt = nil
		user.KycRejectionReason = nil
		for i := range cars {
			cars[i].TechPassportPhoto = nil
			cars[i].DriverLicensePhoto = nil
			cars[i].ModerationStatus = ""
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"cars": cars,
	})
}

// GetAdminStats godoc
// @Summary Fetch user statistics
// @Description Fetch aggregated statistics for users for admin dashboard
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AdminStats
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/admin/stats [get]
func (h *UserHandler) GetAdminStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	stats, err := h.useCase.GetAdminStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAdminUsers retrieves all users for the admin panel
// @Summary Get All Users (Admin)
// @Description Gets all users. Requires Admin role.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} User
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/users/admin/list [get]
func (h *UserHandler) GetAdminUsers(c *gin.Context) {
	role := c.GetString("role")

	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	users, err := h.useCase.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if users == nil {
		users = []*domain.User{}
	}

	c.JSON(http.StatusOK, users)
}

// SearchAdminUsers searches users with pagination
// @Summary Search Users (Admin)
// @Description Searches users with pagination. Requires Admin role.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name or phone"
// @Param role query string false "Filter by role (driver/passenger)"
// @Param status query string false "Filter by status (blocked/active)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/users/admin/search [get]
func (h *UserHandler) SearchAdminUsers(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	page := int32(1)
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = int32(p)
	}
	limit := int32(20)
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		limit = int32(l)
	}

	search := c.Query("search")
	searchRole := c.Query("role")
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")

	users, totalCount, err := h.useCase.SearchAdminUsers(c.Request.Context(), page, limit, search, searchRole, status, kycStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if users == nil {
		users = []*domain.User{}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}

// UpdateProfile handles the request to update basic user details
// @Summary Update User Profile
// @Description Updates user profile details like first name, last name, and gender
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Update profile data"
// @Success 200 {object} User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	user, err := h.useCase.UpdateProfile(c.Request.Context(), userID, req.FirstName, req.LastName, req.Gender, req.Locale, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

var _ domain.User

// GetMe retrieves the current user profile
// @Summary Get Current User Profile
// @Description Gets user profile details for the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	user, cars, err := h.useCase.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if cars == nil {
		cars = []domain.Car{}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"cars": cars,
	})
}

// CreateCar adds a new car for the current user
// @Summary Add a new car
// @Description Adds a new car to the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCarRequest true "Car data"
// @Success 201 {object} Car
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/me/cars [post]
func (h *UserHandler) CreateCar(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	var req CreateCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	car, err := h.useCase.CreateCar(c.Request.Context(), userID, req.ToCarDomain())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, car)
}

// GetCars gets all cars for the current user
// @Summary Get user cars
// @Description Gets all cars for the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Car
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/me/cars [get]
func (h *UserHandler) GetCars(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	cars, err := h.useCase.GetCars(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if cars == nil {
		cars = []domain.Car{}
	}

	c.JSON(http.StatusOK, cars)
}

// UpdateCar godoc
// @Summary Update car details
// @Description Updates details of an existing car belonging to the current user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Car ID"
// @Param request body UpdateCarRequest true "Updated car data"
// @Success 200 {object} Car
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/me/cars/{id} [put]
func (h *UserHandler) UpdateCar(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	carIDStr := c.Param("id")
	carID, err := uuid.Parse(carIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car id format"})
		return
	}

	var req UpdateCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
		return
	}

	car, err := h.useCase.UpdateCar(c.Request.Context(), userID, carID, req.ToCarDomain())
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, car)
}

// DeleteCar godoc
// @Summary Delete car
// @Description Deletes a car belonging to the current user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "Car ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/me/cars/{id} [delete]
func (h *UserHandler) DeleteCar(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	carIDStr := c.Param("id")
	carID, err := uuid.Parse(carIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car id format"})
		return
	}

	err = h.useCase.DeleteCar(c.Request.Context(), userID, carID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "car deleted successfully"})
}

// DeleteAccount godoc
// @Summary Удалить аккаунт
// @Description Помечает аккаунт как удаленный
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/me [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id"})
		return
	}

	err = h.useCase.DeleteAccount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

// DeleteUser permanently deletes a user
// @Summary Delete User
// @Description Permanently deletes a user from the system
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Must be admin or moderator
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	if err := h.useCase.DeleteUser(c.Request.Context(), targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// ApproveDriverKYC approves a driver KYC application
// @Summary Approve Driver KYC
// @Description Approves a pending driver application. Requires Admin/Moderator role.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/admin/kyc/{id}/approve [post]
func (h *UserHandler) ApproveDriverKYC(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	if err := h.useCase.ApproveDriverKYC(c.Request.Context(), targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver application approved successfully"})
}

type RejectKYCRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RejectDriverKYC rejects a driver KYC application
// @Summary Reject Driver KYC
// @Description Rejects a pending driver application. Requires Admin/Moderator role.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body RejectKYCRequest true "Rejection Reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/admin/kyc/{id}/reject [post]
func (h *UserHandler) RejectDriverKYC(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var req RejectKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, reason is required"})
		return
	}

	if err := h.useCase.RejectDriverKYC(c.Request.Context(), targetID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver application rejected successfully"})
}

// BlockUser godoc
// @Summary Заблокировать пользователя
// @Description Блокирует другого пользователя (вы больше не будете видеть его поездки)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to block"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/{id}/block [post]
func (h *UserHandler) BlockUser(c *gin.Context) {
	blockerIDStr := c.GetString("user_id")
	blockedIDStr := c.Param("id")

	blockerID, err := uuid.Parse(blockerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target user id"})
		return
	}

	err = h.useCase.BlockUser(c.Request.Context(), blockerID, blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "blocked"})
}

// UpdateUserStatus (Admin)
// @Summary Update User Status (Admin)
// @Description Updates user status (active/blocked)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param status body map[string]string true "Status payload (status: active/blocked)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/users/admin/{id}/status [post]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err = h.useCase.UpdateUserStatus(c.Request.Context(), userID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

// UnblockUser godoc
// @Summary Разблокировать пользователя
// @Description Разблокирует ранее заблокированного пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to unblock"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/blocked/{id} [delete]
func (h *UserHandler) UnblockUser(c *gin.Context) {
	blockerIDStr := c.GetString("user_id")
	blockedIDStr := c.Param("id")

	blockerID, err := uuid.Parse(blockerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target user id"})
		return
	}

	err = h.useCase.UnblockUser(c.Request.Context(), blockerID, blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unblocked"})
}

// GetBlockedUsers godoc
// @Summary Получить список заблокированных пользователей
// @Description Возвращает ID пользователей, которых вы заблокировали
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} User
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/blocked [get]
func (h *UserHandler) GetBlockedUsers(c *gin.Context) {
	blockerIDStr := c.GetString("user_id")
	blockerID, err := uuid.Parse(blockerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	users, err := h.useCase.GetBlockedUsers(c.Request.Context(), blockerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	c.JSON(http.StatusOK, users)
}

// CreateReport godoc
// @Summary Пожаловаться на пользователя
// @Description Отправляет жалобу на другого пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reported User ID (UUID)"
// @Param request body map[string]interface{} true "Данные жалобы (reason)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/{id}/report [post]
func (h *UserHandler) CreateReport(c *gin.Context) {
	reporterIDStr := c.GetString("user_id")
	reporterID, err := uuid.Parse(reporterIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	reportedIDStr := c.Param("id")
	reportedID, err := uuid.Parse(reportedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reported_id format in path"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	err = h.useCase.CreateReport(c.Request.Context(), reporterID, reportedID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "reported"})
}

// GetVehicleByID godoc
// @Summary Get vehicle by ID
// @Description Returns public info about a vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} Car
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vehicles/{id} [get]
func (h *UserHandler) GetVehicleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle id"})
		return
	}

	car, err := h.useCase.GetCarByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}

	// IDOR Protection: Sanitize sensitive document photos if requester is not the car owner or admin
	requesterIDStr := c.GetString("user_id")
	requesterID, _ := uuid.Parse(requesterIDStr)
	role := c.GetString("role")

	if car.UserID != requesterID && role != "admin" && role != "super_admin" {
		car.TechPassportPhoto = nil
		car.DriverLicensePhoto = nil
	}

	c.JSON(http.StatusOK, car)
}
