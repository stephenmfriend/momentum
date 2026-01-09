package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	cfg := Config{
		Rate:     2,
		Interval: time.Second,
		Burst:    3,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	ip := "192.168.1.1"

	// First 3 requests should be allowed (burst)
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("Request %d should be allowed (within burst)", i+1)
		}
	}

	// 4th request should be denied
	if limiter.Allow(ip) {
		t.Error("Request 4 should be denied (exceeded burst)")
	}
}

func TestLimiter_TokenRefill(t *testing.T) {
	cfg := Config{
		Rate:     2,
		Interval: 50 * time.Millisecond,
		Burst:    2,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	ip := "192.168.1.2"

	// Exhaust the bucket
	limiter.Allow(ip)
	limiter.Allow(ip)

	// Should be denied
	if limiter.Allow(ip) {
		t.Error("Should be denied after exhausting bucket")
	}

	// Wait for token refill
	time.Sleep(60 * time.Millisecond)

	// Should be allowed after refill
	if !limiter.Allow(ip) {
		t.Error("Should be allowed after token refill")
	}
}

func TestLimiter_DifferentIPs(t *testing.T) {
	cfg := Config{
		Rate:     1,
		Interval: time.Second,
		Burst:    1,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Each IP should get its own bucket
	if !limiter.Allow(ip1) {
		t.Error("IP1 first request should be allowed")
	}
	if !limiter.Allow(ip2) {
		t.Error("IP2 first request should be allowed")
	}

	// Both should be denied now
	if limiter.Allow(ip1) {
		t.Error("IP1 second request should be denied")
	}
	if limiter.Allow(ip2) {
		t.Error("IP2 second request should be denied")
	}
}

func TestLimiter_Middleware(t *testing.T) {
	cfg := Config{
		Rate:     1,
		Interval: time.Minute,
		Burst:    2,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	handlerCalled := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	})

	wrapped := limiter.Middleware(handler)

	// First two requests should pass (burst = 2)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Request 3: expected 429, got %d", rec.Code)
	}

	if handlerCalled != 2 {
		t.Errorf("Handler should be called 2 times, got %d", handlerCalled)
	}

	// Check Retry-After header is set
	if rec.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header should be set")
	}
}

func TestLimiter_MiddlewareFunc(t *testing.T) {
	cfg := Config{
		Rate:     1,
		Interval: time.Minute,
		Burst:    1,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	handlerCalled := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}

	wrapped := limiter.MiddlewareFunc(handler)

	// First request should pass
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	rec := httptest.NewRecorder()
	wrapped(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("First request: expected 200, got %d", rec.Code)
	}

	// Second request should be rate limited
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	rec = httptest.NewRecorder()
	wrapped(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected 429, got %d", rec.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expected   string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "198.51.100.178"},
			expected:   "198.51.100.178",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"X-Real-IP":       "198.51.100.178",
			},
			expected: "203.0.113.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, ip)
			}
		})
	}
}

func TestLimiter_Reset(t *testing.T) {
	cfg := Config{
		Rate:     1,
		Interval: time.Minute,
		Burst:    1,
	}
	limiter := NewLimiter(cfg)
	defer limiter.ResetAll()

	ip := "192.168.1.50"

	// Exhaust the bucket
	limiter.Allow(ip)
	if limiter.Allow(ip) {
		t.Error("Should be denied after exhausting bucket")
	}

	// Reset the IP
	limiter.Reset(ip)

	// Should be allowed again
	if !limiter.Allow(ip) {
		t.Error("Should be allowed after reset")
	}
}

func TestDefaultAuthConfig(t *testing.T) {
	cfg := DefaultAuthConfig()

	if cfg.Rate != 5 {
		t.Errorf("Expected rate 5, got %d", cfg.Rate)
	}
	if cfg.Interval != time.Minute {
		t.Errorf("Expected interval 1m, got %v", cfg.Interval)
	}
	if cfg.Burst != 10 {
		t.Errorf("Expected burst 10, got %d", cfg.Burst)
	}
}
