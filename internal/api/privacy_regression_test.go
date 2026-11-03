package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"corridor/internal/auth"
	"corridor/internal/db"
	"corridor/internal/prioritize"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
)

func TestSensitivePrivacyAndEmptyRisk(t *testing.T) {
	const (
		rawLat  = 44.976543
		rawLon  = -110.684321
		snapLat = 44.977100
		snapLon = -110.685000
	)
	analystKey := "ranger-precision-key"
	authenticator := auth.NewAuthenticator(auth.Config{
		Secret: "test-secret-key-32-bytes-auth-jwt",
		APIKeys: map[string]auth.User{
			analystKey: {Subject: "ranger-1", Role: auth.RoleAnalyst, Agency: "NPS"},
		},
	})
	broadcaster := NewBroadcaster()
	snapper := func(context.Context, orb.Point, float64) (uuid.UUID, float64, orb.Point, error) {
		return uuid.MustParse("00000000-0000-0000-0000-000000000010"), 8.5, orb.Point{snapLon, snapLat}, nil
	}
	h := NewRouter(Dependencies{
		Authenticator: authenticator,
		Broadcaster:   broadcaster,
		ReportStore:   newMockReportStore(),
		Snapper:       snapper,
	})

	t.Run("public report generalizes sensitive coordinates", func(t *testing.T) {
		body := reportPayload(t, "Grizzly bear", rawLat, rawLon, time.Now().UTC().Add(-45*24*time.Hour))
		rec := postReport(t, h, body, "")
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var env struct {
			Data FieldReport `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode: %v", err)
		}
		assertGeneralized(t, rawLat, rawLon, env.Data.Latitude, env.Data.Longitude)
		if env.Data.SnappedLatitude == nil || env.Data.SnappedLongitude == nil {
			t.Fatal("expected generalized snapped coordinates")
		}
		assertGeneralized(t, snapLat, snapLon, *env.Data.SnappedLatitude, *env.Data.SnappedLongitude)
		if strings.Contains(rec.Body.String(), "44.976543") || strings.Contains(rec.Body.String(), "-110.684321") {
			t.Fatalf("raw coordinates leaked in public body: %s", rec.Body.String())
		}
	})

	t.Run("analyst report keeps full precision", func(t *testing.T) {
		body := reportPayload(t, "gray_wolf", rawLat, rawLon, time.Now().UTC())
		rec := postReport(t, h, body, analystKey)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var env struct {
			Data FieldReport `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if math.Abs(env.Data.Latitude-rawLat) > 1e-9 || math.Abs(env.Data.Longitude-rawLon) > 1e-9 {
			t.Fatalf("analyst lost precision: lat=%v lon=%v", env.Data.Latitude, env.Data.Longitude)
		}
		if env.Data.SnappedLatitude == nil || math.Abs(*env.Data.SnappedLatitude-snapLat) > 1e-9 {
			t.Fatalf("analyst snapped latitude redacted: %v", env.Data.SnappedLatitude)
		}
	})

	t.Run("public stream delays fresh sensitive reports", func(t *testing.T) {
		server := httptest.NewServer(h)
		defer server.Close()

		publicCh := listenReports(t, server.URL, "")
		analystCh := listenReports(t, server.URL, analystKey)

		freshLat, freshLon := 44.428912, -110.588734
		body := reportPayload(t, "Canada lynx", freshLat, freshLon, time.Now().UTC())
		rec := postReport(t, h, body, "")
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		select {
		case got := <-analystCh:
			if math.Abs(got.Latitude-freshLat) > 1e-6 || math.Abs(got.Longitude-freshLon) > 1e-6 {
				t.Fatalf("analyst stream expected raw coordinates, got %v,%v", got.Latitude, got.Longitude)
			}
			if got.SpeciesCommonName != "Canada lynx" {
				t.Fatalf("unexpected species %q", got.SpeciesCommonName)
			}
		case <-time.After(time.Second):
			t.Fatal("analyst stream did not receive the fresh sensitive report")
		}
		select {
		case got := <-publicCh:
			t.Fatalf("public stream received a quarantined sensitive report: %+v", got)
		case <-time.After(250 * time.Millisecond):
		}
	})

	t.Run("unmonitored hotspots are not zero risk", func(t *testing.T) {
		id := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
		emptyHotspots := NewRouter(Dependencies{
			Store: &mockSegmentStore{
				segments: []db.SegmentSummary{{
					ID:          id,
					Coordinates: [][]float64{{-110.70, 44.40}, {-110.69, 44.41}},
					RiskScore:   0,
				}},
			},
		})
		rec := request(emptyHotspots, "GET", "/v1/hotspots")
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 for unmonitored segments, got %d: %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), `"score"`) || strings.Contains(strings.ToLower(rec.Body.String()), "low") || strings.Contains(rec.Body.String(), "Not Significant") {
			t.Fatalf("empty observations encoded as zero risk: %s", rec.Body.String())
		}

		failed := NewRouter(Dependencies{Store: &mockSegmentStore{err: errHotspotStore}})
		failRec := request(failed, "GET", "/v1/hotspots")
		if failRec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 when hotspot store fails, got %d: %s", failRec.Code, failRec.Body.String())
		}
		if strings.Contains(failRec.Body.String(), "ca-120") || strings.Contains(failRec.Body.String(), "private") {
			t.Fatalf("unavailable store fell back to reference hotspots: %s", failRec.Body.String())
		}

		far := NewRouter(Dependencies{
			Store: &mockSegmentStore{
				segments: []db.SegmentSummary{{
					ID:          id,
					Coordinates: [][]float64{{-110.70, 44.40}, {-110.69, 44.41}},
					RiskScore:   2.5,
				}},
			},
		})
		miss := request(far, "GET", "/v1/hotspots?bbox=-82.0,26.0,-81.0,27.0")
		if miss.Code != http.StatusOK {
			t.Fatalf("expected 200 for a bbox with no segments, got %d: %s", miss.Code, miss.Body.String())
		}
		var env hotspotsResponse
		if err := json.Unmarshal(miss.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(env.Data) != 0 {
			t.Fatalf("expected no hotspot items outside coverage, got %d", len(env.Data))
		}
	})

	t.Run("mitigation does not publish zero collisions", func(t *testing.T) {
		provider := &mockMitigationProvider{
			candidates: []prioritize.CandidateSite{{
				ID:                        uuid.New(),
				Segments:                  []uuid.UUID{uuid.New()},
				State:                     "WY",
				HighwayRef:                "US-191",
				RecommendedType:           "fencing",
				ExpectedAvoidedCollisions: 0,
				BCR:                       0,
				Rationale:                 "no carcass observations",
			}},
		}
		router := NewRouter(Dependencies{MitigationProvider: provider})
		rec := request(router, "GET", "/v1/mitigation")
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 for zero-collision candidates, got %d: %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "expected_avoided_collisions") || strings.Contains(rec.Body.String(), `"bcr":0`) {
			t.Fatalf("zero collisions serialized as a mitigation site: %s", rec.Body.String())
		}
	})
}

