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
	_ "github.com/hextec/yolyolakay-backend/docs/antifraud_service"
	antifraudHttp "github.com/hextec/yolyolakay-backend/internal/antifraud/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/antifraud/repository"
	"github.com/hextec/yolyolakay-backend/internal/antifraud/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Anti-Fraud & Scoring API
// @version 1.0
// @description Microservice for trust scoring and malicious behavior detection.
// @host localhost:8089
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("antifraud_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// 1. Connect to Database (yolyolakay_fraud is in Cluster A, DB_A)
	dbURL := os.Getenv("DB_A_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_a:5432/yolyolakay_fraud?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// 2. Initialize Layers
	repo := repository.NewPostgresAntiFraudRepository(pool)
	uc := usecase.NewAntiFraudUseCase(repo)
	handler := antifraudHttp.NewHandler(uc)

	// 3. Setup Router
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("antifraud_service"))
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
	{
		fraud := api.Group("/fraud")
		fraud.Use(authMiddleware)
		{
			fraud.GET("/score", handler.GetTrustScore)
			fraud.POST("/report", handler.ReportMockLocation)
		}
	}

	// 4. Start Server
	srv := &http.Server{
		Addr:    ":8089",
		Handler: r,
	}
	log.Printf("antifraud_service running on port :8089")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
