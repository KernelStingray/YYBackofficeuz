package middleware

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InternalAuth is a lightweight middleware for internal microservices.
// It assumes the API Gateway has already validated the JWT and injected X-User-Id and X-User-Role.
func InternalAuth(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isSystemCall := c.GetHeader("X-System-Call") == "true"
		userIDStr := c.GetHeader("X-User-Id")
		role := c.GetHeader("X-User-Role")

		expectedSecret := os.Getenv("INTERNAL_API_SECRET")
		if expectedSecret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_API_SECRET not configured"})
			return
		}
		providedSecret := c.GetHeader("X-Internal-Secret")
		if providedSecret != expectedSecret {
			slog.Warn("internal auth failed", "ip", c.ClientIP(), "endpoint", c.Request.URL.Path, "reason", "invalid internal secret")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: invalid internal secret"})
			return
		}

		if isSystemCall {
			// Bypass strict UUID validation for internal S2S requests
			c.Set("role", "system")
			c.Set("user_id", userIDStr) // optional, might not be a UUID
			c.Next()
			return
		}

		if userIDStr == "" {
			slog.Warn("internal auth failed", "ip", c.ClientIP(), "endpoint", c.Request.URL.Path, "reason", "missing X-User-Id header from gateway")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-Id header from gateway"})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			slog.Warn("internal auth failed", "ip", c.ClientIP(), "endpoint", c.Request.URL.Path, "reason", "invalid X-User-Id header format", "user_id", userIDStr)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid X-User-Id header format"})
			return
		}

		// Role-Based Access Control (RBAC)
		if len(allowedRoles) > 0 {
			hasRole := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					hasRole = true
					break
				}
			}
			if !hasRole {
				slog.Warn("internal auth failed", "ip", c.ClientIP(), "endpoint", c.Request.URL.Path, "reason", "insufficient permissions", "role", role)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
				return
			}
		}

		// Inject into context just like the old RequireAuth middleware
		c.Set("user_id", userID.String())
		c.Set("role", role)
		sessionIDStr := c.GetHeader("X-Session-Id")
		if sessionIDStr != "" {
			c.Set("session_id", sessionIDStr)
		}

		c.Next()
	}
}
