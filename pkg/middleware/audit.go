package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type auditPayload struct {
	AdminID   string                 `json:"admin_id"`
	Action    string                 `json:"action"`
	Entity    string                 `json:"entity"`
	IPAddress string                 `json:"ip_address"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// AuditMiddleware logs POST/PUT/DELETE requests to RabbitMQ
func AuditMiddleware(rmqClient *rabbitmq.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Proceed with the request
		method := c.Request.Method

		// We only want to audit state-changing requests
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete && method != http.MethodPatch {
			c.Next()
			return
		}

		// Read body for logging
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// Do not log if request failed at routing level
		if c.Writer.Status() == http.StatusNotFound || c.Writer.Status() == http.StatusMethodNotAllowed {
			return
		}

		adminID := c.GetString("user_id") // set by auth middleware
		if adminID == "" {
			adminID = "system" // Or ignore
		}

		var payloadMap map[string]interface{}
		if len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &payloadMap)
		}

		logPayload := auditPayload{
			AdminID:   adminID,
			Action:    method,
			Entity:    c.Request.URL.Path,
			IPAddress: c.ClientIP(),
			Payload:   payloadMap,
		}

		eventBytes, err := json.Marshal(logPayload)
		if err == nil && rmqClient != nil {
			// Publish to audit logs
			_ = rmqClient.SetupExchangeQueue("audit.events", "topic", "audit.logs", "audit.log.*")
			if err := rmqClient.Publish(c.Request.Context(), "audit.events", "audit.log.created", eventBytes); err != nil {
				slog.Error("Failed to publish", "error", err)
			}
		}
	}
}
