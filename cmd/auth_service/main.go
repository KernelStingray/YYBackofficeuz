package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/hextec/yolyolakay-backend/docs/auth_service"
	authAmqp "github.com/hextec/yolyolakay-backend/internal/auth/delivery/amqp"
	authHttp "github.com/hextec/yolyolakay-backend/internal/auth/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/auth/domain"
	authSms "github.com/hextec/yolyolakay-backend/pkg/integration/eskiz"
	authRedis "github.com/hextec/yolyolakay-backend/internal/auth/repository/redis"
	authUseCase "github.com/hextec/yolyolakay-backend/internal/auth/usecase"
	userRepoImpl "github.com/hextec/yolyolakay-backend/internal/user/repository"
	userPostgres "github.com/hextec/yolyolakay-backend/internal/user/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/hextec/yolyolakay-backend/pkg/graceful"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"github.com/hextec/yolyolakay-backend/pkg/vault"
	redis "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// @title Yolyolakay Auth Service API
// @version 1.0
// @description Authentication, Passwordless Login and JWT management.
// @host localhost:8082
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("auth_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	cfg := config.LoadConfig()
	if err := logger.InitSentry(os.Getenv("SENTRY_DSN"), os.Getenv("ENV")); err != nil {
		log.Printf("failed to init sentry: %v", err)
	}
	defer logger.FlushSentry(2 * time.Second)
	appLogger := logger.InitLogger(cfg.Env)
	appLogger.Info("Starting Auth Service", "port", cfg.AuthPort)

	// 1. Initialize DB Connection (Cluster A)
	pool, err := database.NewPostgresPool(context.Background(), cfg.DB_A_URL)
	if err != nil {
		log.Fatalf("Failed to connect to Cluster A DB: %v", err)
	}
	defer pool.Close()

	// Initialize SQLC queries for User Service
	queries := userPostgres.New(pool)
	uRepo := userRepoImpl.NewUserRepository(queries)

	// 2. Initialize Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       0, // Use default DB
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	// 3. Instantiate Auth Repositories & Integrations
	otpRepo := authRedis.NewOTPRepository(rdb)
	sessionRepo := authRedis.NewSessionRepository(rdb)

	// Select SMS provider: real Eskiz if credentials are set, mock otherwise
	var smsProvider domain.SMSProvider
	eskizEmail := os.Getenv("ESKIZ_EMAIL")
	eskizPassword := os.Getenv("ESKIZ_PASSWORD")
	eskizAPIURL := os.Getenv("ESKIZ_API_URL")
	if eskizEmail != "" && eskizPassword != "" {
		log.Println("Using Eskiz SMS Provider (production mode)")
		smsProvider = authSms.NewEskizProvider(eskizEmail, eskizPassword, eskizAPIURL)
	} else {
		log.Println("Using Mock SMS Provider (development mode)")
		smsProvider = authSms.NewMockSMSProvider()
	}

	// 4. Load RSA Keys from Vault & Instantiate JWT Manager
	privateKey, err := vault.LoadPrivateKey()
	if err != nil {
		log.Fatalf("Failed to load private key from Vault: %v", err)
	}
	publicKey, err := vault.LoadPublicKey()
	if err != nil {
		log.Fatalf("Failed to load public key from Vault: %v", err)
	}

	jwtManager := jwt.NewJWTManager(privateKey, publicKey)

	// 5. Instantiate Auth UseCase
	// Initialize RabbitMQ
	rmqClient := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer rmqClient.Close()

	useCase := authUseCase.NewAuthUseCase(
		otpRepo,
		sessionRepo,
		smsProvider,
		uRepo,
		jwtManager,
		rmqClient,
	)

	authConsumer := authAmqp.NewAuthEventsConsumer(rmqClient, useCase)
	go func() {
		for {
			if err := authConsumer.Start(context.Background()); err != nil {
				log.Printf("AuthEventsConsumer start failed (retrying in 5s): %v", err)
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("AuthEventsConsumer started successfully")
				break
			}
		}
	}()

	// 6. Setup Gin HTTP Router
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("auth_service"))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "service": "auth-service"})
	})
	swaggerUser := os.Getenv("SWAGGER_USER")
	if swaggerUser == "" {
		swaggerUser = "admin"
	}
	swaggerPass := os.Getenv("SWAGGER_PASSWORD")
	if swaggerPass == "" {
		swaggerPass = "admin"
	}
	swaggerGroup := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		swaggerUser: swaggerPass,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1/auth")
	authMiddleware := middleware.InternalAuth()
	authHttp.NewAuthHandler(api, useCase, authMiddleware)

	// 7. Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.AuthPort,
		Handler: r,
	}
	log.Printf("auth_service running on %s", ":"+cfg.AuthPort)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
