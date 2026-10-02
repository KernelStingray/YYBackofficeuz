package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hextec/yolyolakay-backend/pkg/metrics"
)

// MetricsMiddleware tracks request count, duration, and status code
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // Gets the registered route path instead of actual URL, useful for grouping (e.g. /users/:id)

		if path == "" {
			path = c.Request.URL.Path // fallback if route not found
		}

		// Update metrics
		metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
