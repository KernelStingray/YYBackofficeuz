package middleware

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SentryMiddleware is a gin middleware that integrates with Sentry.
// It will automatically recover from panics, log them to Sentry, and return a 500 error.
func SentryMiddleware() gin.HandlerFunc {
	// Initialize the Sentry middleware for Gin
	// Repanic: false means it will return a 500 status code automatically and not crash the server.
	return sentrygin.New(sentrygin.Options{
		Repanic:         false,
		WaitForDelivery: false,
	})
}
