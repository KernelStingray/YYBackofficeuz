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
	_ "github.com/hextec/yolyolakay-backend/docs/helpdesk_service"
	helpdeskHttp "github.com/hextec/yolyolakay-backend/internal/helpdesk/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/helpdesk/repository"
	"github.com/hextec/yolyolakay-backend/internal/helpdesk/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Helpdesk & Support API
// @version 1.0
// @description Microservice for customer support and ticketing.
// @host localhost:8090
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("helpdesk_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// 1. Connect to Database (yolyolakay_helpdesk is in Cluster A)
	dbURL := os.Getenv("DB_A_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@postgres_a:5432/yolyolakay_helpdesk?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// 2. Initialize Layers
	repo := repository.NewPostgresHelpdeskRepository(pool)
	uc := usecase.NewHelpdeskUseCase(repo)
	handler := helpdeskHttp.NewHandler(uc)

	// 3. Setup Router
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("helpdesk_service"))
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
		helpdesk := api.Group("/helpdesk")
		helpdesk.Use(authMiddleware)
		{
			helpdesk.POST("/tickets", handler.CreateTicket)
			helpdesk.GET("/tickets", handler.GetUserTickets)
			helpdesk.POST("/tickets/resolve", handler.ResolveTicket)
		}
	}

	// 4. Start Server
	srv := &http.Server{
		Addr:    ":8090",
		Handler: r,
	}
	log.Printf("helpdesk_service running on port :8090")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
