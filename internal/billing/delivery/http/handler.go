package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	billingDomain "github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/internal/billing/worker"
)

type Config struct {
	WebhookAllowedIPs string
}

type Handler struct {
	usecase     billingDomain.BillingUseCase
	distributor worker.TaskDistributor
	config      Config
}

func NewHandler(usecase billingDomain.BillingUseCase, distributor worker.TaskDistributor, cfg Config) *Handler {
	return &Handler{
		usecase:     usecase,
		distributor: distributor,
		config:      cfg,
	}
}

func (h *Handler) InitRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	api := router.Group("/api/v1/billing")

	// Webhooks (No auth middleware, but IP whitelist)
	api.POST("/click/webhook", IPWhitelistMiddleware(h.config.WebhookAllowedIPs), h.HandleClickWebhook)
	api.POST("/click/prepare", IPWhitelistMiddleware(h.config.WebhookAllowedIPs), h.HandleClickPrepare)
	api.POST("/click/complete", IPWhitelistMiddleware(h.config.WebhookAllowedIPs), h.HandleClickComplete)
	api.GET("/click/webhook", h.HandleClickPing)
	api.GET("/click/prepare", h.HandleClickPing)
	api.GET("/click/complete", h.HandleClickPing)
	api.POST("/payme/webhook", IPWhitelistMiddleware(h.config.WebhookAllowedIPs), h.HandlePaymeWebhook)
	api.GET("/payme/webhook", h.HandleClickPing)

	// Catalog is public
	api.GET("/plans", h.GetPlans)

	protected := api.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/subscription", h.GetDriverSubscription)
		protected.POST("/cards", h.AddPaymentMethod)
		protected.GET("/cards", h.GetPaymentMethods)
		protected.POST("/pay", h.SubscribeDriver)
		protected.GET("/transactions/:id", h.GetTransactionStatus)

		// Admin endpoint for manual gift injection
		protected.POST("/admin/gift-inject", h.GiftInject)
		protected.GET("/admin/stats", h.GetAdminStats)
	}
}

// GetPlans godoc
// @Summary Get available billing plans
// @Description returns a static catalog of available subscription plans
// @Tags Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/billing/plans [get]
func (h *Handler) GetPlans(c *gin.Context) {
	driverPlans := []map[string]interface{}{
		{
			"plan_key":           "driver_trial",
			"plan_name":          "Free Trial",
			"target_role":        "driver",
			"price":              0,
			"currency":           "UZS",
			"rides_quota":        1,
			"free_parcels_quota": 1,
			"features":           []string{"1 trial ride", "1 trial parcel"},
			"features_uz":        []string{"1 ta sinov safari", "1 ta sinov posilkasi"},
			"features_ru":        []string{"1 пробная поездка", "1 пробная посылка"},
			"features_en":        []string{"1 trial ride", "1 trial parcel"},
		},
		{
			"plan_key":           "driver_plus",
			"plan_name":          "Plus",
			"target_role":        "driver",
			"price":              5000,
			"currency":           "UZS",
			"rides_quota":        2,
			"free_parcels_quota": 2,
			"features":           []string{"Up to 2 rides", "2 free parcels"},
			"features_uz":        []string{"2 tagacha safar", "2 ta bepul posilka"},
			"features_ru":        []string{"До 2 поездок", "2 бесплатные посылки"},
			"features_en":        []string{"Up to 2 rides", "2 free parcels"},
		},
		{
			"plan_key":           "driver_pro",
			"plan_name":          "Pro",
			"target_role":        "driver",
			"price":              7000,
			"currency":           "UZS",
			"rides_quota":        5,
			"free_parcels_quota": 5,
			"features":           []string{"Up to 5 rides", "5 free parcels"},
			"features_uz":        []string{"5 tagacha safar", "5 ta bepul posilka"},
			"features_ru":        []string{"До 5 поездок", "5 бесплатных посылок"},
			"features_en":        []string{"Up to 5 rides", "5 free parcels"},
		},
		{
			"plan_key":           "driver_premium",
			"plan_name":          "Premium",
			"target_role":        "driver",
			"price":              10000,
			"currency":           "UZS",
			"rides_quota":        60,
			"free_parcels_quota": -1,
			"features":           []string{"Up to 60 rides", "Unlimited free parcels"},
			"features_uz":        []string{"60 tagacha safar", "Cheksiz bepul posilkalar"},
			"features_ru":        []string{"До 60 поездок", "Безлимитные бесплатные посылки"},
			"features_en":        []string{"Up to 60 rides", "Unlimited free parcels"},
		},
	}

	role := strings.ToLower(strings.TrimSpace(c.Query("role")))
	var activePlans []map[string]interface{}
	if role == "passenger" {
		activePlans = []map[string]interface{}{}
	} else {
		activePlans = driverPlans
	}

	response := gin.H{
		"plans":           activePlans,
		"driver_plans":    driverPlans,
		"passenger_plans": []map[string]interface{}{},
	}

	c.JSON(http.StatusOK, response)
}

