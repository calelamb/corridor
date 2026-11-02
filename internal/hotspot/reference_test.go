package hotspot_test

import (
	"strings"
	"testing"

	"corridor/internal/hotspot"
)

func TestReferenceYosemiteCorridor(t *testing.T) {
	c := hotspot.ReferenceYosemiteCorridor()

	if c.HighwayRef != "CA-120" {
		t.Errorf("expected highway CA-120, got %s", c.HighwayRef)
	}
	if c.State != "CA" {
		t.Errorf("expected state CA, got %s", c.State)
	}
	if c.Park != "Yosemite" {
		t.Errorf("expected park Yosemite, got %s", c.Park)
	}
	if len(c.Graph.Segments()) != 30 {
		t.Fatalf("expected 30 segments, got %d", len(c.Graph.Segments()))
	}
	if len(c.Events) == 0 {
		t.Fatalf("expected events, got 0")
	}

	// Verify events contain bear and deer tags
	var hasBear, hasDeer bool
	for _, ev := range c.Events {
		if strings.Contains(ev.ID, "bear") {
			hasBear = true
		}
		if strings.Contains(ev.ID, "deer") {
			hasDeer = true
		}
	}
	if !hasBear {
		t.Errorf("expected bear crossing events")
	}
	if !hasDeer {
		t.Errorf("expected deer crossing events")
	}

	// Run AnalyzeHotspots with NKDE
	results, err := c.Graph.AnalyzeHotspots(c.Events, "nkde", 500.0)
	if err != nil {
		t.Fatalf("AnalyzeHotspots nkde failed: %v", err)
	}
	if len(results) != 30 {
		t.Fatalf("expected 30 results, got %d", len(results))
	}

	// Top result should be the Crane Flat bear cluster (ca-120-seg-04)
	if results[0].SegmentID != "ca-120-seg-04" {
		t.Errorf("expected top segment ca-120-seg-04, got %s", results[0].SegmentID)
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", results[0].Score)
	}
	if results[0].PValue == nil || *results[0].PValue > 0.05 {
		t.Errorf("expected significant p-value for top cluster, got %v", results[0].PValue)
	}
	if !strings.Contains(results[0].Significance, "Hotspot") {
		t.Errorf("expected Hotspot significance tier, got %s", results[0].Significance)
	}

	// Run AnalyzeHotspots with Getis-Ord
	giResults, err := c.Graph.AnalyzeHotspots(c.Events, "getis_ord", 500.0)
	if err != nil {
		t.Fatalf("AnalyzeHotspots getis_ord failed: %v", err)
	}
	if len(giResults) != 30 {
		t.Fatalf("expected 30 giResults, got %d", len(giResults))
	}
	if giResults[0].ZScore == nil || *giResults[0].ZScore <= 0 {
		t.Errorf("expected positive z-score for top cluster")
	}
}

func TestReferenceWyomingCorridor(t *testing.T) {
	c := hotspot.ReferenceWyomingCorridor()

	if c.HighwayRef != "US-191" {
		t.Errorf("expected highway US-191, got %s", c.HighwayRef)
	}
	if c.State != "WY" {
		t.Errorf("expected state WY, got %s", c.State)
	}
	if len(c.Graph.Segments()) != 25 {
		t.Fatalf("expected 25 segments, got %d", len(c.Graph.Segments()))
	}

	results, err := c.Graph.AnalyzeHotspots(c.Events, "nkde", 500.0)
	if err != nil {
		t.Fatalf("AnalyzeHotspots failed: %v", err)
	}
	if results[0].SegmentID != "seg-12" {
		t.Errorf("expected top segment seg-12, got %s", results[0].SegmentID)
	}
}

