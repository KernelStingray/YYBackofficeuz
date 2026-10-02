package main

import (
	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"net/http"
	"os"
	"strings"
	"time"

	notificationHttp "github.com/hextec/yolyolakay-backend/internal/notification/delivery/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/hextec/yolyolakay-backend/docs/notification_service"
	amqpConsumer "github.com/hextec/yolyolakay-backend/internal/notification/delivery/amqp"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/hextec/yolyolakay-backend/pkg/integration/fcm"
	"github.com/hextec/yolyolakay-backend/internal/notification/repository"
	"github.com/hextec/yolyolakay-backend/internal/notification/usecase"
	userPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Notification & Chat Service API
// @version 1.0
// @description API for real-time chat and push notifications.
// @host localhost:8086
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("notification_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	dbURL := os.Getenv("DB_A_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_a:5432/yolyolakay_notifications?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database yolyolakay_notifications: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to yolyolakay_notifications on Cluster A")

	// Redis Connect
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisPass := os.Getenv("REDIS_PASS")
	redisClient, err := cache.NewRedisClient(redisAddr, redisPass, 0)
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis in notification_service")

	// Dependencies
	notifRepo := repository.NewPostgresNotificationRepository(pool)
	tokenRepo := repository.NewRedisDeviceTokenRepository(redisClient.GetRawClient())

	// Init RabbitMQ
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://yolyolakay:secret@rabbitmq:5672/"
	}
	rmqClient := rabbitmq.NewClient(rmqURL)
	defer rmqClient.Close()
	log.Println("RabbitMQ client initialized in notification_service")

	// Init gRPC User Client
	userGrpcAddr := os.Getenv("USER_GRPC_ADDR")
	if userGrpcAddr == "" {
		userGrpcAddr = "user_service:50052"
	}
	userConn, err := grpc.Dial(userGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user grpc: %v", err)
	}
	defer userConn.Close()
	userClient := userPb.NewUserServiceClient(userConn)

	// Initialize Push Sender (FCM)
	fcmCreds := os.Getenv("FCM_CREDENTIALS_FILE")
	if fcmCreds == "" {
		fcmCreds = "serviceAccountKey.json"
	}
	var pushSender domain.PushSender
	if sender := fcm.NewPushSender(fcmCreds); sender != nil {
		pushSender = sender
	}

	notifUC := usecase.NewNotificationUseCase(notifRepo, tokenRepo, redisClient.GetRawClient(), rmqClient, pushSender, userClient)

	// Init RabbitMQ Consumer
	consumer := amqpConsumer.NewConsumer(rmqClient, notifUC)
	consumer.Start()

	allowedWSOriginsEnv := os.Getenv("ALLOWED_WS_ORIGINS")
	var allowedWSOrigins []string
	if allowedWSOriginsEnv != "" {
		allowedWSOrigins = strings.Split(allowedWSOriginsEnv, ",")
	}

	httpCfg := notificationHttp.Config{
		VapidPublicKey:   os.Getenv("VAPID_PUBLIC_KEY"),
		AllowedWSOrigins: allowedWSOrigins,
	}

	notifHandler := notificationHttp.NewHandler(notifUC, httpCfg)

	// Router setup
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("notification_service"))
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

	// Load Public Key for JWT validation
	authMiddleware := middleware.InternalAuth()

	// API Routes
	api := r.Group("/api/v1")
	{
		notifications := api.Group("/notifications")
		notifications.Use(authMiddleware)
		{
			notifications.POST("/tokens", notifHandler.RegisterToken)
			notifications.GET("/history", notifHandler.GetHistory)
			notifications.GET("/chats", notifHandler.GetChats)
			notifications.GET("", notifHandler.GetSystemNotifications)
			notifications.PUT("/:id/read", notifHandler.MarkSystemNotificationAsRead)
			notifications.GET("/vapid-key", notifHandler.GetVapidKey)
			notifications.GET("/admin/online-status", notifHandler.GetAdminOnlineStatuses)
		}

		chats := api.Group("/chats")
		chats.Use(authMiddleware)
		{
			chats.GET("", notifHandler.GetChats)
			chats.GET("/:peer_id/messages", notifHandler.GetChatMessages)
		}
	}

	// WebSocket Route
	r.GET("/ws", authMiddleware, notifHandler.WSHandler)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8086",
		Handler: r,
	}
	log.Printf("notification_service running on port :8086")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
