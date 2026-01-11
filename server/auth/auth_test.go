package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirsjg/momentum/ratelimit"
)

func newTestHandler() *Handler {
	// Use a high burst for most tests so rate limiting doesn't interfere
	cfg := ratelimit.Config{
		Rate:     100,
		Interval: time.Second,
		Burst:    100,
	}
	return NewHandlerWithConfig(cfg)
}

func TestHandler_Login_Success(t *testing.T) {
	h := newTestHandler()

	body := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Error("Expected token in response")
	}
	if resp.UserID == "" {
		t.Error("Expected user_id in response")
	}
}

func TestHandler_Login_MissingFields(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name string
		body LoginRequest
	}{
		{"missing email", LoginRequest{Password: "password123"}},
		{"missing password", LoginRequest{Email: "test@example.com"}},
		{"both missing", LoginRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandler_Login_InvalidMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

func TestHandler_Register_Success(t *testing.T) {
	h := newTestHandler()

	body := RegisterRequest{
		Email:    "newuser@example.com",
		Password: "securepassword123",
		Name:     "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Message == "" {
		t.Error("Expected message in response")
	}
}

func TestHandler_Register_ShortPassword(t *testing.T) {
	h := newTestHandler()

	body := RegisterRequest{
		Email:    "newuser@example.com",
		Password: "short",
		Name:     "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message in response")
	}
}

func TestHandler_Register_MissingFields(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name string
		body RegisterRequest
	}{
		{"missing email", RegisterRequest{Password: "password123"}},
		{"missing password", RegisterRequest{Email: "test@example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandler_ResetPassword_Success(t *testing.T) {
	h := newTestHandler()

	body := ResetPasswordRequest{
		Email: "user@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Message == "" {
		t.Error("Expected message in response")
	}
}

func TestHandler_ResetPassword_MissingEmail(t *testing.T) {
	h := newTestHandler()

	body := ResetPasswordRequest{}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandler_RateLimiting(t *testing.T) {
	// Create handler with strict rate limiting
	cfg := ratelimit.Config{
		Rate:     1,
		Interval: time.Minute,
		Burst:    2,
	}
	h := NewHandlerWithConfig(cfg)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	// First two requests should succeed (burst = 2)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
		req.RemoteAddr = "192.168.1.100:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Request 3: expected 429, got %d", rec.Code)
	}
}

func TestHandler_RateLimiting_AllEndpoints(t *testing.T) {
	// Each auth endpoint should be rate limited
	cfg := ratelimit.Config{
		Rate:     1,
		Interval: time.Minute,
		Burst:    1,
	}
	h := NewHandlerWithConfig(cfg)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	endpoints := []struct {
		path string
		body interface{}
	}{
		{"/auth/login", LoginRequest{Email: "test@example.com", Password: "password123"}},
		{"/auth/register", RegisterRequest{Email: "test@example.com", Password: "password123456"}},
		{"/auth/reset-password", ResetPasswordRequest{Email: "test@example.com"}},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			// Reset rate limiter for each test
			h.Limiter().ResetAll()

			jsonBody, _ := json.Marshal(ep.body)

			// First request should succeed
			req := httptest.NewRequest(http.MethodPost, ep.path, bytes.NewReader(jsonBody))
			req.RemoteAddr = "192.168.1.50:12345"
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code >= 400 && rec.Code != http.StatusCreated {
				// Allow 2xx status codes
				if rec.Code == http.StatusTooManyRequests {
					t.Errorf("%s: first request should not be rate limited", ep.path)
				}
			}

			// Second request should be rate limited
			req = httptest.NewRequest(http.MethodPost, ep.path, bytes.NewReader(jsonBody))
			req.RemoteAddr = "192.168.1.50:12345"
			req.Header.Set("Content-Type", "application/json")
			rec = httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusTooManyRequests {
				t.Errorf("%s: second request should be rate limited, got %d", ep.path, rec.Code)
			}
		})
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"login", h.Login},
		{"register", h.Register},
		{"reset-password", h.ResetPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/"+tt.name, bytes.NewReader([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", rec.Code)
			}
		})
	}
}