var errHotspotStore = errString("segment store unavailable")

type errString string

func (e errString) Error() string { return string(e) }

func reportPayload(t *testing.T, species string, lat, lon float64, observed time.Time) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"id":                  uuid.NewString(),
		"species_common_name": species,
		"latitude":            lat,
		"longitude":           lon,
		"observed_at":         observed.Format(time.RFC3339),
		"confidence":          "confirmed",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return body
}

func postReport(t *testing.T, h http.Handler, body []byte, apiKey string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertGeneralized(t *testing.T, rawLat, rawLon, gotLat, gotLon float64) {
	t.Helper()
	wantLat, wantLon, err := generalizeToH3(rawLat, rawLon, publicH3Resolution)
	if err != nil {
		t.Fatalf("h3: %v", err)
	}
	if math.Abs(gotLat-rawLat) < 1e-4 && math.Abs(gotLon-rawLon) < 1e-4 {
		t.Fatalf("public response kept raw coordinates %v,%v", gotLat, gotLon)
	}
	if math.Abs(gotLat-wantLat) > 1e-6 || math.Abs(gotLon-wantLon) > 1e-6 {
		t.Fatalf("expected H3 resolution 6 centroid %v,%v, got %v,%v", wantLat, wantLon, gotLat, gotLon)
	}
}

func listenReports(t *testing.T, baseURL, apiKey string) <-chan FieldReport {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/stream/reports", nil)
	if err != nil {
		t.Fatalf("stream request: %v", err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stream connect: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stream status %d", resp.StatusCode)
	}
	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "connected") {
		t.Fatalf("stream handshake %q (%v)", line, err)
	}
	out := make(chan FieldReport, 1)
	go func() {
		var data string
		for {
			next, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			trimmed := strings.TrimSpace(next)
			if strings.HasPrefix(trimmed, "data:") {
				data = strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			}
			if trimmed == "" && data != "" {
				var report FieldReport
				if json.Unmarshal([]byte(data), &report) == nil {
					out <- report
				}
				return
			}
		}
	}()
	return out
}
