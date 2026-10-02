package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	userHttp "github.com/hextec/yolyolakay-backend/internal/user/delivery/http"

	"os"

	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	userGrpcHandler "github.com/hextec/yolyolakay-backend/internal/user/delivery/grpc"
	userGrpc "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"

	amqp_delivery "github.com/hextec/yolyolakay-backend/internal/user/delivery/amqp"
	"github.com/hextec/yolyolakay-backend/internal/user/repository"
	"github.com/hextec/yolyolakay-backend/internal/user/repository/postgres"
	"github.com/hextec/yolyolakay-backend/internal/user/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"

	_ "github.com/hextec/yolyolakay-backend/docs/user_service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title Yolyolakay User Service API
// @version 1.0
// @description This is the user management microservice for Yolyolakay Platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@hextec.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("user_service")
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
	appLogger.Info("Starting User Service", "port", cfg.Port)

	// 1. Initialize Database Connection (Cluster A)
	pool, err := database.NewPostgresPool(context.Background(), cfg.DB_A_URL)
	if err != nil {
		log.Fatalf("Failed to connect to Cluster A (users db): %v", err)
	}
	defer pool.Close()

	// 2. Instantiate SQLC Queries
	queries := postgres.New(pool)

	// 3. Instantiate Repositories
	userRepo := repository.NewUserRepository(queries)
	carRepo := repository.NewCarRepository(queries)

	// 4. Initialize Redis Cache
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPass, 0)
	if err != nil {
		appLogger.Error("Failed to connect to Redis", "error", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// 6. Initialize RabbitMQ and Consumers
	amqpClient := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer amqpClient.Close()
	amqpConsumer := amqp_delivery.NewUserEventsConsumer(amqpClient, userRepo, redisClient)
	if err := amqpConsumer.Start(context.Background()); err != nil {
		slog.Error("failed to start user events consumer", "error", err)
	}

	// 6. Instantiate UseCase
	userUseCase := usecase.NewUserUseCase(userRepo, carRepo, amqpClient)

	// 6. Setup Gin HTTP Router
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("user_service"))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "service": "user-service"})
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

	// Load Public Key for JWT validation
	authMiddleware := middleware.InternalAuth()

	api := r.Group("/api/v1")

	userHttp.NewUserHandler(api, userUseCase, authMiddleware)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("failed to listen on port 50052: %v", err)
		}
		grpcServer := grpc.NewServer()
		userGrpc.RegisterUserServiceServer(grpcServer, userGrpcHandler.NewUserGrpcServer(userUseCase))
		log.Printf("User gRPC server listening on :50052")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 7. Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	log.Printf("user_service running on %s", ":"+cfg.Port)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
