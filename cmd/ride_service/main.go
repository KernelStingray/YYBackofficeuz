package main

import (
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	amqp_delivery "github.com/hextec/yolyolakay-backend/internal/ride/delivery/amqp"
	rideHttp "github.com/hextec/yolyolakay-backend/internal/ride/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/ride/integration"
	"github.com/hextec/yolyolakay-backend/internal/ride/repository"
	"github.com/hextec/yolyolakay-backend/internal/ride/usecase"
	"github.com/hextec/yolyolakay-backend/internal/ride/worker"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	_ "github.com/hextec/yolyolakay-backend/docs/ride_service"

	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"

	rideGrpc "github.com/hextec/yolyolakay-backend/internal/ride/delivery/grpc"
	ridepb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/ride"
	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title Yolyolakay Ride Service API
// @version 1.0
// @description API for managing rides and bookings in Cluster B.
// @host localhost:8083
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("ride_service")
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

	// 2. Load Public Key from Vault for JWT Verification
	authMiddleware := middleware.InternalAuth()

	// Connect to Cluster B (yolyolakay_rides)
	dbPool, err := pgxpool.New(context.Background(), cfg.DB_B_URL)
	if err != nil {
		log.Fatalf("Unable to connect to database B: %v", err)
	}
	defer dbPool.Close()

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPass, 0)
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Connect to RabbitMQ
	amqpClient := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer amqpClient.Close()

	// Init layers
	repo := repository.NewPostgresRideRepository(dbPool)
	subRepo := repository.NewPostgresRouteSubscriptionRepository(dbPool)
	rideCache := repository.NewRedisRideCache(redisClient)

	// Setup Publisher
	publisher := amqp_delivery.NewEventPublisher(amqpClient)

	// Init gRPC Client
	billingURL := os.Getenv("BILLING_GRPC_URL")
	if billingURL == "" {
		billingURL = "billing_service:50051"
	}
	userURL := os.Getenv("USER_GRPC_URL")
	if userURL == "" {
		userURL = "user_service:50052"
	}

	grpcClient, err := integration.NewGrpcClient(billingURL, userURL)
	if err != nil {
		log.Printf("Failed to initialize gRPC clients: %v. Continuing without RPC.", err)
	} else {
		defer grpcClient.Close()
	}

	rideUseCase := usecase.NewRideUseCase(repo, subRepo, rideCache, grpcClient, publisher)

	// Start gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Printf("Failed to listen for gRPC on port %s: %v", grpcPort, err)
	} else {
		grpcServer := grpc.NewServer()
		ridepb.RegisterRideServiceServer(grpcServer, rideGrpc.NewHandler(rideUseCase))
		go func() {
			log.Printf("Ride Service gRPC running on port %s", grpcPort)
			if err := grpcServer.Serve(lis); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
		defer grpcServer.GracefulStop()
	}

	// Start AMQP Consumer
	rideConsumer := amqp_delivery.NewRideConsumer(amqpClient, repo, rideCache, rideUseCase)
	if err := rideConsumer.Start(context.Background()); err != nil {
		log.Printf("Failed to start ride consumer: %v", err)
	}

	// Start Background Workers
	negWorker := worker.NewNegotiationWorker(rideUseCase)
	go negWorker.Start()
	defer negWorker.Stop()

	remWorker := worker.NewReminderWorker(dbPool, publisher)
	remWorker.Start()
	defer remWorker.Stop()

	handler := rideHttp.NewHandler(rideUseCase)

	router := gin.Default()
	router.Use(middleware.TraceMiddleware())
	router.Use(middleware.SentryMiddleware())
	router.Use(otelgin.Middleware("ride_service"))

	handler.InitRoutes(router, authMiddleware)

	log.Println("Ride Service running on port 8083")
	srv := &http.Server{
		Addr:    ":8083",
		Handler: router,
	}
	log.Printf("ride_service running on %s", ":8083")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
