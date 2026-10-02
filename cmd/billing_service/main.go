package main

import (
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"os"

	"net"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	_ "github.com/hextec/yolyolakay-backend/docs/billing_service" // swagger docs
	amqp_delivery "github.com/hextec/yolyolakay-backend/internal/billing/delivery/amqp"
	billingGrpcHandler "github.com/hextec/yolyolakay-backend/internal/billing/delivery/grpc"
	billingHttp "github.com/hextec/yolyolakay-backend/internal/billing/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/billing/integration"
	"github.com/hextec/yolyolakay-backend/internal/billing/repository"
	"github.com/hextec/yolyolakay-backend/internal/billing/usecase"
	asyncWorker "github.com/hextec/yolyolakay-backend/internal/billing/worker"
	billingGrpc "github.com/hextec/yolyolakay-backend/pkg/api/grpc/billing"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/integration/click"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/hibiken/asynq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title Yolyolakay Billing Service API
// @version 1.0
// @description Billing, Payments and Subscriptions for Yolyolakay Platform.
// @host localhost:8084
// @BasePath /
func main() {
	profiler.Start("6063")

	shutdownTracer, err := tracing.InitTracer("billing_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	cfg := config.LoadConfig()
	logger.InitLogger(os.Getenv("LOG_LEVEL"))

	defer logger.FlushSentry(2 * time.Second)

	// Connect to yolyolakay_billing in Cluster B
	dbURL := os.Getenv("DB_C_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_b:5432/yolyolakay_billing?sslmode=disable"
	}
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	repo := repository.NewPostgresBillingRepository(dbPool)

	// Init RabbitMQ
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://yolyolakay:secret@rabbitmq:5672/"
	}
	rmqClient := rabbitmq.NewClient(rmqURL)
	defer rmqClient.Close()
	log.Println("RabbitMQ client initialized in billing_service")

	// Init Payment Gateways
	gateways := map[string]integration.PaymentGateway{
		"click": click.NewClickGateway(cfg.ClickMerchantID, cfg.ClickServiceID, cfg.ClickSecretKey),
	}

	uc := usecase.NewBillingUseCase(repo, rmqClient, gateways)

	// AMQP Consumer
	amqpConsumer := amqp_delivery.NewBillingEventsConsumer(rmqClient, repo, uc)
	go func() {
		for {
			if err := amqpConsumer.Start(context.Background()); err != nil {
				log.Printf("failed to start amqp consumer (retrying in 5s): %v", err)
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("BillingEventsConsumer started successfully")
				break
			}
		}
	}()

	// Init Asynq Redis
	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPass}

	// Start Asynq Processor (Worker)
	taskProcessor := asyncWorker.NewRedisTaskProcessor(redisOpt, repo, rmqClient)
	go func() {
		if err := taskProcessor.Start(); err != nil {
			log.Fatalf("could not start asynq server: %v", err)
		}
	}()

	// Start Asynq Scheduler (Cron)
	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{})

	// Init Asynq Distributor
	taskDistributor := asyncWorker.NewRedisTaskDistributor(redisOpt)

	if err := taskDistributor.RegisterCronTasks(scheduler); err != nil {
		log.Fatalf("could not register cron tasks: %v", err)
	}

	go func() {
		if err := scheduler.Run(); err != nil {
			log.Fatalf("could not start asynq scheduler: %v", err)
		}
	}()

	httpCfg := billingHttp.Config{
		WebhookAllowedIPs: os.Getenv("WEBHOOK_ALLOWED_IPS"),
	}
	handler := billingHttp.NewHandler(uc, taskDistributor, httpCfg)

	router := gin.Default()
	router.Use(middleware.TraceMiddleware())
	router.Use(middleware.SentryMiddleware())
	router.Use(otelgin.Middleware("billing_service"))

	// Load Public Key for JWT validation
	authMiddleware := middleware.InternalAuth()

	handler.InitRoutes(router, authMiddleware)

	// Swagger endpoint
	swaggerUser := os.Getenv("SWAGGER_USER")
	if swaggerUser == "" {
		swaggerUser = "admin"
	}
	swaggerPass := os.Getenv("SWAGGER_PASSWORD")
	if swaggerPass == "" {
		swaggerPass = "admin"
	}
	swaggerGroup := router.Group("/swagger", gin.BasicAuth(gin.Accounts{
		swaggerUser: swaggerPass,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := cfg.Port // By default we can override this using env var or hardcode for now
	// To avoid conflict, let's hardcode 8084 if env is not strictly providing it for billing,
	// but better practice is:
	// port := os.Getenv("BILLING_PORT")
	port = "8084"

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen on port 50051: %v", err)
		}
		grpcServer := grpc.NewServer()
		billingGrpc.RegisterBillingServiceServer(grpcServer, billingGrpcHandler.NewBillingGrpcServer(uc))
		log.Printf("Billing gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Billing Service running on port %s", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("billing_service running on %s", ":"+port)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
