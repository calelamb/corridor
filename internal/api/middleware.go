package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