type AddPaymentMethodRequest struct {
	Provider  string `json:"provider" binding:"required"`
	Token     string `json:"token" binding:"required"`
	MaskedPan string `json:"masked_pan" binding:"required"`
}

type SubscriptionResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	PlanKey          string    `json:"plan_key"`
	PlanName         string    `json:"plan_name"`
	Status           string    `json:"status"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	HasUsedTrial     bool      `json:"has_used_trial"`
	RidesLeft        int32     `json:"rides_left"`
	FreeParcelsQuota int32     `json:"free_parcels_quota"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SubscribeDriverRequest struct {
	PlanName       string  `json:"plan_name"`
	PlanKey        string  `json:"plan_key"`
	PaymentTokenID *string `json:"payment_token_id"`
	PromoCode      *string `json:"promo_code"`
	ReturnURL      *string `json:"return_url,omitempty"`
}

// GetDriverSubscription godoc
// @Summary Get active subscription for a driver
// @Description returns active subscription details
// @Tags Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SubscriptionResponse
// @Router /api/v1/billing/subscription [get]
func (h *Handler) GetDriverSubscription(c *gin.Context) {
	role := c.GetString("role")
	if strings.ToLower(role) == "passenger" {
		c.JSON(http.StatusOK, nil)
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	sub, err := h.usecase.GetDriverSubscription(c.Request.Context(), driverID)
	if err != nil {
		// If subscription is not found (or any other error), return 200 OK with null
		// This prevents the frontend from treating it as a hard error and displaying an error banner.
		c.JSON(http.StatusOK, nil)
		return
	}

	if sub != nil {
		if sub.RidesLeft < 0 && sub.RidesLeft != -1 {
			sub.RidesLeft = 0
		}
		if sub.FreeParcelsQuota < 0 && sub.FreeParcelsQuota != -1 {
			sub.FreeParcelsQuota = 0
		}
	}

	c.JSON(http.StatusOK, sub)
}

// AddPaymentMethod godoc
// @Summary Save a payment card token
// @Description Mocks the saving of a card token from Click/Payme
// @Tags Billing
// @Accept json
// @Produce json
// @Param req body AddPaymentMethodRequest true "Card token details"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/billing/cards [post]
func (h *Handler) AddPaymentMethod(c *gin.Context) {
	var req AddPaymentMethodRequest
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

	pt, err := h.usecase.AddPaymentMethod(c.Request.Context(), userID, req.Provider, req.Token, req.MaskedPan)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, pt)
}

// GetPaymentMethods godoc
// @Summary Get saved cards for a user
// @Description returns all active payment tokens for a user
// @Tags Billing
// @Accept json
// @Produce json
// @Router /api/v1/billing/cards [get]
func (h *Handler) GetPaymentMethods(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	tokens, err := h.usecase.GetPaymentMethods(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// SubscribeDriver godoc
// @Summary Pay for a subscription
// @Description Deducts money using a saved card and activates a SaaS plan
// @Tags Billing
// @Accept json
// @Produce json
// @Param req body SubscribeDriverRequest true "Subscription payment details"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/billing/pay [post]
func (h *Handler) SubscribeDriver(c *gin.Context) {
	var req SubscribeDriverRequest
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

	targetPlan := strings.TrimSpace(req.PlanKey)
	if targetPlan == "" {
		targetPlan = strings.TrimSpace(req.PlanName)
	}
	if targetPlan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_key or plan_name is required"})
		return
	}

	if strings.HasPrefix(strings.ToLower(targetPlan), "passenger_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passenger service is 100% free, paid subscriptions are for drivers only"})
		return
	}

	var pTokenID *uuid.UUID
	if req.PaymentTokenID != nil {
		id, err := uuid.Parse(*req.PaymentTokenID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_token_id format"})
			return
		}
		pTokenID = &id
	}

	paymentURL, txID, err := h.usecase.SubscribeDriver(c.Request.Context(), driverID, targetPlan, pTokenID, req.PromoCode, req.ReturnURL)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "insufficient") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "payment initiated",
		"payment_url":    paymentURL,
		"transaction_id": txID.String(),
	})
}

