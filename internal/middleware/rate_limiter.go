// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiter() gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, exists := limiters[ip]
		if !exists {
			l = rate.NewLimiter(1, 5) // 1 request/detik, burst maksimal 5
			limiters[ip] = l
		}
		return l
	}

	return func(c *gin.Context) {
		limiter := getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, slow down"})
			c.Abort()
			return
		}
		c.Next()
	}
}