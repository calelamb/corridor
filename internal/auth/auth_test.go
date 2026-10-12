package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRBACRolesAndContext(t *testing.T) {
	ctx := context.Background()
	// Unauthenticated defaults to RolePublic
	if r := RoleFromContext(ctx); r != RolePublic {
		t.Fatalf("expected RolePublic for empty context, got %s", r)
	}
	if UserFromContext(ctx) != nil {
		t.Fatalf("expected nil user for empty context")
	}

	user := &User{
		Subject: "ranger-123",
		Role:    RoleAnalyst,
		Email:   "ranger@nps.gov",
		Name:    "Ranger John",
		Agency:  "National Park Service",
	}
	ctxWithUser := WithUser(ctx, user)

	if r := RoleFromContext(ctxWithUser); r != RoleAnalyst {
		t.Fatalf("expected RoleAnalyst, got %s", r)
	}
	if u := UserFromContext(ctxWithUser); u == nil || u.Subject != "ranger-123" {
		t.Fatalf("expected user 'ranger-123', got %+v", u)
	}

	if !HasRole(ctxWithUser, RoleAnalyst, RoleAdmin) {
		t.Fatalf("expected HasRole to match RoleAnalyst")
	}
	if HasRole(ctxWithUser, RoleReporter, RolePublic) {
		t.Fatalf("expected HasRole not to match Reporter or Public")
	}
}

func TestPrivacyFilters(t *testing.T) {
	// Public
	pubFilter := SensitiveFilterForRole(RolePublic)
	if pubFilter.AllowRawCoordinates {
		t.Fatalf("public role must not allow raw coordinates")
	}
	if pubFilter.MaxH3Resolution != 6 {
		t.Fatalf("public role must cap H3 resolution at 6, got %d", pubFilter.MaxH3Resolution)
	}
	if pubFilter.MinAgeDelay != 30*24*time.Hour {
		t.Fatalf("public role must impose 30-day delay, got %v", pubFilter.MinAgeDelay)
	}

	// Reporter
	repFilter := SensitiveFilterForRole(RoleReporter)
	if repFilter.AllowRawCoordinates {
		t.Fatalf("reporter role must not allow raw coordinates")
	}

	// Analyst (Ranger / Biologist)
	analystFilter := SensitiveFilterForRole(RoleAnalyst)
	if !analystFilter.AllowRawCoordinates {
		t.Fatalf("analyst role must allow raw coordinates")
	}
	if analystFilter.MaxH3Resolution != 15 {
		t.Fatalf("analyst role should have resolution 15, got %d", analystFilter.MaxH3Resolution)
	}
	if analystFilter.MinAgeDelay != 0 {
		t.Fatalf("analyst role must not have delay, got %v", analystFilter.MinAgeDelay)
	}

	// Admin
	adminFilter := SensitiveFilterForRole(RoleAdmin)
	if !adminFilter.AllowRawCoordinates || adminFilter.MinAgeDelay != 0 {
		t.Fatalf("admin role must have un-redacted access")
	}
}

func TestOIDCAuthentication(t *testing.T) {
	secret := "super-secure-hmac-secret-test-key-32b"
	cfg := Config{
		Issuer:   "https://auth.nps.gov",
		Audience: "corridor-api",
		Secret:   secret,
		APIKeys: map[string]User{
			"key-dispatch-1": {
				Subject: "dispatcher-1",
				Role:    RoleReporter,
				Agency:  "WYDOT",
			},
			"key-ranger-2": {
				Subject: "ranger-2",
				Role:    RoleAnalyst,
				Agency:  "NPS-Yellowstone",
			},
		},
	}
	auth := NewAuthenticator(cfg)

	t.Run("Valid Analyst JWT", func(t *testing.T) {
		claims := Claims{
			Issuer:    "https://auth.nps.gov",
			Subject:   "ranger-007",
			Audience:  "corridor-api",
			Role:      RoleAnalyst,
			Email:     "ranger.james@nps.gov",
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}
		tok, err := SignToken(claims, secret)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		user, err := auth.AuthenticateToken(context.Background(), tok)
		if err != nil {
			t.Fatalf("expected token validation to succeed, got %v", err)
		}
		if user.Subject != "ranger-007" || user.Role != RoleAnalyst {
			t.Fatalf("unexpected user: %+v", user)
		}
	})

	t.Run("Expired JWT", func(t *testing.T) {
		claims := Claims{
			Issuer:    "https://auth.nps.gov",
			Subject:   "ranger-old",
			Audience:  "corridor-api",
			Role:      RoleAnalyst,
			ExpiresAt: time.Now().Add(-10 * time.Minute).Unix(),
		}
		tok, err := SignToken(claims, secret)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = auth.AuthenticateToken(context.Background(), tok)
		if err == nil || err != ErrTokenExpired {
			t.Fatalf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("Tampered Signature", func(t *testing.T) {
		claims := Claims{
			Issuer:    "https://auth.nps.gov",
			Subject:   "ranger-bad",
			Audience:  "corridor-api",
			Role:      RoleAdmin,
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}
		tok, err := SignToken(claims, "different-secret-key")
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = auth.AuthenticateToken(context.Background(), tok)
		if err == nil || err != ErrInvalidSignature {
			t.Fatalf("expected ErrInvalidSignature, got %v", err)
		}
	})

	t.Run("API Key Validation", func(t *testing.T) {
		user, err := auth.AuthenticateAPIKey(context.Background(), "key-ranger-2")
		if err != nil {
			t.Fatalf("expected valid API key, got %v", err)
		}
		if user.Role != RoleAnalyst || user.Subject != "ranger-2" {
			t.Fatalf("unexpected user: %+v", user)
		}

		_, err = auth.AuthenticateAPIKey(context.Background(), "invalid-key")
		if err != ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("AuthenticateRequest Headers", func(t *testing.T) {
		// Bearer header
		claims := Claims{
			Issuer:    "https://auth.nps.gov",
			Subject:   "ranger-bearer",
			Audience:  "corridor-api",
			Role:      RoleAnalyst,
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}
		tok, _ := SignToken(claims, secret)
		req := httptest.NewRequest(http.MethodGet, "/v1/reports", nil)
		req.Header.Set("Authorization", "Bearer "+tok)

		user, err := auth.AuthenticateRequest(req)
		if err != nil || user == nil || user.Subject != "ranger-bearer" {
			t.Fatalf("expected valid user from Authorization header, got %+v (err: %v)", user, err)
		}

		// API Key header
		reqKey := httptest.NewRequest(http.MethodGet, "/v1/reports", nil)
		reqKey.Header.Set("X-API-Key", "key-dispatch-1")
		userKey, err := auth.AuthenticateRequest(reqKey)
		if err != nil || userKey == nil || userKey.Subject != "dispatcher-1" {
			t.Fatalf("expected valid user from X-API-Key header, got %+v", userKey)
		}

		// Empty header
		reqEmpty := httptest.NewRequest(http.MethodGet, "/v1/reports", nil)
		userEmpty, err := auth.AuthenticateRequest(reqEmpty)
		if err != nil || userEmpty != nil {
			t.Fatalf("expected nil user for unauthenticated request, got %+v (err: %v)", userEmpty, err)
		}
	})
}
