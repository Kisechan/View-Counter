package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu          sync.Mutex
	ips         map[string]*rateInfo
	maxRequests int
	window      time.Duration
}

type rateInfo struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		ips:         make(map[string]*rateInfo),
		maxRequests: maxRequests,
		window:      window,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for range time.Tick(time.Minute) {
		rl.mu.Lock()
		for ip, info := range rl.ips {
			if time.Since(info.lastSeen) > rl.window*2 {
				delete(rl.ips, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		info, exists := rl.ips[ip]
		if !exists {
			info = &rateInfo{}
			rl.ips[ip] = info
		}

		if time.Since(info.lastSeen) > rl.window {
			info.count = 0
		}

		info.count++
		info.lastSeen = time.Now()

		if info.count > rl.maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