// GetTransactionStatus godoc
// @Summary Get status of a payment transaction
// @Description returns transaction status (pending, success, failed, cancelled)
// @Tags Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/billing/transactions/{id} [get]
func (h *Handler) GetTransactionStatus(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}
	txIDStr := c.Param("id")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	tx, err := h.usecase.GetTransaction(c.Request.Context(), userID, txID)
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

	planName := ""
	if strings.HasPrefix(tx.Type, "subscription:") {
		planName = strings.TrimPrefix(tx.Type, "subscription:")
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"status":         tx.Status,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"provider":       tx.Provider,
		"plan":           planName,
		"created_at":     tx.CreatedAt,
	})
}


func IPWhitelistMiddleware(allowedIPsStr string) gin.HandlerFunc {
	if allowedIPsStr == "" {
		allowedIPsStr = "*"
	}
	allowedIPs := strings.Split(allowedIPsStr, ",")

	return func(c *gin.Context) {
		if allowedIPsStr == "*" || strings.Contains(allowedIPsStr, "*") {
			c.Next()
			return
		}
		clientIPStr := c.GetHeader("CF-Connecting-IP")
		if clientIPStr == "" {
			clientIPStr = c.GetHeader("X-Forwarded-For")
		}
		if clientIPStr == "" {
			clientIPStr = c.ClientIP()
		}
		if strings.Contains(clientIPStr, ",") {
			clientIPStr = strings.Split(clientIPStr, ",")[0]
		}
		clientIPStr = strings.TrimSpace(clientIPStr)

		clientIP := net.ParseIP(clientIPStr)
		allowed := false
		for _, ipStr := range allowedIPs {
			ipStr = strings.TrimSpace(ipStr)
			if ipStr == "" {
				continue
			}
			if ipStr == "*" {
				allowed = true
				break
			}
			if strings.Contains(ipStr, "/") {
				_, subnet, err := net.ParseCIDR(ipStr)
				if err == nil && subnet.Contains(clientIP) {
					allowed = true
					break
				}
			} else {
				if ipStr == clientIPStr {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			// Click and Payme provide cryptographic signatures with a shared secret key.
			// If a request provides a Click or Payme signature parameter, we log a warning
			// and let the cryptographically secured usecase verify the signature,
			// preventing false positive downtime from Click's dynamic IPs.
			hasSign := c.PostForm("sign_string") != "" || c.GetHeader("X-Auth-Sign") != "" || c.Query("sign_string") != ""
			if hasSign {
				slog.Warn("Payment webhook received from non-whitelisted IP, proceeding to signature verification",
					"ip", clientIPStr, "path", c.Request.URL.Path)
				c.Next()
				return
			}

			slog.Warn("Payment webhook rejected: IP not allowed and no signature provided",
				"ip", clientIPStr, "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not allowed"})
			return
		}
		c.Next()
	}
}

// WebhookRequest represents a common webhook payload (simplified)
type WebhookRequest struct {
	ExternalTxID string  `json:"external_tx_id"`
	TxID         string  `json:"tx_id"`
	Amount       float64 `json:"amount"`
	Timestamp    int64   `json:"timestamp"`
}

func (h *Handler) HandleClickPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "click_webhook",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) HandleClickPrepare(c *gin.Context) {
	h.processClick(c, "0")
}

func (h *Handler) HandleClickComplete(c *gin.Context) {
	h.processClick(c, "1")
}

func (h *Handler) HandleClickWebhook(c *gin.Context) {
	h.processClick(c, "")
}

