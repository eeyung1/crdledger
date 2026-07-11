package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a minimal fixed-window limiter keyed by client IP. It's
// intentionally simple — this app has ~200 users, not internet-scale
// traffic — but login/register endpoints need *some* brake against
// credential-stuffing and account-enumeration attempts.
type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	attempts map[string]*bucket
}

type bucket struct {
	count      int
	windowEnds time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:    limit,
		window:   window,
		attempts: make(map[string]*bucket),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, b := range rl.attempts {
			if now.After(b.windowEnds) {
				delete(rl.attempts, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.attempts[key]
	if !ok || now.After(b.windowEnds) {
		rl.attempts[key] = &bucket{count: 1, windowEnds: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.limit {
		return false
	}
	b.count++
	return true
}

// Limit only throttles unsafe methods (form submissions) so it never
// blocks someone just viewing the login page.
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if !rl.allow(clientIP(r)) {
				http.Error(w, "too many attempts — please wait a minute and try again", http.StatusTooManyRequests)
				return
			}
		}
		next(w, r)
	}
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
