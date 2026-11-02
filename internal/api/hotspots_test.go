package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"corridor/internal/db"

	"github.com/google/uuid"
)

type hotspotsResponse struct {
	Status string        `json:"status"`
	Data   []HotspotItem `json:"data"`
	Error  any           `json:"error"`
	Meta   any           `json:"meta"`
}

func TestHotspots_Default(t *testing.T) {
	h := NewRouter(Dependencies{
		Store: testStore{},
		Ready: func(context.Context) error { return nil },
	})

	r := request(h, "GET", "/v1/hotspots")
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", r.Code, r.Body.String())
	}

	var resp hotspotsResponse
	if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected status success, got %s", resp.Status)
	}
	if len(resp.Data) == 0 {
		t.Fatalf("expected hotspot items, got 0")
	}

	// Verify fields on items
	for i, item := range resp.Data {
		if item.SegmentId == "" {
			t.Errorf("item %d missing SegmentId", i)
		}
		if item.Score < 0 {
			t.Errorf("item %d has negative score: %f", i, item.Score)
		}
		if item.PValue == nil {
			t.Errorf("item %d missing PValue", i)
		}
		if item.ZScore == nil {
			t.Errorf("item %d missing ZScore", i)
		}
		if item.Significance == nil || *item.Significance == "" {
			t.Errorf("item %d missing Significance", i)
		}
		if item.SignificanceTier == nil || *item.SignificanceTier == "" {
			t.Errorf("item %d missing SignificanceTier", i)
		}
		if item.Method == nil || *item.Method != "nkde" {
			t.Errorf("item %d expected method nkde, got %v", i, item.Method)
		}
		if item.RawCount == nil {
			t.Errorf("item %d missing RawCount", i)
		}
	}

	// Verify sorted descending by Score
	for i := 1; i < len(resp.Data); i++ {
		if resp.Data[i].Score > resp.Data[0].Score {
			t.Errorf("items not sorted descending: [0]=%f, [%d]=%f",
				resp.Data[0].Score, i, resp.Data[i].Score)
		}
	}
}

func TestHotspots_MethodSelection(t *testing.T) {
	h := NewRouter(Dependencies{
		Store: testStore{},
		Ready: func(context.Context) error { return nil },
	})

	methods := []string{"nkde", "getis_ord", "density"}
	for _, m := range methods {
		t.Run("method="+m, func(t *testing.T) {
			r := request(h, "GET", "/v1/hotspots?method="+m)
			if r.Code != http.StatusOK {
				t.Fatalf("expected 200 for method %s, got %d: %s", m, r.Code, r.Body.String())
			}

			var resp hotspotsResponse
			if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal json: %v", err)
			}
			if len(resp.Data) == 0 {
				t.Fatalf("expected items for method %s, got 0", m)
			}
			if resp.Data[0].Method == nil || *resp.Data[0].Method != m {
				t.Errorf("expected item method %q, got %v", m, resp.Data[0].Method)
			}
			if resp.Data[0].Significance == nil || *resp.Data[0].Significance == "" {
				t.Errorf("expected significance tier")
			}
		})
	}
}

func TestHotspots_BandwidthParameter(t *testing.T) {
	h := NewRouter(Dependencies{
		Store: testStore{},
		Ready: func(context.Context) error { return nil },
	})

	for _, bw := range []string{"250", "500", "1000"} {
		t.Run("bandwidth="+bw, func(t *testing.T) {
			r := request(h, "GET", "/v1/hotspots?bandwidth="+bw)
			if r.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", r.Code)
			}
			var resp hotspotsResponse
			if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal json: %v", err)
			}
			if len(resp.Data) == 0 {
				t.Fatalf("expected items, got 0")
			}
		})
	}
}