func (h *Handler) processClick(c *gin.Context, defaultAction string) {
	// 1. Parse parameters (supports both application/x-www-form-urlencoded and application/json)
	clickTransID := strings.TrimSpace(c.PostForm("click_trans_id"))
	serviceID := strings.TrimSpace(c.PostForm("service_id"))
	merchantTransID := strings.TrimSpace(c.PostForm("merchant_trans_id"))
	merchantPrepareID := strings.TrimSpace(c.PostForm("merchant_prepare_id"))
	amount := strings.TrimSpace(c.PostForm("amount"))
	action := strings.TrimSpace(c.PostForm("action"))
	signTime := strings.TrimSpace(c.PostForm("sign_time"))
	signString := strings.TrimSpace(c.PostForm("sign_string"))
	errorStr := strings.TrimSpace(c.PostForm("error"))

	// Fallback to JSON payload if form data was empty
	if clickTransID == "" && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		var bodyMap map[string]interface{}
		if err := c.ShouldBindJSON(&bodyMap); err == nil {
			getStr := func(k string) string {
				if v, ok := bodyMap[k]; ok {
					return fmt.Sprintf("%v", v)
				}
				return ""
			}
			clickTransID = getStr("click_trans_id")
			serviceID = getStr("service_id")
			merchantTransID = getStr("merchant_trans_id")
			merchantPrepareID = getStr("merchant_prepare_id")
			amount = getStr("amount")
			action = getStr("action")
			signTime = getStr("sign_time")
			signString = getStr("sign_string")
			errorStr = getStr("error")
		}
	}

	// Infer action from endpoint path if not explicitly provided
	if action == "" {
		if defaultAction != "" {
			action = defaultAction
		} else if strings.HasSuffix(c.Request.URL.Path, "/prepare") {
			action = "0"
		} else if strings.HasSuffix(c.Request.URL.Path, "/complete") {
			action = "1"
		}
	}

	slog.Info("Click webhook received",
		"path", c.Request.URL.Path,
		"action", action,
		"click_trans_id", clickTransID,
		"merchant_trans_id", merchantTransID,
		"merchant_prepare_id", merchantPrepareID,
		"amount", amount,
		"sign_time", signTime,
		"ip", c.ClientIP(),
	)

	// If click reports an external error, acknowledge cancellation
	if errorStr != "" && errorStr != "0" {
		slog.Warn("Click reported external error", "error", errorStr, "click_trans_id", clickTransID)
		c.JSON(http.StatusOK, gin.H{
			"error":      -9,
			"error_note": "Transaction cancelled",
		})
		return
	}

	var resp map[string]interface{}
	var err error

	switch action {
	case "0":
		resp, err = h.usecase.ClickPrepare(c.Request.Context(), clickTransID, serviceID, merchantTransID, amount, action, signTime, signString)
	case "1":
		resp, err = h.usecase.ClickComplete(c.Request.Context(), clickTransID, serviceID, merchantTransID, merchantPrepareID, amount, action, signTime, signString)
	default:
		slog.Warn("Click webhook received unknown action", "action", action, "click_trans_id", clickTransID)
		c.JSON(http.StatusOK, gin.H{
			"error":      -3,
			"error_note": "Action not found",
		})
		return
	}

	if err != nil {
		slog.Warn("Click action failed", "action", action, "error", err, "click_trans_id", clickTransID, "resp", resp)
		if resp == nil {
			c.JSON(http.StatusOK, gin.H{
				"error":      -1,
				"error_note": "System Error",
			})
			return
		}
	} else {
		slog.Info("Click action succeeded", "action", action, "click_trans_id", clickTransID, "merchant_trans_id", merchantTransID)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandlePaymeWebhook(c *gin.Context) {
	rawBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var req WebhookRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if time.Now().Unix()-req.Timestamp > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook expired (replay protection)"})
		return
	}

	signature := c.GetHeader("X-Auth-Sign")

	txID, err := uuid.Parse(req.TxID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tx_id"})
		return
	}

	err = h.usecase.ProcessWebhook(c.Request.Context(), "payme", rawBody, signature, req.ExternalTxID, txID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type GiftInjectRequest struct {
	Quota int32 `json:"quota" binding:"required,gt=0"`
}

// GiftInject godoc
// @Summary Inject free parcels to premium drivers
// @Description Creates an async task to add free parcels to all active premium drivers
// @Tags Billing Admin
// @Accept json
// @Produce json
// @Param req body GiftInjectRequest true "Gift quota details"
// @Success 202 {object} map[string]string
// @Router /api/v1/billing/admin/gift-inject [post]
func (h *Handler) GiftInject(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: admin role required"})
		return
	}

	var req GiftInjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := &worker.GiftInjectPayload{
		Quota: req.Quota,
	}

	if h.distributor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task distributor not initialized"})
		return
	}

	err := h.distributor.DistributeTaskGiftInject(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "gift inject task enqueued successfully"})
}

// GetAdminStats godoc
// @Summary Fetch billing statistics
// @Description Fetch aggregated statistics for billing for admin dashboard
// @Tags Billing Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} billingDomain.AdminStats
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/billing/admin/stats [get]
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
