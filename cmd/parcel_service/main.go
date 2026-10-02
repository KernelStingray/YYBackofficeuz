package main

import (
	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/hextec/yolyolakay-backend/docs/parcel_service"
	parcelAmqp "github.com/hextec/yolyolakay-backend/internal/parcel/delivery/amqp"
	parcelHttp "github.com/hextec/yolyolakay-backend/internal/parcel/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/parcel/repository"
	"github.com/hextec/yolyolakay-backend/internal/parcel/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	billingPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/billing"
	userPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/hextec/yolyolakay-backend/pkg/graceful"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title YolYolakay Parcel Service API
// @version 1.0
// @description API for managing blind logistics and parcels.
// @host localhost:8085
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("parcel_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	dbURL := os.Getenv("DB_B_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_b:5432/yolyolakay_parcels?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database yolyolakay_parcels: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to yolyolakay_parcels on Cluster B")

	// gRPC Client setup
	billingURL := os.Getenv("BILLING_GRPC_URL")
	if billingURL == "" {
		billingURL = "billing_service:50051"
	}
	conn, err := grpc.Dial(billingURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to billing grpc: %v", err)
	}
	defer conn.Close()
	billingClient := billingPb.NewBillingServiceClient(conn)

	// User gRPC Client setup
	userURL := os.Getenv("USER_GRPC_URL")
	if userURL == "" {
		userURL = "user_service:50052"
	}
	userConn, err := grpc.Dial(userURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to user grpc: %v", err)
	}
	defer userConn.Close()
	userClient := userPb.NewUserServiceClient(userConn)

	// RabbitMQ setup
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	amqpClient := rabbitmq.NewClient(rabbitURL)
	defer amqpClient.Close()
	publisher := parcelAmqp.NewEventPublisher(amqpClient)

	// Dependencies
	parcelRepo := repository.NewPostgresParcelRepository(pool)
	parcelUC := usecase.NewParcelUseCase(parcelRepo, billingClient, publisher)
	parcelHandler := parcelHttp.NewHandler(parcelUC, userClient)

	// Router setup
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("parcel_service"))
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
		parcels := api.Group("/parcels")
		parcels.Use(authMiddleware)
		{
			parcels.POST("", parcelHandler.CreateParcel)
			parcels.POST("/", parcelHandler.CreateParcel)
			parcels.GET("/search", parcelHandler.SearchParcels)
			// Admin routes
			parcels.GET("/admin/stats", parcelHandler.GetAdminStats)
			parcels.GET("/admin/user/:id/stats", parcelHandler.GetUserStats)
			parcels.GET("/admin/list", parcelHandler.GetAdminParcels)
			parcels.POST("/:id/accept", parcelHandler.AcceptParcel)
			parcels.GET("/my", parcelHandler.GetMyParcels)
			parcels.GET("/:id", parcelHandler.GetParcelByID)
			parcels.POST("/:id/respond", parcelHandler.RespondToParcel)
			parcels.GET("/:id/responses", parcelHandler.GetParcelResponses)
			parcels.POST("/:id/responses/:response_id/accept", parcelHandler.AcceptParcelResponse)
			parcels.POST("/:id/responses/:response_id/reject", parcelHandler.RejectParcelResponse)
			parcels.POST("/:id/cancel", parcelHandler.CancelParcel)
			parcels.POST("/:id/pickup", parcelHandler.PickupParcel)
			parcels.POST("/:id/deliver", parcelHandler.DeliverParcel)
			parcels.POST("/:id/confirm-delivery", parcelHandler.ConfirmDelivery)
		}
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8085",
		Handler: r,
	}
	log.Printf("parcel_service running on port :8085")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
