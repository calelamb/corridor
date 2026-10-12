package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"corridor/internal/prioritize"

	"github.com/google/uuid"
)

func TestMitigationBriefPDF_Success(t *testing.T) {
	siteID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	segID := uuid.New()

	provider := &mockMitigationProvider{
		candidates: []prioritize.CandidateSite{
			{
				ID:                        siteID,
				Segments:                  []uuid.UUID{segID},
				State:                     "WY",
				HighwayRef:                "US-191",
				MilepostStart:             12.0,
				MilepostEnd:               12.8,
				LengthM:                   1287.4,
				RecommendedType:           "overpass_with_fencing",
				EstCostUSD:                5200000.0,
				ExpectedAvoidedCollisions: 230.0,
				ExpectedSavingsUSD:        10350000.0,
				BCR:                       1.99,
				P10BCR:                    1.20,
				P90BCR:                    2.98,
				Rationale:                 "Critical migration corridor connecting high-density seasonal ungulate range.",
			},
		},
	}

	h := NewRouter(Dependencies{
		MitigationProvider: provider,
		Ready:              func(context.Context) error { return nil },
	})

	path := fmt.Sprintf("/v1/mitigation/%s/brief.pdf", siteID.String())
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/pdf") {
		t.Errorf("expected Content-Type application/pdf, got %q", contentType)
	}

	contentDisp := w.Header().Get("Content-Disposition")
	expectedDisp := fmt.Sprintf(`inline; filename="funding-brief-%s.pdf"`, siteID.String())
	if contentDisp != expectedDisp {
		t.Errorf("expected Content-Disposition %q, got %q", expectedDisp, contentDisp)
	}

	body := w.Body.Bytes()
	if !bytes.HasPrefix(body, []byte("%PDF-")) {
		t.Errorf("expected PDF magic header %%PDF-")
	}
	if !bytes.Contains(body, []byte("%%EOF")) {
		t.Errorf("expected PDF %%EOF marker")
	}
	if len(body) < 1024 {
		t.Errorf("expected PDF body > 1024 bytes, got %d", len(body))
	}
}

func TestMitigationBriefPDF_InvalidUUID(t *testing.T) {
	provider := &mockMitigationProvider{}
	h := NewRouter(Dependencies{
		MitigationProvider: provider,
		Ready:              func(context.Context) error { return nil },
	})

	// Case 1: Malformed UUID string
	req := httptest.NewRequest("GET", "/v1/mitigation/not-a-valid-uuid/brief.pdf", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for malformed UUID, got %d: %s", w.Code, w.Body.String())
	}

	// Case 2: Nil UUID (00000000-0000-0000-0000-000000000000)
	reqNil := httptest.NewRequest("GET", fmt.Sprintf("/v1/mitigation/%s/brief.pdf", uuid.Nil.String()), nil)
	wNil := httptest.NewRecorder()
	h.ServeHTTP(wNil, reqNil)

	if wNil.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for nil UUID, got %d: %s", wNil.Code, wNil.Body.String())
	}
}

func TestMitigationBriefPDF_NotFound(t *testing.T) {
	unknownID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	provider := &mockMitigationProvider{
		candidates: []prioritize.CandidateSite{},
	}
	h := NewRouter(Dependencies{
		MitigationProvider: provider,
		Ready:              func(context.Context) error { return nil },
	})

	path := fmt.Sprintf("/v1/mitigation/%s/brief.pdf", unknownID.String())
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}

	var env ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("expected JSON error envelope, got: %v", err)
	}
	if env.Status != Error {
		t.Errorf("expected error status, got %q", env.Status)
	}
}

func TestMitigationBriefPDF_Unavailable(t *testing.T) {
	siteID := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	// Case 1: Nil provider
	hNil := NewRouter(Dependencies{
		Ready: func(context.Context) error { return nil },
	})
	path := fmt.Sprintf("/v1/mitigation/%s/brief.pdf", siteID.String())
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	hNil.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for nil provider, got %d: %s", w.Code, w.Body.String())
	}

	// Case 2: Provider returns error
	providerErr := &mockMitigationProvider{err: errors.New("db pool down")}
	hErr := NewRouter(Dependencies{
		MitigationProvider: providerErr,
		Ready:              func(context.Context) error { return nil },
	})
	req2 := httptest.NewRequest("GET", path, nil)
	w2 := httptest.NewRecorder()
	hErr.ServeHTTP(w2, req2)

	if w2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for provider error, got %d: %s", w2.Code, w2.Body.String())
	}

	if strings.Contains(w2.Body.String(), "db pool down") {
		t.Errorf("internal database error must not leak into response")
	}
}
