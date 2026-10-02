package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/hextec/yolyolakay-backend/pkg/profiler"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/domain"
	backofficeHttp "github.com/hextec/yolyolakay-backend/internal/backoffice/delivery/http"
	backofficeAmqp "github.com/hextec/yolyolakay-backend/internal/backoffice/delivery/amqp"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/repository"
	"github.com/hextec/yolyolakay-backend/internal/backoffice/usecase"
	"github.com/hextec/yolyolakay-backend/pkg/cache"
	"github.com/hextec/yolyolakay-backend/pkg/graceful"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/hextec/yolyolakay-backend/pkg/tracing"
	"github.com/hextec/yolyolakay-backend/pkg/vault"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	_ "github.com/hextec/yolyolakay-backend/docs/backoffice_service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Backoffice API
// @version 1.0
// @description Admin panel and Remote Config service for Yolyolakay

// @host localhost:8092
// @BasePath /
func main() {
	profiler.Start("6060")

	shutdownTracer, err := tracing.InitTracer("backoffice_service")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	dbURL := os.Getenv("DB_A_URL")
	if dbURL == "" {
		dbURL = "postgresql://KernelStingray:KeRSt!n90$8!@group_jupiter:5432/kibertron?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Initialize Redis Cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisClient, err := cache.NewRedisClient(redisAddr, os.Getenv("REDIS_PASS"), 0)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize RabbitMQ
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://StingrayRab:KeRSt!n90$8!@rabbitmq:5672/"
	}
	rmqClient := rabbitmq.NewClient(rmqURL)
	defer rmqClient.Close()

	repo := repository.NewPostgresBackofficeRepository(pool)
	uc := usecase.NewBackofficeUseCase(repo, redisClient)

	// Seed default admin user KernelStingray
	ctxSeed := context.Background()
	if _, err := repo.GetAdminByLogin(ctxSeed, "KernelStingray"); err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("KeRSt!n90$8!"), bcrypt.DefaultCost)
		_ = repo.CreateAdmin(ctxSeed, &domain.Admin{
			ID:           uuid.New(),
			Login:        "KernelStingray",
			PasswordHash: string(hash),
			Role:         "admin",
		})
		log.Println("Seeded default admin user: KernelStingray")
	}

	// Start Audit Consumer
	auditConsumer := backofficeAmqp.NewAuditConsumer(rmqClient, uc)
	auditConsumer.Start(context.Background())

	r := gin.Default()
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.SentryMiddleware())
	r.Use(otelgin.Middleware("backoffice_service"))
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

	// Load Private and Public Keys for JWT validation and generation
	privateKey, err := vault.LoadPrivateKey()
	if err != nil {
		log.Fatalf("Failed to load private key from Vault: %v", err)
	}
	publicKey, err := vault.LoadPublicKey()
	if err != nil {
		log.Fatalf("Failed to load public key from Vault: %v", err)
	}
	authMiddleware := middleware.InternalAuth("admin", "moderator")
	auditMiddleware := middleware.AuditMiddleware(rmqClient)

	jwtManager := jwt.NewJWTManager(privateKey, publicKey)
	backofficeHttp.NewBackofficeHandler(r, uc, authMiddleware, auditMiddleware, jwtManager)

	log.Println("Starting Backoffice service on :8092")
	srv := &http.Server{
		Addr:    ":8092",
		Handler: r,
	}
	log.Printf("backoffice_service running on %s", ":8092")
	graceful.RunHTTPServer(srv, 5*time.Duration(1000000000))
}
