package middleware

import (
	"net/http"
	"sync"
	"time"
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
	
	// 定期清理旧的记录
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

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		
		rl.mu.Lock()
		defer rl.mu.Unlock()
		
		info, exists := rl.ips[ip]
		if !exists {
			info = &rateInfo{}
			rl.ips[ip] = info
		}
		
		// 如果超过时间窗口，重置
		if time.Since(info.lastSeen) > rl.window {
			info.count = 0
		}
		
		info.count++
		info.lastSeen = time.Now()
		
		if info.count > rl.maxRequests {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}