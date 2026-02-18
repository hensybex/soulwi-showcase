package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiterMiddleware struct {
	// Map of firebase_uid to individual rate limiters
	limiters map[string]*rate.Limiter
	mu       sync.Mutex // Protects the map
	rate     rate.Limit // Tokens per second
	burst    int        // Maximum burst capacity
}

// NewRateLimiterMiddleware creates a new instance
func NewRateLimiterMiddleware(r rate.Limit, b int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// CheckRateLimit is the Gin middleware handler
func (m *RateLimiterMiddleware) CheckRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get firebase_uid from context
		firebaseUIDVal, exists := c.Get("firebase_uid")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
			return
		}

		firebaseUID, ok := firebaseUIDVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
			return
		}

		// Get or create rate limiter for this user
		limiter := m.getLimiter(firebaseUID)

		// Check if request is allowed
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "RPM limit exceeded",
				"message":     "Too many requests, please try again later",
				"rate_limit":  m.rate * 60, // RPM for display
				"rate_period": "minute",
			})
			return
		}

		c.Next()
	}
}

func (m *RateLimiterMiddleware) getLimiter(uid string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[uid]
	if !exists {
		limiter = rate.NewLimiter(m.rate, m.burst)
		m.limiters[uid] = limiter
	}
	return limiter
}
