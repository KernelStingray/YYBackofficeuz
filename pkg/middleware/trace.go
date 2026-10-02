package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/pkg/logger"
)

// TraceMiddleware generates a trace_id if one doesn't exist,
// injects it into the request context, and adds it to the response headers.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Writer.Header().Set("X-Request-ID", traceID)

		ctx := context.WithValue(c.Request.Context(), logger.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
