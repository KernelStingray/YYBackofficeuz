package main

import (
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"github.com/hextec/yolyolakay-backend/pkg/vault"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	_ "github.com/hextec/yolyolakay-backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Backend API
// @version 2.0
// @description Единая API документация для всех микросервисов платформы YolYolakay — райдшеринг, посылки, биллинг, телематика и др.
// @termsOfService https://yolyolakay.uz/terms

// @contact.name HEXTEC Group
// @contact.url https://hextec.uz
// @contact.email support@hextec.uz

// @license.name Proprietary
// @license.url https://yolyolakay.uz/license

// @host api.yolyolakay.com
// @BasePath /
// @schemes https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer токен. Формат: "Bearer {token}"
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("api_gateway")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("[AUDIT ALERT] Panic recovered: %v - IP: %s, URL: %s", recovered, c.ClientIP(), c.Request.URL.Path)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}))
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("api_gateway"))

	// Load Config
	cfg := config.LoadConfig()
	if err := logger.InitSentry(os.Getenv("SENTRY_DSN"), os.Getenv("ENV")); err != nil {
		log.Printf("failed to init sentry: %v", err)
	}
	defer logger.FlushSentry(2 * time.Second)

	// Connect to Redis for Rate Limiting
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       0,
	})

	// Global Rate Limiting: 100 requests per minute
	r.Use(middleware.RateLimit(rdb, 100, 60*time.Second))

	// Strict Rate Limiting on sensitive authentication & OTP endpoints
	authSendOTPLimit := middleware.RateLimitWithPrefix(rdb, "auth_send_otp", 5, 60*time.Second)
	authVerifyOTPLimit := middleware.RateLimitWithPrefix(rdb, "auth_verify_otp", 10, 60*time.Second)
	authLoginLimit := middleware.RateLimitWithPrefix(rdb, "auth_login", 10, 60*time.Second)
	adminLoginLimit := middleware.RateLimitWithPrefix(rdb, "admin_login", 5, 60*time.Second)
	routingRouteLimit := middleware.RateLimitWithPrefix(rdb, "routing_route", 30, 60*time.Second)
	routingGeocodeLimit := middleware.RateLimitWithPrefix(rdb, "routing_geocode", 60, 60*time.Second)

	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/routing/route") {
			routingRouteLimit(c)
			return
		}
		if strings.HasPrefix(path, "/api/v1/routing/geocode") || strings.HasPrefix(path, "/api/v1/routing/reverse-geocode") {
			routingGeocodeLimit(c)
			return
		}
		if strings.HasPrefix(path, "/api/v1/routing") {
			routingRouteLimit(c)
			return
		}
		switch path {
		case "/api/v1/auth/send-otp":
			authSendOTPLimit(c)
		case "/api/v1/auth/verify-otp":
			authVerifyOTPLimit(c)
		case "/api/v1/auth/login":
			authLoginLimit(c)
		case "/api/v1/backoffice/admins/login":
			adminLoginLimit(c)
		}
	})

	// Strict CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Temporarily commented out Vercel preview deployments
			if strings.HasPrefix(origin, "https://yolyolakay-") && strings.HasSuffix(origin, ".vercel.app") {
				return true
			}
			allowed := map[string]bool{
				"http://localhost:3000":                 true,
				"http://localhost:3001":                 true,
				"http://localhost:3002":                 true,
				"http://localhost:3003":                 true,
				"https://yolyolakay.com":                true,
				"https://www.yolyolakay.com":            true,
				"https://app.yolyolakay.com":            true,
				"https://api.yolyolakay.com":            true,
				"https://admin.yolyolakay.com":          true,
				"https://backoffice.yolyolakay.com":     true,
				"https://yolyolakay.uz":                 true,
				"https://www.yolyolakay.uz":             true,
				"https://yolyolakay.co.uz":              true,
				"https://www.yolyolakay.co.uz":          true,
				"http://yolyolakay.co.uz":               true,
				"http://www.yolyolakay.co.uz":           true,
				"https://yolyolakay-frontend.vercel.app": true,
			}
			if allowed[origin] {
				return true
			}
			customOrigins := os.Getenv("ALLOWED_CORS_ORIGINS")
			if customOrigins == "" {
				customOrigins = os.Getenv("ALLOWED_ORIGINS")
			}
			if customOrigins != "" {
				for _, o := range strings.Split(customOrigins, ",") {
					if strings.TrimSpace(o) == origin {
						return true
					}
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Device-ID", "X-OS-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Security Headers Middleware (Helmet-like)
	r.Use(secure.New(secure.Config{
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
	}))

	// Apply OpenTelemetry Middleware

	// Apply Prometheus Metrics Middleware
	r.Use(middleware.MetricsMiddleware())

	// 1. Load Public Key from Vault
	publicKey, err := vault.LoadPublicKey()
	if err != nil {
		log.Fatalf("Failed to load public key from Vault: %v", err)
	}

	// 2. Initialize JWT Manager (no private key needed for validation)
	jwtManager := jwt.NewJWTManager(nil, publicKey)

	// 3. Define Public Routes (paths that don't need token)
	publicRoutes := []string{
		"/metrics",
		"/health",
		"/api/v1/auth/send-otp",
		"/api/v1/auth/verify-otp",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/backoffice/admins/login",
		"/api/v1/backoffice/configs/*",
		"/api/v1/backoffice/configs",
		"/api/v1/config/*",
		"/api/v1/config",
		"/api/v1/docs/*",
		"GET:/api/v1/media/*",
		"GET:/api/v1/rides/search",
		"GET:/api/v1/rides/recommendations",
		"GET:/api/v1/rides/driver/:uuid",
		"GET:/api/v1/rides/:uuid/location",
		"GET:/api/v1/rides/:uuid",
		"GET:/api/v1/users/:uuid/rides",
		"GET:/api/v1/vehicles/:uuid",
		"GET:/api/v1/rating/*",
		"GET:/api/v1/rating",
		"GET:/api/v1/routing/*",
		"GET:/api/v1/routing",
		"GET:/api/v1/billing/plans",
		"POST:/api/v1/billing/click/*",
		"GET:/api/v1/billing/click/*",
		"POST:/api/v1/billing/payme/*",
		"GET:/api/v1/billing/payme/*",
	}


	// 4. Apply GatewayAuth Middleware
	r.Use(middleware.GatewayAuth(jwtManager, publicRoutes))

	// Prometheus Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api_gateway",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Swagger Documentation with Basic Auth
	swaggerUser := os.Getenv("SWAGGER_USER")
	if swaggerUser == "" {
		swaggerUser = "admin"
	}
	swaggerPass := os.Getenv("SWAGGER_PASSWORD")
	if swaggerPass == "" {
		swaggerPass = "admin"
	}

	docsGroup := r.Group("/api/v1/docs")
	docsGroup.Use(gin.BasicAuth(gin.Accounts{
		swaggerUser: swaggerPass,
	}))
	docsGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Remote Config Endpoint
	r.GET("/api/v1/config/:key", func(c *gin.Context) {
		key := c.Param("key")
		cacheKey := "config:" + key

		// Try Redis first
		val, err := rdb.Get(c.Request.Context(), cacheKey).Result()
		if err == nil && val != "" {
			c.Data(http.StatusOK, "application/json", []byte(val))
			return
		}

		// Cache miss, fallback to backoffice service
		backofficeURL := cfg.BackofficeServiceURL + "/api/v1/backoffice/configs/" + key
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(backofficeURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch remote config"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "config not found"})
			return
		}

		// Read the body completely so we can cache it
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}

		// Set to Redis with TTL 10 minutes (ignoring error here since it's just a cache)
		_ = rdb.Set(c.Request.Context(), cacheKey, string(bodyBytes), 10*time.Minute).Err()

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
	})

	// Setup Reverse Proxies dynamically based on config
	ServiceRoutes := map[string]string{
		"/api/v1/users":         cfg.UserServiceURL,
		"/api/v1/vehicles":      cfg.UserServiceURL,
		"/api/v1/auth":          cfg.AuthServiceURL,
		"/api/v1/parcels":       cfg.ParcelServiceURL,
		"/api/v1/telematics":    cfg.TelematicsServiceURL,
		"/api/v1/billing":       cfg.BillingServiceURL,
		"/api/v1/rides":         cfg.RideServiceURL,
		"/api/v1/bookings":      cfg.RideServiceURL,
		"/api/v1/notifications": cfg.NotificationServiceURL,
		"/api/v1/chats":         cfg.NotificationServiceURL,
		"/ws":                   cfg.NotificationServiceURL, // WebSockets for notification service
		"/api/v1/routing":       cfg.RoutingServiceURL,
		"/api/v1/fraud":         cfg.FraudServiceURL,
		"/api/v1/helpdesk":      cfg.HelpdeskServiceURL,
		"/api/v1/analytics":     cfg.AnalyticsServiceURL,
		"/api/v1/backoffice":    cfg.BackofficeServiceURL,
		"/api/v1/media":         cfg.MediaServiceURL,
		"/api/v1/rating":        cfg.RatingServiceURL,
	}

	// Proxy to Ride Service for routing /api/v1/users/:id/rides
	var rideProxy *httputil.ReverseProxy
	if rideURL, err := url.Parse(cfg.RideServiceURL); err == nil {
		rp := httputil.NewSingleHostReverseProxy(rideURL)
		rp.FlushInterval = 100 * time.Millisecond
		origDirector := rp.Director
		rp.Director = func(req *http.Request) {
			origDirector(req)
			if cfg.InternalAPISecret != "" {
				req.Header.Set("X-Internal-Secret", cfg.InternalAPISecret)
			}
		}
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for ride service: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Service Unavailable"))
		}
		rideProxy = rp
	}

	for prefix, targetURL := range ServiceRoutes {
		setupProxy(r, prefix, targetURL, cfg.InternalAPISecret, rideProxy)
	}

	log.Printf("Starting API Gateway on port %s...", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	log.Printf("api_gateway running on %s", ":"+port)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}

func setupProxy(r *gin.Engine, prefix, targetURL, internalSecret string, rideProxy *httputil.ReverseProxy) {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target URL for prefix %s: %v", prefix, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = 100 * time.Millisecond

	// Modify request to add internal secret
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if internalSecret != "" {
			req.Header.Set("X-Internal-Secret", internalSecret)
		}
	}

	// Custom Error Handler for Proxy
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Service Unavailable"))
	}

	r.Any(prefix+"/*path", func(c *gin.Context) {
		if prefix == "/api/v1/users" && rideProxy != nil && c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/rides") {
			parts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
			// parts: ["api", "v1", "users", "<id>", "rides"]
			if len(parts) == 5 && parts[2] == "users" && parts[4] == "rides" {
				c.Request.URL.Path = "/api/v1/rides/driver/" + parts[3]
				c.Request.URL.RawPath = ""
				rideProxy.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Also handle the exact prefix without trailing slash
	r.Any(prefix, func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
