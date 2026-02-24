// Package auth provides authentication HTTP handlers with rate limiting.
package auth

import (
	"encoding/json"
	"net/http"

	"github.com/stephenmfriend/momentum/ratelimit"
)

// Handler provides HTTP handlers for authentication endpoints.
type Handler struct {
	limiter *ratelimit.Limiter
}

// NewHandler creates a new auth handler with rate limiting.
func NewHandler() *Handler {
	return &Handler{
		limiter: ratelimit.NewLimiter(ratelimit.DefaultAuthConfig()),
	}
}

// NewHandlerWithConfig creates a new auth handler with custom rate limit config.
func NewHandlerWithConfig(cfg ratelimit.Config) *Handler {
	return &Handler{
		limiter: ratelimit.NewLimiter(cfg),
	}
}

// LoginRequest represents a login request payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request payload.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// ResetPasswordRequest represents a password reset request payload.
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// AuthResponse represents a successful auth response.
type AuthResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Login handles POST /auth/login with rate limiting.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual authentication logic here
	// This is a placeholder that should be replaced with real auth
	writeJSON(w, AuthResponse{
		Token:   "placeholder_token",
		Message: "Login successful",
		UserID:  "user_123",
	}, http.StatusOK)
}

// Register handles POST /auth/register with rate limiting.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		writeError(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual registration logic here
	// This is a placeholder that should be replaced with real registration
	writeJSON(w, AuthResponse{
		Message: "Registration successful",
		UserID:  "user_123",
	}, http.StatusCreated)
}

// ResetPassword handles POST /auth/reset-password with rate limiting.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		writeError(w, "Email is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual password reset logic here
	// Always return success to prevent email enumeration attacks
	writeJSON(w, AuthResponse{
		Message: "If the email exists, a password reset link has been sent",
	}, http.StatusOK)
}

// RegisterRoutes registers auth endpoints with rate limiting on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/login", h.limiter.MiddlewareFunc(h.Login))
	mux.HandleFunc("/auth/register", h.limiter.MiddlewareFunc(h.Register))
	mux.HandleFunc("/auth/reset-password", h.limiter.MiddlewareFunc(h.ResetPassword))
}

// Limiter returns the rate limiter (useful for testing).
func (h *Handler) Limiter() *ratelimit.Limiter {
	return h.limiter
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, ErrorResponse{Error: message}, status)
}
