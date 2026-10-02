package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
)

// GatewayAuth creates a Gin handler for the API Gateway.
// It checks if the route is public, otherwise it validates the JWT and injects headers.
func GatewayAuth(jwtManager *jwt.JWTManager, publicRoutes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Security: mock-otp is NEVER public without explicit valid X-Test-Secret
		if strings.HasPrefix(path, "/api/v1/auth/mock-otp") {
			testSecret := os.Getenv("TEST_SECRET")
			if testSecret == "" {
				testSecret = os.Getenv("INTERNAL_API_SECRET")
			}
			if testSecret == "" {
				testSecret = "super-internal-secret-for-microservices"
			}
			clientSecret := c.GetHeader("X-Test-Secret")
			if clientSecret != "" && clientSecret == testSecret {
				c.Next()
				return
			}
			slog.Warn("mock-otp blocked: missing or invalid test secret", "ip", c.ClientIP(), "endpoint", path)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		isPublic := false
		// 1. Check if the path is in the public routes list
		for _, publicRoute := range publicRoutes {
			if matchesPublicRoute(publicRoute, c.Request.Method, path) {
				isPublic = true
				break
			}
		}

		// 2. Extract Authorization Header
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				if isPublic {
					c.Next()
					return
				}
				slog.Warn("gateway auth failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "invalid authorization header format")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				return
			}
			tokenString = parts[1]
		} else {
			// Fallback to query parameter (useful for WebSockets)
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			if isPublic {
				c.Next()
				return
			}

			slog.Warn("gateway auth failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "missing token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header or token query parameter is required"})
			return
		}

		// 3. Validate Token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			if isPublic {
				c.Next()
				return
			}
			slog.Warn("gateway auth failed", "ip", c.ClientIP(), "user_agent", c.GetHeader("User-Agent"), "endpoint", c.Request.URL.Path, "reason", "invalid or expired token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// 4. Inject Headers for downstream microservices
		c.Request.Header.Set("X-User-Id", claims.UserID.String())
		c.Request.Header.Set("X-User-Role", claims.Role)
		c.Request.Header.Set("X-Session-Id", claims.SessionID.String())

		c.Next()
	}
}

func matchesPublicRoute(publicRoute, method, path string) bool {
	routeMethod := ""
	routePath := publicRoute
	if strings.Contains(publicRoute, ":") {
		parts := strings.SplitN(publicRoute, ":", 2)
		routeMethod = parts[0]
		routePath = parts[1]
	}

	if routeMethod != "" && routeMethod != method {
		return false
	}

	// Exact match
	if path == routePath {
		return true
	}

	// Prefix match for wildcards (e.g., /api/v1/docs/* or /api/v1/media/*)
	if strings.HasSuffix(routePath, "*") {
		prefix := strings.TrimSuffix(routePath, "*")
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// Path parameter match (e.g., /api/v1/rides/:uuid or /api/v1/users/:id)
	if strings.Contains(routePath, ":") {
		patternParts := strings.Split(strings.Trim(routePath, "/"), "/")
		pathParts := strings.Split(strings.Trim(path, "/"), "/")

		if len(patternParts) == len(pathParts) {
			matched := true
			for i := 0; i < len(patternParts); i++ {
				if patternParts[i] == ":uuid" {
					if _, err := uuid.Parse(pathParts[i]); err != nil {
						matched = false
						break
					}
				} else if strings.HasPrefix(patternParts[i], ":") {
					if pathParts[i] == "" {
						matched = false
						break
					}
				} else if patternParts[i] != pathParts[i] {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}

	return false
}

