// Package ratelimit provides rate limiting middleware for HTTP handlers.
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter per client IP.
type Limiter struct {
	mu       sync.Mutex
	clients  map[string]*bucket
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	burst    int           // max tokens (bucket capacity)
	cleanup  time.Duration // cleanup interval for stale entries
}

type bucket struct {
	tokens    int
	lastCheck time.Time
}

// Config holds rate limiter configuration.
type Config struct {
	Rate     int           // requests allowed per interval
	Interval time.Duration // time interval for rate
	Burst    int           // maximum burst size
}

// DefaultAuthConfig returns sensible defaults for auth endpoints.
// Allows 5 requests per minute with a burst of 10.
func DefaultAuthConfig() Config {
	return Config{
		Rate:     5,
		Interval: time.Minute,
		Burst:    10,
	}
}

// NewLimiter creates a new rate limiter with the given configuration.
func NewLimiter(cfg Config) *Limiter {
	l := &Limiter{
		clients:  make(map[string]*bucket),
		rate:     cfg.Rate,
		interval: cfg.Interval,
		burst:    cfg.Burst,
		cleanup:  5 * time.Minute,
	}

	// Start background cleanup goroutine
	go l.cleanupLoop()

	return l
}

// cleanupLoop removes stale entries periodically.
func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, b := range l.clients {
			// Remove entries that haven't been accessed in 10 minutes
			if now.Sub(b.lastCheck) > 10*time.Minute {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed.
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, exists := l.clients[ip]
	if !exists {
		// New client starts with full bucket
		l.clients[ip] = &bucket{
			tokens:    l.burst - 1, // consume one token for this request
			lastCheck: now,
		}
		return true
	}

	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(b.lastCheck)
	tokensToAdd := int(elapsed / l.interval) * l.rate

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		b.lastCheck = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// Middleware returns an HTTP middleware that applies rate limiting.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !l.Allow(ip) {
			w.Header().Set("Retry-After", l.interval.String())
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc returns an HTTP middleware function for use with HandlerFunc.
func (l *Limiter) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !l.Allow(ip) {
			w.Header().Set("Retry-After", l.interval.String())
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// getClientIP extracts the client IP from the request.
// It checks X-Forwarded-For and X-Real-IP headers for proxied requests.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// Strip port if present
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// Reset clears rate limit state for a specific IP (useful for testing).
func (l *Limiter) Reset(ip string) {
	l.mu.Lock()
	delete(l.clients, ip)
	l.mu.Unlock()
}

// ResetAll clears all rate limit state (useful for testing).
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	l.clients = make(map[string]*bucket)
	l.mu.Unlock()
}