func TestFindReferenceCorridor(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		park       string
		wantState  string
		wantPark   string
		wantHwy    string
		shouldFind bool
	}{
		{"YOSE by code", "", "YOSE", "CA", "Yosemite", "CA-120", true},
		{"YOSE by name", "", "Yosemite National Park", "CA", "Yosemite", "CA-120", true},
		{"YOSE by state", "CA", "", "CA", "Yosemite", "CA-120", true},
		{"YELL by code", "", "YELL", "WY", "Yellowstone", "US-191", true},
		{"YELL by name", "", "Yellowstone", "WY", "Yellowstone", "US-191", true},
		{"YELL by state", "WY", "", "WY", "Yellowstone", "US-191", true},
		{"GRTE by code", "", "GRTE", "WY", "Grand Teton", "US-89", true},
		{"GRTE by name", "", "Grand Teton", "WY", "Grand Teton", "US-89", true},
		{"GLAC by code", "", "GLAC", "MT", "Glacier", "US-2", true},
		{"GLAC by name", "", "Glacier National Park", "MT", "Glacier", "US-2", true},
		{"GLAC by state", "MT", "", "MT", "Glacier", "US-2", true},
		{"ROMO by code", "", "ROMO", "CO", "Rocky Mountain", "US-34", true},
		{"ROMO by name", "", "Rocky Mountain", "CO", "Rocky Mountain", "US-34", true},
		{"ROMO by state", "CO", "", "CO", "Rocky Mountain", "US-34", true},
		{"Unknown", "FL", "Everglades", "", "", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := hotspot.FindReferenceCorridor(tc.state, tc.park)
			if !tc.shouldFind {
				if c != nil {
					t.Fatalf("expected nil for %s/%s, got %+v", tc.state, tc.park, c)
				}
				return
			}
			if c == nil {
				t.Fatalf("expected corridor for %s/%s, got nil", tc.state, tc.park)
			}
			if c.State != tc.wantState {
				t.Errorf("expected state %s, got %s", tc.wantState, c.State)
			}
			if c.Park != tc.wantPark {
				t.Errorf("expected park %s, got %s", tc.wantPark, c.Park)
			}
			if c.HighwayRef != tc.wantHwy {
				t.Errorf("expected highway %s, got %s", tc.wantHwy, c.HighwayRef)
			}
		})
	}
}

func TestAllReferenceCorridors(t *testing.T) {
	corridors := hotspot.AllReferenceCorridors()
	if len(corridors) != 5 {
		t.Fatalf("expected 5 corridors, got %d", len(corridors))
	}
	expectedParks := map[string]bool{
		"Yellowstone":    false,
		"Grand Teton":    false,
		"Glacier":        false,
		"Rocky Mountain": false,
		"Yosemite":       false,
	}
	for _, c := range corridors {
		expectedParks[c.Park] = true
		if len(c.Graph.Segments()) == 0 {
			t.Errorf("corridor %s has empty graph", c.Name)
		}
		if len(c.Events) == 0 {
			t.Errorf("corridor %s has no events", c.Name)
		}
	}
	for park, found := range expectedParks {
		if !found {
			t.Errorf("expected corridor for park %s", park)
		}
	}
}

func TestBBoxOverlapsAndSegmentInBBox(t *testing.T) {
	bboxYose := [4]float64{-119.85, 37.74, -119.34, 37.89}
	queryExact := [4]float64{-119.85, 37.74, -119.34, 37.89}
	queryOutside := [4]float64{-80.0, 25.0, -79.0, 26.0}

	if !hotspot.CorridorOverlapsBBox(bboxYose, queryExact) {
		t.Errorf("expected overlap")
	}
	if hotspot.CorridorOverlapsBBox(bboxYose, queryOutside) {
		t.Errorf("did not expect overlap")
	}

	coords := [][2]float64{
		{-119.8142, 37.7505},
		{-119.8000, 37.7600},
	}
	if !hotspot.SegmentInBBox(coords, bboxYose) {
		t.Errorf("expected segment inside bbox")
	}
	if hotspot.SegmentInBBox(coords, queryOutside) {
		t.Errorf("did not expect segment inside Florida bbox")
	}
}
