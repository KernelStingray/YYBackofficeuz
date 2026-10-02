package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	AuthPort         string
	Env              string
	DB_A_URL         string
	DB_B_URL         string
	RedisAddr        string
	RedisPass        string
	RabbitMQURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3BucketName     string
	InternalAPISecret string
	ClickMerchantID   string
	ClickServiceID    string
	ClickSecretKey    string
	ClickMerchantUserID string

	// Services URLs
	UserServiceURL         string
	AuthServiceURL         string
	ParcelServiceURL       string
	TelematicsServiceURL   string
	BillingServiceURL      string
	RideServiceURL         string
	NotificationServiceURL string
	RoutingServiceURL      string
	FraudServiceURL        string
	HelpdeskServiceURL     string
	AnalyticsServiceURL    string
	BackofficeServiceURL   string
	MediaServiceURL        string
	RatingServiceURL       string
}

func LoadConfig() *Config {
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Println("No configs/.env file found, using system environment variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		AuthPort:         getEnv("AUTH_PORT", "8081"),
		Env:              getEnv("ENV", "development"),
		DB_A_URL:         getEnv("DB_A_URL", ""),
		DB_B_URL:         getEnv("DB_B_URL", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:        getEnv("REDIS_PASS", ""),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "super-secret-access-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "super-secret-refresh-key"),
		S3Endpoint:       getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:      getEnv("S3_SECRET_KEY", "minioadmin"),
		S3BucketName:     getEnv("S3_BUCKET_NAME", "yolyolakay-media"),
		InternalAPISecret: getEnv("INTERNAL_API_SECRET", "super-internal-secret-for-microservices"),
		ClickMerchantID:   getEnv("CLICK_MERCHANT_ID", ""),
		ClickServiceID:    getEnv("CLICK_SERVICE_ID", ""),
		ClickSecretKey:    getEnv("CLICK_SECRET_KEY", ""),
		ClickMerchantUserID: getEnv("CLICK_MERCHANT_USER_ID", ""),

		// Services URLs with localhost fallback
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		ParcelServiceURL:       getEnv("PARCEL_SERVICE_URL", "http://localhost:8085"),
		TelematicsServiceURL:   getEnv("TELEMATICS_SERVICE_URL", "http://localhost:8087"),
		BillingServiceURL:      getEnv("BILLING_SERVICE_URL", "http://localhost:8084"),
		RideServiceURL:         getEnv("RIDE_SERVICE_URL", "http://localhost:8083"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8086"),
		RoutingServiceURL:      getEnv("ROUTING_SERVICE_URL", "http://localhost:8088"),
		FraudServiceURL:        getEnv("FRAUD_SERVICE_URL", "http://localhost:8089"),
		HelpdeskServiceURL:     getEnv("HELPDESK_SERVICE_URL", "http://localhost:8090"),
		AnalyticsServiceURL:    getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8091"),
		BackofficeServiceURL:   getEnv("BACKOFFICE_SERVICE_URL", "http://localhost:8092"),
		MediaServiceURL:        getEnv("MEDIA_SERVICE_URL", "http://localhost:8093"),
		RatingServiceURL:       getEnv("RATING_SERVICE_URL", "http://localhost:8094"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
