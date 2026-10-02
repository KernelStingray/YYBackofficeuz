package main

import (
	ridepb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/ride"
	userpb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	"github.com/hextec/yolyolakay-backend/pkg/logger"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"
	"google.golang.org/grpc"

	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"os"
	"strings"

	"github.com/gin-gonic/gin"
	ratingHttp "github.com/hextec/yolyolakay-backend/internal/rating/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/rating/repository"
	"github.com/hextec/yolyolakay-backend/internal/rating/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"

	"net/http"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)


func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("rating_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// 1. Load config
	cfg := config.LoadConfig()
	if err := logger.InitSentry(os.Getenv("SENTRY_DSN"), os.Getenv("ENV")); err != nil {
		log.Printf("failed to init sentry: %v", err)
	}
	defer logger.FlushSentry(2 * time.Second)

	// 2. Connect to Database B
	dbConnString := os.Getenv("DB_B_URL")
	if dbConnString == "" {
		log.Fatal("DB_B_URL is not set")
	}
	db, err := database.NewPostgresPool(context.Background(), dbConnString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migration for ratings (ideally this should be in migrate tool,
	// but for MVP we can execute a CREATE TABLE IF NOT EXISTS here or assume it's created).
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS feedbacks (
			id UUID PRIMARY KEY,
			ride_id UUID NOT NULL,
			reviewer_id UUID NOT NULL,
			reviewee_id UUID NOT NULL,
			score INT NOT NULL,
			comment TEXT,
			tags TEXT[] DEFAULT '{}',
			created_at TIMESTAMP NOT NULL
		);
		ALTER TABLE feedbacks ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
	`)
	if err != nil {
		log.Fatalf("Failed to run schema migration: %v", err)
	}

	// 3. Connect to RabbitMQ
	amqpConn := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer amqpConn.Close()

	// 4. Initialize layers
	repo := repository.NewPostgresRatingRepository(db)

	rideGrpcURL := os.Getenv("RIDE_GRPC_URL")
	if rideGrpcURL == "" {
		rideGrpcURL = "ride_service:50053"
	}
	rideGrpcURL = cleanGrpcURL(rideGrpcURL)
	rideConn, err := grpc.Dial(rideGrpcURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial ride service: %v", err)
	}
	defer rideConn.Close()
	rideClient := ridepb.NewRideServiceClient(rideConn)

	userGrpcURL := os.Getenv("USER_GRPC_URL")
	if userGrpcURL == "" {
		userGrpcURL = cfg.UserServiceURL
	}
	if userGrpcURL == "" {
		userGrpcURL = "user_service:50052"
	}
	userGrpcURL = cleanGrpcURL(userGrpcURL)
	userConn, err := grpc.Dial(userGrpcURL, grpc.WithInsecure())
	if err != nil {
		log.Printf("Warning: Failed to dial user service: %v", err)
	}
	var userClient userpb.UserServiceClient
	if userConn != nil {
		defer userConn.Close()
		userClient = userpb.NewUserServiceClient(userConn)
	}

	uc := usecase.NewRatingUseCase(repo, amqpConn, rideClient, userClient)
	handler := ratingHttp.NewHandler(uc)


	// 5. Setup Gin Router
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("rating_service"))
	authMiddleware := middleware.InternalAuth()
	handler.InitRoutes(r, authMiddleware)

	// 6. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8094"
	}

	log.Printf("Rating Service running on port %s", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	log.Printf("rating_service running on %s", ":"+port)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}

func cleanGrpcURL(addr string) string {
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	return strings.TrimRight(addr, "/")
}
