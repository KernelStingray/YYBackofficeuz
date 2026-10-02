package http

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authDomain "github.com/hextec/yolyolakay-backend/internal/auth/domain"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}

type SessionResponse struct {
	SessionID string    `json:"session_id"`
	DeviceID  string    `json:"device_id"`
	DeviceIP  string    `json:"device_ip"`
	UserAgent string    `json:"user_agent"`
	OSType    string    `json:"os_type"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AuthHandler struct {
	useCase authDomain.AuthUseCase
}

func NewAuthHandler(router *gin.RouterGroup, uc authDomain.AuthUseCase, authMiddleware gin.HandlerFunc) {
	handler := &AuthHandler{
		useCase: uc,
	}

	router.POST("/send-otp", handler.SendOTP)
	router.POST("/verify-otp", handler.VerifyOTP)
	router.POST("/login", handler.Login)
	router.POST("/refresh", handler.Refresh)
	router.GET("/mock-otp/*phone", handler.GetMockOTP)

	// Protected routes for session management
	if authMiddleware != nil {
		protected := router.Group("/")
		protected.Use(authMiddleware)
		protected.POST("/setup-password", handler.SetupPassword)
		protected.GET("/sessions", handler.GetSessions)
		protected.DELETE("/sessions/:device_id", handler.DeleteSession)
	}
}

type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Type        string `json:"type" binding:"required"`
}

// SendOTP godoc
// @Summary Send OTP via SMS
// @Description Sends a 5-digit code to the specified phone number
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body SendOTPRequest true "Phone Number"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "User not found (for login/frgt_pswd)"
// @Failure 409 {object} map[string]string "User already exists (for register)"
// @Failure 429 {object} map[string]string
// @Router /api/v1/auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or missing 'type'"})
		return
	}

	if req.Type != "register" && req.Type != "login" && req.Type != "frgt_pswd" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type (must be register, login, or frgt_pswd)"})
		return
	}

	ip := c.ClientIP()
	err := h.useCase.SendOTP(c.Request.Context(), req.PhoneNumber, ip, req.Type)
	if err != nil {
		if err.Error() == "rate limit exceeded: max 3 requests per 5 minutes" ||
			err.Error() == "rate limit exceeded: max 10 requests per hour from this IP" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Password    string `json:"password,omitempty"`
	Type        string `json:"type"` // e.g. "sms", "flashcall"
}

// VerifyOTP godoc
// @Summary Verify OTP and Login
// @Description Verifies the OTP and issues JWT tokens (optionally sets password if provided)
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body VerifyOTPRequest true "Phone, Code and optional Password"
// @Param X-Device-ID header string false "Device ID"
// @Param X-OS-Type header string false "OS Type"
// @Param User-Agent header string false "User Agent"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Password != "" && len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters long"})
		return
	}

	ip := c.ClientIP()
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = uuid.NewString() // Generate a temporary one if not provided
	}
	osType := c.GetHeader("X-OS-Type")
	userAgent := c.GetHeader("User-Agent")

	tokens, err := h.useCase.VerifyOTP(c.Request.Context(), req.PhoneNumber, req.Code, ip, deviceID, userAgent, osType, req.Password)
	if err != nil {
		if err.Error() == "account is temporarily locked due to too many failed attempts" {
			c.JSON(http.StatusLocked, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh godoc
// @Summary Refresh JWT Token
// @Description Exchanges a valid refresh token for a new pair of tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body RefreshRequest true "Refresh Token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ip := c.ClientIP()
	tokens, err := h.useCase.RefreshToken(c.Request.Context(), req.RefreshToken, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// GetSessions godoc
// @Summary Get active sessions
// @Description Retrieves all active sessions for the authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SessionResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}

	sessions, err := h.useCase.GetActiveSessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []SessionResponse
	for _, s := range sessions {
		response = append(response, SessionResponse{
			SessionID: s.SessionID.String(),
			DeviceID:  s.DeviceID,
			DeviceIP:  s.DeviceIP,
			UserAgent: s.UserAgent,
			OSType:    s.OSType,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSession godoc
// @Summary Terminate session
// @Description Terminates a specific active session by device_id
// @Tags Auth
// @Param device_id path string true "Device ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/sessions/{device_id} [delete]
func (h *AuthHandler) DeleteSession(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}

	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing device_id param"})
		return
	}

	err = h.useCase.TerminateSession(c.Request.Context(), userID, deviceID)
	if err != nil {
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session terminated"})
}

type SetupPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// SetupPassword godoc
// @Summary Setup or update password
// @Description Sets up a password for the authenticated user to login without OTP in the future
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req body SetupPasswordRequest true "New Password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/setup-password [post]
func (h *AuthHandler) SetupPassword(c *gin.Context) {
	var req SetupPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body (password min 6 chars)"})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}

	err = h.useCase.SetupPassword(c.Request.Context(), userID, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// Login godoc
// @Summary Login with password
// @Description Authenticates a user using phone number and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body LoginRequest true "Phone and Password"
// @Param X-Device-ID header string false "Device ID"
// @Param X-OS-Type header string false "OS Type"
// @Param User-Agent header string false "User Agent"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string "New device detected, OTP required"
// @Failure 423 {object} map[string]string "Account temporarily locked"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ip := c.ClientIP()
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = uuid.NewString() // Generate a temporary one if not provided
	}
	osType := c.GetHeader("X-OS-Type")
	userAgent := c.GetHeader("User-Agent")

	tokens, err := h.useCase.LoginWithPassword(c.Request.Context(), req.PhoneNumber, req.Password, ip, deviceID, userAgent, osType)
	if err != nil {
		if err.Error() == "account is temporarily locked due to too many failed attempts" {
			c.JSON(http.StatusLocked, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "new_device_otp_sent" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "new_device",
				"message": "Новое устройство. СМС-код успешно отправлен.",
			})
			return
		}
		if strings.HasPrefix(err.Error(), "new_device_sms_failed:") {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "new_device_sms_failed",
				"message": "Код не отправлен, попробуйте позже.",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// GetMockOTP returns mock OTP for automated E2E testing - strictly restricted to test accounts and X-Test-Secret
func (h *AuthHandler) GetMockOTP(c *gin.Context) {
	testSecret := os.Getenv("TEST_SECRET")
	if testSecret == "" {
		testSecret = os.Getenv("INTERNAL_API_SECRET")
	}
	if testSecret == "" {
		testSecret = "super-internal-secret-for-microservices"
	}
	if c.GetHeader("X-Test-Secret") != testSecret {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	phone := c.Param("phone")
	phone = strings.TrimPrefix(phone, "/")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number is required"})
		return
	}

	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "+998991234567"
	}

	// SECURITY: Only allow known test accounts. Real accounts return 404 and NEVER generate OTP.
	isTestAccount := phone == testPhone || phone == "+998991234567" || strings.HasPrefix(phone, "+99800")
	if !isTestAccount {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	testOTP := os.Getenv("TEST_OTP")
	if testOTP == "" {
		testOTP = "99887"
	}

	_ = h.useCase.SaveMockOTP(c.Request.Context(), phone, testOTP)

	c.JSON(http.StatusOK, gin.H{
		"phone": phone,
		"otp":   testOTP,
	})
}

var _ SessionResponse
