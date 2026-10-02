package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
)

// RateLimit implements a fixed-window rate limiter using Redis with a default "global" prefix.
func RateLimit(rdb *redis.Client, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return RateLimitWithPrefix(rdb, "global", maxRequests, window)
}

// RateLimitWithPrefix implements a fixed-window rate limiter using Redis with a specific key prefix.
func RateLimitWithPrefix(rdb *redis.Client, prefix string, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get real client IP (prioritizing Cloudflare and proxy headers)
		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
			if ip != "" {
				parts := strings.Split(ip, ",")
				ip = strings.TrimSpace(parts[0])
			}
		}
		if ip == "" {
			ip = c.ClientIP()
		}
		if ip == "" {
			ip = "unknown"
		}

		if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
			c.Next()
			return
		}

		// Fixed window algorithm using the current minute/window as part of the key
		windowKey := time.Now().Unix() / int64(window.Seconds())
		key := fmt.Sprintf("ratelimit:%s:ip:%s:%d", prefix, ip, windowKey)

		// Increment the counter
		ctx := c.Request.Context()
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			fmt.Printf("Rate limit redis error: %v\n", err)
			c.Next()
			return
		}

		// Set expiry on the very first increment
		if count == 1 {
			rdb.Expire(ctx, key, window+10*time.Second) // add a small buffer
		}

		if count > maxRequests {
			c.Header("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests, please try again later",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

