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
	_ "github.com/hextec/yolyolakay-backend/docs/routing_service"
	routingHttp "github.com/hextec/yolyolakay-backend/internal/routing/delivery/http"
	"github.com/hextec/yolyolakay-backend/internal/routing/provider"
	"github.com/hextec/yolyolakay-backend/internal/routing/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hextec/yolyolakay-backend/pkg/graceful"
)

// @title YolYolakay Routing & Maps Service API
// @version 1.0
// @description Microservice for finding optimal routes and ETAs using OSRM.
// @host localhost:8088
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("routing_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// 0. Initialize Redis Cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		redisPass = os.Getenv("REDIS_PASS")
	}
	redisClient, err := cache.NewRedisClient(redisAddr, redisPass, 0)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer redisClient.Close()

	// 1. Initialize OSRM Provider
	osrmURL := os.Getenv("OSRM_URL")
	if osrmURL == "" {
		osrmURL = "http://osrm:5000"
	}
	osrmProvider := provider.NewOSRMProvider(osrmURL)

	// 2. Initialize Geocoding Provider
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}
	geocodingProvider := provider.NewNominatimProvider(nominatimURL)

	// 3. Initialize UseCase
	routingUC := usecase.NewRoutingUseCase(osrmProvider, geocodingProvider, redisClient)

	// 3. Initialize HTTP Handler
	routingHandler := routingHttp.NewHandler(routingUC)

	// 4. Setup Gin Router
	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("routing_service"))
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

	api := r.Group("/api/v1")
	{
		routing := api.Group("/routing")
		{
			routing.GET("/route", routingHandler.GetRoute)
			routing.GET("/geocode", routingHandler.Geocode)
			routing.GET("/reverse-geocode", routingHandler.ReverseGeocode)
		}
	}

	// 5. Start Server
	srv := &http.Server{
		Addr:    ":8088",
		Handler: r,
	}
	log.Printf("routing_service running on port :8088")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