func TestHotspots_BBoxFiltering(t *testing.T) {
	h := NewRouter(Dependencies{
		Store: testStore{},
		Ready: func(context.Context) error { return nil },
	})

	t.Run("Yosemite CA-120 BBox", func(t *testing.T) {
		r := request(h, "GET", "/v1/hotspots?bbox=-119.85,37.74,-119.34,37.89")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", r.Code)
		}
		var resp hotspotsResponse
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal json: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Fatalf("expected Yosemite segments, got 0")
		}
		for _, item := range resp.Data {
			if !strings.HasPrefix(item.SegmentId, "ca-120-seg-") {
				t.Errorf("expected only ca-120-seg- items in Yosemite bbox, got %s", item.SegmentId)
			}
		}
	})

	t.Run("Wyoming US-191 BBox", func(t *testing.T) {
		r := request(h, "GET", "/v1/hotspots?bbox=-110.75,44.38,-110.65,44.46")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", r.Code)
		}
		var resp hotspotsResponse
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal json: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Fatalf("expected Wyoming segments, got 0")
		}
		for _, item := range resp.Data {
			if !strings.HasPrefix(item.SegmentId, "seg-") {
				t.Errorf("expected only seg- items in Wyoming bbox, got %s", item.SegmentId)
			}
		}
	})

	t.Run("Disjoint BBox returns empty list", func(t *testing.T) {
		r := request(h, "GET", "/v1/hotspots?bbox=-82.0,26.0,-81.0,27.0")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", r.Code)
		}
		var resp hotspotsResponse
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal json: %v", err)
		}
		if len(resp.Data) != 0 {
			t.Fatalf("expected 0 items for Florida bbox, got %d", len(resp.Data))
		}
	})
}

func TestHotspots_ValidationErrors(t *testing.T) {
	h := NewRouter(Dependencies{
		Store: testStore{},
		Ready: func(context.Context) error { return nil },
	})

	tests := []struct {
		name string
		url  string
	}{
		{"invalid method", "/v1/hotspots?method=invalid"},
		{"negative bandwidth", "/v1/hotspots?bandwidth=-10"},
		{"zero bandwidth", "/v1/hotspots?bandwidth=0"},
		{"invalid bbox string", "/v1/hotspots?bbox=not-a-bbox"},
		{"incomplete bbox", "/v1/hotspots?bbox=-110,44,-109"},
		{"inverted longitude bbox", "/v1/hotspots?bbox=10,20,-10,30"},
		{"inverted latitude bbox", "/v1/hotspots?bbox=-110,45,-109,44"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := request(h, "GET", tc.url)
			if r.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %s, got %d: %s", tc.url, r.Code, r.Body.String())
			}
		})
	}
}

func TestHotspots_WithDatabaseStore(t *testing.T) {
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mockStore := &mockSegmentStore{
		segments: []db.SegmentSummary{
			{
				ID:             id1,
				Coordinates:    [][]float64{{-110.70, 44.40}, {-110.70, 44.41}},
				RiskScore:      4.5,
				PercentileRank: 98.0,
			},
			{
				ID:             id2,
				Coordinates:    [][]float64{{-110.70, 44.41}, {-110.70, 44.42}},
				RiskScore:      1.2,
				PercentileRank: 80.0,
			},
		},
	}

	h := NewRouter(Dependencies{
		Store: mockStore,
		Ready: func(context.Context) error { return nil },
	})

	r := request(h, "GET", "/v1/hotspots")
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200 with DB store, got %d: %s", r.Code, r.Body.String())
	}

	var resp hotspotsResponse
	if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 hotspot items from DB segments, got %d", len(resp.Data))
	}
	// Highest risk segment (id1) should be first
	if resp.Data[0].SegmentId != id1.String() {
		t.Errorf("expected top segment %s, got %s", id1.String(), resp.Data[0].SegmentId)
	}
}

type mockHotspotProvider struct {
	items []HotspotItem
	err   error
}

func (m *mockHotspotProvider) Hotspots(ctx context.Context, params HotspotsParams) ([]HotspotItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func TestHotspots_CustomProvider(t *testing.T) {
	testSig := "Hotspot (99% confidence)"
	testMethod := "nkde"
	customItems := []HotspotItem{
		{
			SegmentId:        "custom-seg-01",
			Score:            9.99,
			Significance:     &testSig,
			SignificanceTier: &testSig,
			Method:           &testMethod,
		},
	}

	h := NewRouter(Dependencies{
		HotspotProvider: &mockHotspotProvider{items: customItems},
		Ready:           func(context.Context) error { return nil },
	})

	r := request(h, "GET", "/v1/hotspots")
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", r.Code)
	}

	var resp hotspotsResponse
	if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].SegmentId != "custom-seg-01" {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}

	// Provider error returns 503
	hFail := NewRouter(Dependencies{
		HotspotProvider: &mockHotspotProvider{err: errors.New("provider failure")},
		Ready:           func(context.Context) error { return nil },
	})
	rFail := request(hFail, "GET", "/v1/hotspots")
	if rFail.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on provider failure, got %d", rFail.Code)
	}
}
