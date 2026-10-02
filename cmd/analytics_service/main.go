package main

import (
	"context"
	"log"

	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"os"

	"github.com/gin-gonic/gin"
	analyticsHttp "github.com/hextec/yolyolakay-backend/internal/analytics/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/analytics/repository"
	"github.com/hextec/yolyolakay-backend/internal/analytics/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/hextec/yolyolakay-backend/docs/analytics_service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title Analytics Service API
// @version 1.0
// @description DWH & BI Analytics service for Yolyolakay

// @host localhost:8091
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("analytics_service")
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
		dbURL = "postgresql://postgres:password@postgres_b:5432/yolyolakay_analytics?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresAnalyticsRepository(pool)

	// Init RabbitMQ
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://yolyolakay:secret@rabbitmq:5672/"
	}
	rmqClient := rabbitmq.NewClient(rmqURL)
	defer rmqClient.Close()
	log.Println("RabbitMQ client initialized in analytics_service")

	uc := usecase.NewAnalyticsUseCase(repo, rmqClient)
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("analytics_service"))
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

	analyticsHttp.NewAnalyticsHandler(r, uc, authMiddleware)

	log.Println("Starting Analytics service on :8091")
	srv := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}
	log.Printf("analytics_service running on %s", ":8091")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
