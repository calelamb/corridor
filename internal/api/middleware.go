package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"corridor/internal/auth"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/tiles/")
}
func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{Status: Error, Error: message})
}
func protect(logger *slog.Logger, limiter *rate.Limiter, timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.NewString()
			w.Header().Set("X-Request-ID", id)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-store")
			defer func() {
				if recover() != nil {
					logger.ErrorContext(r.Context(), "request panic", "request_id", id)
					writeError(w, 500, "Request failed")
				}
			}()
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				writeError(w, 429, "Too many requests")
				return
			}
			if r.ContentLength > 1<<20 {
				writeError(w, 413, "Request too large")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Authenticate extracts and verifies user identity from Authorization or X-API-Key headers.
func Authenticate(authenticator auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authenticator == nil {
				next.ServeHTTP(w, r)
				return
			}
			user, err := authenticator.AuthenticateRequest(r)
			if err != nil {
				if errors.Is(err, auth.ErrTokenExpired) {
					writeError(w, http.StatusUnauthorized, "Token expired")
					return
				}
				if errors.Is(err, auth.ErrUnauthorized) || errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrInvalidSignature) {
					writeError(w, http.StatusUnauthorized, "Unauthorized")
					return
				}
			}
			if user != nil {
				ctx := auth.WithUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole enforces that the authenticated user possesses one of the allowed roles.
func RequireRole(roles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth.HasRole(r.Context(), roles...) {
				next.ServeHTTP(w, r)
				return
			}
			user := auth.UserFromContext(r.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "Authentication required")
				return
			}
			writeError(w, http.StatusForbidden, "Forbidden: insufficient permissions")
		})
	}
}
