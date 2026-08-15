package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	limiter  *time.Time
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]int)
)

// RateLimiter is a simple IP-based rate limiter to prevent Brute Force attacks
func RateLimiter() gin.HandlerFunc {
	// Clean up memory every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			clients = make(map[string]int) // Reset counters periodically
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		clients[ip]++
		count := clients[ip]
		mu.Unlock()

		// Allow a maximum of 20 requests per minute per IP for sensitive routes
		if count > 20 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
