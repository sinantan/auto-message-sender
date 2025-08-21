package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

type ClientInfo struct {
	requests []time.Time
	mu       sync.Mutex
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   limit,
		window:  window,
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.RLock()
	client, exists := rl.clients[clientIP]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		client = &ClientInfo{
			requests: make([]time.Time, 0),
		}
		rl.clients[clientIP] = client
		rl.mu.Unlock()
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	now := time.Now()

	var validRequests []time.Time
	for _, req := range client.requests {
		if now.Sub(req) <= rl.window {
			validRequests = append(validRequests, req)
		}
	}
	client.requests = validRequests

	if len(client.requests) >= rl.limit {
		return false
	}

	client.requests = append(client.requests, now)
	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientIP, client := range rl.clients {
		client.mu.Lock()
		hasRecentRequests := false
		for _, req := range client.requests {
			if now.Sub(req) <= rl.window {
				hasRecentRequests = true
				break
			}
		}
		client.mu.Unlock()

		if !hasRecentRequests {
			delete(rl.clients, clientIP)
		}
	}
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
