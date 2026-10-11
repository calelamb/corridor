package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"corridor/internal/db"

	"github.com/google/uuid"
)

func TestSegmentsEndpoint(t *testing.T) {
	// Create router with default dependencies (mock store)
	mockStore := newMockSegmentStore()
	h := NewRouter(Dependencies{
		Store: mockStore,
		Ready: func(context.Context) error { return nil },
	})

	t.Run("GET /v1/segments returns 200 with segment list", func(t *testing.T) {
		r := request(h, "GET", "/v1/segments")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", r.Code, r.Body.String())
		}

		var env struct {
			Status string `json:"status"`
			Data   []struct {
				ID             string  `json:"id"`
				RiskScore      float64 `json:"risk_score"`
				PercentileRank float64 `json:"percentile_rank"`
				P10            float64 `json:"p10"`
				P90            float64 `json:"p90"`
				Geometry       struct {
					Type        string      `json:"type"`
					Coordinates [][]float64 `json:"coordinates"`
				} `json:"geometry"`
			} `json:"data"`
			Error any `json:"error"`
			Meta  any `json:"meta"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if env.Status != "success" {
			t.Fatalf("expected status 'success', got %q", env.Status)
		}
		if len(env.Data) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(env.Data))
		}
		if env.Data[0].Geometry.Type != "LineString" {
			t.Fatalf("expected LineString geometry, got %s", env.Data[0].Geometry.Type)
		}
		if len(env.Data[0].Geometry.Coordinates) < 2 {
			t.Fatalf("expected at least 2 coordinate points, got %d", len(env.Data[0].Geometry.Coordinates))
		}
	})

	t.Run("GET /v1/segments with min_risk filter", func(t *testing.T) {
		r := request(h, "GET", "/v1/segments?min_risk=0.7")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", r.Code, r.Body.String())
		}
		var env struct {
			Status string `json:"status"`
			Data   []struct {
				ID        string  `json:"id"`
				RiskScore float64 `json:"risk_score"`
			} `json:"data"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(env.Data) != 1 {
			t.Fatalf("expected 1 segment matching min_risk=0.7, got %d", len(env.Data))
		}
		if env.Data[0].RiskScore < 0.7 {
			t.Fatalf("segment risk score %f < 0.7", env.Data[0].RiskScore)
		}
	})

	t.Run("GET /v1/segments with bbox filter", func(t *testing.T) {
		// BBox covering only segment 1: -110.5,43.5,-110.0,44.0
		r := request(h, "GET", "/v1/segments?bbox=-110.5,43.5,-110.0,44.0")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", r.Code, r.Body.String())
		}
		var env struct {
			Status string `json:"status"`
			Data   []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(env.Data) != 1 {
			t.Fatalf("expected 1 segment in bbox, got %d", len(env.Data))
		}
		if env.Data[0].ID != "00000000-0000-0000-0000-000000000001" {
			t.Fatalf("expected segment 1, got %s", env.Data[0].ID)
		}
	})

	t.Run("GET /v1/segments with invalid bbox returns 400", func(t *testing.T) {
		for _, invalidBbox := range []string{"invalid", "1,2,3", "10,20,5,40", "181,0,182,0"} {
			r := request(h, "GET", "/v1/segments?bbox="+invalidBbox)
			if r.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for bbox %q, got %d", invalidBbox, r.Code)
			}
		}
	})

	t.Run("GET /v1/segments with invalid min_risk returns 400", func(t *testing.T) {
		for _, invalidRisk := range []string{"not-a-number", "-0.1"} {
			r := request(h, "GET", "/v1/segments?min_risk="+invalidRisk)
			if r.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for min_risk %q, got %d", invalidRisk, r.Code)
			}
		}
	})

	t.Run("GET /v1/segments/{id} returns 200 with risk details and SHAP drivers", func(t *testing.T) {
		r := request(h, "GET", "/v1/segments/00000000-0000-0000-0000-000000000001")
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", r.Code, r.Body.String())
		}

		var env struct {
			Status string `json:"status"`
			Data   struct {
				ID               string             `json:"id"`
				RiskScore        float64            `json:"risk_score"`
				PercentileRank   float64            `json:"percentile_rank"`
				P10              float64            `json:"p10"`
				P90              float64            `json:"p90"`
				MonthlyProfile   map[string]float64 `json:"monthly_profile"`
				SpeciesBreakdown map[string]float64 `json:"species_breakdown"`
				ShapDrivers      []struct {
					Feature string  `json:"feature"`
					Impact  string  `json:"impact"`
					Weight  float64 `json:"weight"`
				} `json:"shap_drivers"`
			} `json:"data"`
			Error any `json:"error"`
			Meta  any `json:"meta"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to parse detail response: %v", err)
		}
		if env.Status != "success" {
			t.Fatalf("expected status 'success', got %q", env.Status)
		}
		if env.Data.ID != "00000000-0000-0000-0000-000000000001" {
			t.Fatalf("expected segment ID 1, got %s", env.Data.ID)
		}
		if len(env.Data.MonthlyProfile) != 12 {
			t.Fatalf("expected 12 months in profile, got %d", len(env.Data.MonthlyProfile))
		}
		if env.Data.SpeciesBreakdown["mule_deer"] <= 0 {
			t.Fatalf("expected mule_deer probability > 0, got %f", env.Data.SpeciesBreakdown["mule_deer"])
		}
		if len(env.Data.ShapDrivers) == 0 {
			t.Fatal("expected at least one SHAP driver")
		}
		driver := env.Data.ShapDrivers[0]
		if driver.Feature != "Canopy cover" || driver.Impact != "High attraction" || driver.Weight != 0.45 {
			t.Fatalf("unexpected SHAP driver content: %+v", driver)
		}
	})

	t.Run("GET /v1/segments/{id} returns 404 for unknown segment ID", func(t *testing.T) {
		r := request(h, "GET", "/v1/segments/00000000-0000-0000-0000-000000000999")
		if r.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", r.Code, r.Body.String())
		}
		var env struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to parse 404 response: %v", err)
		}
		if env.Status != "error" {
			t.Fatalf("expected error status, got %s", env.Status)
		}
		if !strings.Contains(strings.ToLower(env.Error), "not found") {
			t.Fatalf("expected 'not found' error message, got %s", env.Error)
		}
	})

	t.Run("GET /v1/segments/{id} returns 400 for invalid UUID", func(t *testing.T) {
		r := request(h, "GET", "/v1/segments/not-a-uuid")
		if r.Code != http.StatusBadRequest && r.Code != http.StatusNotFound {
			t.Fatalf("expected 400 or 404, got %d: %s", r.Code, r.Body.String())
		}
	})

	t.Run("GET /v1/segments returns 503 when store is unavailable or fails", func(t *testing.T) {
		hFail := NewRouter(Dependencies{
			Store: &mockSegmentStore{err: errors.New("db down")},
			Ready: func(context.Context) error { return nil },
		})
		r := request(hFail, "GET", "/v1/segments")
		if r.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", r.Code)
		}
		rDetail := request(hFail, "GET", "/v1/segments/00000000-0000-0000-0000-000000000001")
		if rDetail.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", rDetail.Code)
		}
	})
}

type mockSegmentStore struct {
	testStore
	segments []db.SegmentSummary
	details  map[uuid.UUID]db.SegmentDetail
	err      error
}

func (m *mockSegmentStore) Segments(ctx context.Context, filter db.SegmentFilter) ([]db.SegmentSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []db.SegmentSummary
	for _, seg := range m.segments {
		if filter.MinRisk != nil && seg.RiskScore < *filter.MinRisk {
			continue
		}
		if filter.BBox != nil {
			bbox := filter.BBox
			minLon, minLat, maxLon, maxLat := bbox[0], bbox[1], bbox[2], bbox[3]
			inside := false
			for _, pt := range seg.Coordinates {
				if pt[0] >= minLon && pt[0] <= maxLon && pt[1] >= minLat && pt[1] <= maxLat {
					inside = true
					break
				}
			}
			if !inside {
				continue
			}
		}
		res = append(res, seg)
	}
	return res, nil
}

func (m *mockSegmentStore) Segment(ctx context.Context, id uuid.UUID) (*db.SegmentDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	d, ok := m.details[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	return &d, nil
}

func newMockSegmentStore() *mockSegmentStore {
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	hw1 := "US-191"
	hw2 := "WY-28"
	st := "WY"

	monthly1 := make(map[string]float64)
	monthly2 := make(map[string]float64)
	for i := 1; i <= 12; i++ {
		monthly1[fmt.Sprintf("%d", i)] = 0.05 + float64(i)*0.01
		monthly2[fmt.Sprintf("%d", i)] = 0.02 + float64(i)*0.005
	}

	seg1 := db.SegmentSummary{
		ID:             id1,
		Coordinates:    [][]float64{{-110.2, 43.8}, {-110.1, 43.9}},
		RiskScore:      0.85,
		PercentileRank: 95.0,
		P10:            0.65,
		P90:            1.05,
		HighwayRef:     &hw1,
		State:          &st,
	}
	seg2 := db.SegmentSummary{
		ID:             id2,
		Coordinates:    [][]float64{{-108.0, 42.0}, {-107.9, 42.1}},
		RiskScore:      0.35,
		PercentileRank: 40.0,
		P10:            0.20,
		P90:            0.50,
		HighwayRef:     &hw2,
		State:          &st,
	}

	detail1 := db.SegmentDetail{
		ID:             id1,
		Coordinates:    [][]float64{{-110.2, 43.8}, {-110.1, 43.9}},
		RiskScore:      0.85,
		PercentileRank: 95.0,
		P10:            0.65,
		P90:            1.05,
		HighwayRef:     &hw1,
		State:          &st,
		MonthlyProfile: monthly1,
		SpeciesBreakdown: map[string]float64{
			"mule_deer": 0.75,
			"elk":       0.25,
		},
		ShapDrivers: []db.ShapDriver{
			{Feature: "Canopy cover", Impact: "High attraction", Weight: 0.45},
			{Feature: "Distance to water", Impact: "Medium attraction", Weight: 0.25},
		},
	}
	detail2 := db.SegmentDetail{
		ID:             id2,
		Coordinates:    [][]float64{{-108.0, 42.0}, {-107.9, 42.1}},
		RiskScore:      0.35,
		PercentileRank: 40.0,
		P10:            0.20,
		P90:            0.50,
		HighwayRef:     &hw2,
		State:          &st,
		MonthlyProfile: monthly2,
		SpeciesBreakdown: map[string]float64{
			"pronghorn": 0.80,
			"mule_deer": 0.20,
		},
		ShapDrivers: []db.ShapDriver{
			{Feature: "Traffic volume", Impact: "Medium avoidance", Weight: 0.30},
		},
	}

	return &mockSegmentStore{
		segments: []db.SegmentSummary{seg1, seg2},
		details: map[uuid.UUID]db.SegmentDetail{
			id1: detail1,
			id2: detail2,
		},
	}
}

