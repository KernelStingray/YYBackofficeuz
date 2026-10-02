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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"

	_ "github.com/hextec/yolyolakay-backend/docs/telematics_service"
	telematicsHttp "github.com/hextec/yolyolakay-backend/internal/telematics/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/telematics/repository"
	"github.com/hextec/yolyolakay-backend/internal/telematics/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Telematics Service API
// @version 1.0
// @description Highload service for real-time tracking, ETA and map matching.
// @host localhost:8087
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("telematics_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// Postgres connection (Cold Data)
	dbURL := os.Getenv("DB_B_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_b:5432/yolyolakay_telematics?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database yolyolakay_telematics: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to yolyolakay_telematics on Cluster B")

	// Redis connection (Hot Data)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		redisPass = os.Getenv("REDIS_PASS")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,  // use default DB
	})

	ctxRedis, cancelRedis := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelRedis()
	if err := redisClient.Ping(ctxRedis).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis (Hot Data Cache)")

	// Connect to Rides DB on Cluster B (senna) for RBAC passenger verification
	ridesDbURL := os.Getenv("RIDES_DB_URL")
	if ridesDbURL == "" {
		ridesDbURL = strings.Replace(dbURL, "/pascal", "/senna", 1)
		ridesDbURL = strings.Replace(ridesDbURL, "/yolyolakay_telematics", "/senna", 1)
	}
	ridesPool, err := pgxpool.New(context.Background(), ridesDbURL)
	if err != nil {
		log.Printf("Warning: Unable to connect to rides DB for telematics: %v", err)
	} else {
		defer ridesPool.Close()
	}

	// Dependencies
	telematicsRepo := repository.NewTelematicsRepository(redisClient, pgPool, ridesPool)

	// Batch worker configuration: flush every 1000 items or 5 seconds
	batchWorker := usecase.NewBatchWorker(telematicsRepo, 1000, 5*time.Second)
	defer batchWorker.Stop()

	telematicsUC := usecase.NewTelematicsUseCase(telematicsRepo, batchWorker)
	telematicsHandler := telematicsHttp.NewHandler(telematicsUC)

	// Router setup
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("telematics_service"))
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
		telematics := api.Group("/telematics")
		telematics.Use(authMiddleware)
		{
			telematics.POST("/ping", telematicsHandler.Ping)
			telematics.GET("/nearby", telematicsHandler.GetNearby)
			telematics.GET("/drivers/:id/location", telematicsHandler.GetDriverLocation)
			telematics.GET("/location", telematicsHandler.GetLocationByDriverID)
			telematics.GET("/eta", telematicsHandler.CalculateETA)
			telematics.GET("/admin/locations", telematicsHandler.GetAdminLocations)
		}
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8087",
		Handler: r,
	}
	log.Printf("telematics_service running on port :8087")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
