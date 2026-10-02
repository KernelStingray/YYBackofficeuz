package middleware

import (
	"crypto/rsa"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
)

// AuthMiddleware creates a Gin handler that validates RS256 JWT tokens.
// It also checks if the user's role is in the allowedRoles list (if provided).
func RequireAuth(publicKey *rsa.PublicKey, allowedRoles ...string) gin.HandlerFunc {
	// Initialize the JWTManager with only the public key, since it only validates.
	jwtManager := jwt.NewJWTManager(nil, publicKey)

	return func(c *gin.Context) {
		var tokenString string

		// 1. Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				slog.Warn("authentication failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "invalid authorization header format")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				return
			}
			tokenString = parts[1]
		} else {
			// 2. Fallback to query parameter (useful for WebSockets)
			tokenString = c.Query("token")
			if tokenString == "" {
				slog.Warn("authentication failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "missing token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header or token query parameter is required"})
				return
			}
		}

		// 3. Validate Token Signature and Expiry
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			slog.Warn("authentication failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "invalid or expired token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// 4. Role-Based Access Control (RBAC)
		if len(allowedRoles) > 0 {
			hasRole := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}
			if !hasRole {
				slog.Warn("authorization failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "insufficient permissions", "role", claims.Role)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
				return
			}
		}

		// 5. Inject claims into Context for downstream handlers
		c.Set("user_id", claims.UserID.String())
		c.Set("role", claims.Role)
		c.Set("session_id", claims.SessionID.String())

		// Continue to the next handler
		c.Next()
	}
}
