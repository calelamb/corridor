package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"corridor/internal/prioritize"

	"github.com/google/uuid"
)

type mockMitigationProvider struct {
	candidates []prioritize.CandidateSite
	err        error
	lastBudget float64
	lastMinBCR *float64
	lastState  *string
}

func (m *mockMitigationProvider) GenerateCandidates(ctx context.Context, budget float64, minBCR *float64, state *string) ([]prioritize.CandidateSite, error) {
	m.lastBudget = budget
	m.lastMinBCR = minBCR
	m.lastState = state
	if m.err != nil {
		return nil, m.err
	}
	return m.candidates, nil
}

func TestMitigationEndpoint_Success(t *testing.T) {
	siteID := uuid.New()
	segID1 := uuid.New()
	segID2 := uuid.New()

	provider := &mockMitigationProvider{
		candidates: []prioritize.CandidateSite{
			{
				ID:                        siteID,
				Segments:                  []uuid.UUID{segID1, segID2},
				State:                     "WY",
				HighwayRef:                "US-191",
				MilepostStart:             10.0,
				MilepostEnd:               10.2,
				LengthM:                   321.8,
				RecommendedType:           "overpass_with_fencing",
				EstCostUSD:                4977500.0,
				ExpectedAvoidedCollisions: 215.0,
				ExpectedSavingsUSD:        9675000.0,
				BCR:                       1.94,
				P10BCR:                    1.16,
				P90BCR:                    2.91,
				Rationale:                 "Recommended Wildlife Overpass with Continuous Fencing for US-191.",
			},
		},
	}

	h := NewRouter(Dependencies{
		MitigationProvider: provider,
		Ready:              func(context.Context) error { return nil },
	})

	// Test GET /v1/mitigation with budget and min_bcr
	req := httptest.NewRequest("GET", "/v1/mitigation?budget=5000000&min_bcr=1.0&state=WY", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var env MitigationEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal response envelope: %v", err)
	}

	if env.Status != MitigationEnvelopeStatusSuccess {
		t.Fatalf("expected status success, got %s", env.Status)
	}

	if len(env.Data) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(env.Data))
	}

	item := env.Data[0]
	if item.Id != siteID {
		t.Fatalf("expected site id %s, got %s", siteID, item.Id)
	}
	if item.RecommendedType != "overpass_with_fencing" {
		t.Fatalf("expected overpass_with_fencing, got %s", item.RecommendedType)
	}
	if item.Bcr != 1.94 {
		t.Fatalf("expected BCR 1.94, got %.2f", item.Bcr)
	}
	if len(item.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(item.Segments))
	}

	if provider.lastBudget != 5000000.0 {
		t.Fatalf("expected budget 5000000, got %.2f", provider.lastBudget)
	}
	if provider.lastMinBCR == nil || *provider.lastMinBCR != 1.0 {
		t.Fatalf("expected min_bcr 1.0, got %v", provider.lastMinBCR)
	}
	if provider.lastState == nil || *provider.lastState != "WY" {
		t.Fatalf("expected state WY, got %v", provider.lastState)
	}

	// Test lowercase state normalization
	reqLower := httptest.NewRequest("GET", "/v1/mitigation?state=wy", nil)
	wLower := httptest.NewRecorder()
	h.ServeHTTP(wLower, reqLower)
	if wLower.Code != http.StatusOK {
		t.Fatalf("expected status 200 for state=wy, got %d", wLower.Code)
	}
	if provider.lastState == nil || *provider.lastState != "WY" {
		t.Fatalf("expected state normalized to WY, got %v", provider.lastState)
	}
}

func TestMitigationEndpoint_Validation(t *testing.T) {
	provider := &mockMitigationProvider{}
	h := NewRouter(Dependencies{
		MitigationProvider: provider,
		Ready:              func(context.Context) error { return nil },
	})

	tests := []struct {
		name       string
		path       string
		wantCode   int
		errSnippet string
	}{
		{
			name:       "Negative budget",
			path:       "/v1/mitigation?budget=-100",
			wantCode:   http.StatusBadRequest,
			errSnippet: "budget",
		},
		{
			name:       "Negative min_bcr",
			path:       "/v1/mitigation?min_bcr=-0.5",
			wantCode:   http.StatusBadRequest,
			errSnippet: "min_bcr",
		},
		{
			name:       "State too long",
			path:       "/v1/mitigation?state=WYOMING",
			wantCode:   http.StatusBadRequest,
			errSnippet: "state",
		},
		{
			name:       "State too short",
			path:       "/v1/mitigation?state=W",
			wantCode:   http.StatusBadRequest,
			errSnippet: "state",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}

			if !strings.Contains(w.Body.String(), tc.errSnippet) {
				t.Fatalf("expected error body to contain %q, got: %s", tc.errSnippet, w.Body.String())
			}
		})
	}
}

func TestMitigationEndpoint_Unavailable(t *testing.T) {
	// Case 1: Nil provider
	hNil := NewRouter(Dependencies{
		Ready: func(context.Context) error { return nil },
	})
	req := httptest.NewRequest("GET", "/v1/mitigation", nil)
	w := httptest.NewRecorder()
	hNil.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for nil provider, got %d: %s", w.Code, w.Body.String())
	}

	// Case 2: Provider returns error
	providerErr := &mockMitigationProvider{err: errors.New("database connection failed")}
	hErr := NewRouter(Dependencies{
		MitigationProvider: providerErr,
		Ready:              func(context.Context) error { return nil },
	})
	req2 := httptest.NewRequest("GET", "/v1/mitigation", nil)
	w2 := httptest.NewRecorder()
	hErr.ServeHTTP(w2, req2)

	if w2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for provider error, got %d: %s", w2.Code, w2.Body.String())
	}
	// Verify internal error is not leaked
	if strings.Contains(w2.Body.String(), "database connection failed") {
		t.Fatalf("internal error leaked in response: %s", w2.Body.String())
	}
}
