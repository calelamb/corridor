package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"corridor/internal/auth"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key-32-bytes-auth-jwt"
	authenticator := auth.NewAuthenticator(auth.Config{
		Issuer:   "https://auth.nps.gov",
		Audience: "corridor-api",
		Secret:   secret,
		APIKeys: map[string]auth.User{
			"valid-analyst-key": {
				Subject: "ranger-99",
				Role:    auth.RoleAnalyst,
				Agency:  "NPS",
			},
			"valid-reporter-key": {
				Subject: "citizen-1",
				Role:    auth.RoleReporter,
			},
		},
	})

	handler := Authenticate(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := auth.RoleFromContext(r.Context())
		user := auth.UserFromContext(r.Context())
		if user != nil {
			w.Header().Set("X-User-Subject", user.Subject)
		}
		w.Header().Set("X-User-Role", string(role))
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Unauthenticated request defaults to RolePublic", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if role := rec.Header().Get("X-User-Role"); role != string(auth.RolePublic) {
			t.Fatalf("expected public role, got %s", role)
		}
	})

	t.Run("Authenticated with valid API Key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		req.Header.Set("X-API-Key", "valid-analyst-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if role := rec.Header().Get("X-User-Role"); role != string(auth.RoleAnalyst) {
			t.Fatalf("expected analyst role, got %s", role)
		}
		if sub := rec.Header().Get("X-User-Subject"); sub != "ranger-99" {
			t.Fatalf("expected subject ranger-99, got %s", sub)
		}
	})

	t.Run("Authenticated with valid JWT Bearer", func(t *testing.T) {
		tok, err := auth.SignToken(auth.Claims{
			Issuer:    "https://auth.nps.gov",
			Subject:   "ranger-jwt-1",
			Audience:  "corridor-api",
			Role:      auth.RoleAnalyst,
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}, secret)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if role := rec.Header().Get("X-User-Role"); role != string(auth.RoleAnalyst) {
			t.Fatalf("expected analyst role, got %s", role)
		}
	})

	t.Run("RequireRole denies unauthenticated with 401", func(t *testing.T) {
		protectedHandler := Authenticate(authenticator)(
			RequireRole(auth.RoleAnalyst, auth.RoleAdmin)(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		)

		req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
		rec := httptest.NewRecorder()
		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("RequireRole denies insufficient role with 403", func(t *testing.T) {
		protectedHandler := Authenticate(authenticator)(
			RequireRole(auth.RoleAnalyst, auth.RoleAdmin)(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		)

		req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
		req.Header.Set("X-API-Key", "valid-reporter-key") // Reporter, not Analyst
		rec := httptest.NewRecorder()
		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
		}
	})

	t.Run("RequireRole allows analyst role with 200", func(t *testing.T) {
		protectedHandler := Authenticate(authenticator)(
			RequireRole(auth.RoleAnalyst, auth.RoleAdmin)(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		)

		req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
		req.Header.Set("X-API-Key", "valid-analyst-key")
		rec := httptest.NewRecorder()
		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
	})
}
