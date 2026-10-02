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
	mediaHttp "github.com/hextec/yolyolakay-backend/internal/media/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/media/repository"
	"github.com/hextec/yolyolakay-backend/internal/media/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/config"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"

	"net/http"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("media_service")
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

	// 2. Initialize S3 Repository
	s3Repo, err := repository.NewS3MediaRepository(
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 repository: %v", err)
	}

	// 3. Initialize UseCase
	mediaUC := usecase.NewMediaUseCase(s3Repo)

	// 4. Initialize HTTP Handler
	handler := mediaHttp.NewHandler(mediaUC)

	// 5. Setup Gin Router
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("media_service"))

	// Simple internal auth middleware for inter-service communication
	authMiddleware := middleware.InternalAuth()

	// Initialize routes
	handler.InitRoutes(r, authMiddleware)

	// 6. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8093"
	}

	log.Printf("Media Service running on port %s", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	log.Printf("media_service running on %s", ":"+port)
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
