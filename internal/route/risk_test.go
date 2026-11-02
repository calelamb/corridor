package route

import (
	"math"
	"strings"
	"testing"
)

// TestCalculateDiurnalMultiplier verifies that diurnal movement multipliers match
// published empirical ungulate activity rhythms (Huijser et al. 2009, Santos et al. 2011).
func TestCalculateDiurnalMultiplier(t *testing.T) {
	tests := []struct {
		name         string
		hour         int
		expectedMult float64
		expectedPeak bool
		expectedName string
	}{
		// Dawn: 5:00 - 8:59 (2.1x, peak)
		{name: "dawn start hour 5", hour: 5, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},
		{name: "dawn mid hour 6", hour: 6, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},
		{name: "dawn mid hour 7", hour: 7, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},
		{name: "dawn end hour 8", hour: 8, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},

		// Day: 9:00 - 17:59 (1.0x, non-peak)
		{name: "day start hour 9", hour: 9, expectedMult: 1.0, expectedPeak: false, expectedName: "Day"},
		{name: "day mid hour 12", hour: 12, expectedMult: 1.0, expectedPeak: false, expectedName: "Day"},
		{name: "day mid hour 15", hour: 15, expectedMult: 1.0, expectedPeak: false, expectedName: "Day"},
		{name: "day end hour 17", hour: 17, expectedMult: 1.0, expectedPeak: false, expectedName: "Day"},

		// Dusk: 18:00 - 22:59 (2.8x, peak)
		{name: "dusk start hour 18", hour: 18, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "dusk mid hour 19", hour: 19, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "dusk mid hour 20", hour: 20, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "dusk mid hour 21", hour: 21, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "dusk end hour 22", hour: 22, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},

		// Night: 23:00 - 4:59 (1.5x, non-peak)
		{name: "night start hour 23", hour: 23, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "night midnight hour 0", hour: 0, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "night early hour 1", hour: 1, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "night mid hour 3", hour: 3, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "night end hour 4", hour: 4, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},

		// Out-of-range hours normalized via mod 24
		{name: "out of range 24 maps to 0 (Night)", hour: 24, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "out of range 29 maps to 5 (Dawn)", hour: 29, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},
		{name: "out of range 42 maps to 18 (Dusk)", hour: 42, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "out of range -1 maps to 23 (Night)", hour: -1, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
		{name: "out of range -6 maps to 18 (Dusk)", hour: -6, expectedMult: 2.8, expectedPeak: true, expectedName: "Dusk"},
		{name: "out of range -18 maps to 6 (Dawn)", hour: -18, expectedMult: 2.1, expectedPeak: true, expectedName: "Dawn"},
		{name: "out of range 100 maps to 4 (Night)", hour: 100, expectedMult: 1.5, expectedPeak: false, expectedName: "Night"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mult, isPeak := CalculateDiurnalMultiplier(tc.hour)
			if math.Abs(mult-tc.expectedMult) > 1e-4 {
				t.Errorf("hour %d multiplier: expected %.2f, got %.2f", tc.hour, tc.expectedMult, mult)
			}
			if isPeak != tc.expectedPeak {
				t.Errorf("hour %d isPeak: expected %v, got %v", tc.hour, tc.expectedPeak, isPeak)
			}
			periodName := DiurnalPeriodName(tc.hour)
			if periodName != tc.expectedName {
				t.Errorf("hour %d period name: expected %s, got %s", tc.hour, tc.expectedName, periodName)
			}
		})
	}
}

// TestEvaluateRouteRisk_OverallRiskTiers tests overall journey risk tier classifications
// using explicitly synthetic test risk profiles.
func TestEvaluateRouteRisk_OverallRiskTiers(t *testing.T) {
	tests := []struct {
		name         string
		segments     []HazardZone
		hour         int
		expectedTier string
	}{
		{
			name: "synthetic all zero risk profile",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.0},
				{SegmentID: "synth-02", RiskScore: 0.0},
			},
			hour:         12, // Day (1.0x)
			expectedTier: "Low",
		},
		{
			name: "synthetic low risk profile",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.1},
				{SegmentID: "synth-02", RiskScore: 0.15},
				{SegmentID: "synth-03", RiskScore: 0.12},
			},
			hour:         12, // Day (1.0x)
			expectedTier: "Low",
		},
		{
			name: "synthetic moderate risk profile",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.3},
				{SegmentID: "synth-02", RiskScore: 0.35},
				{SegmentID: "synth-03", RiskScore: 0.28},
			},
			hour:         12, // Day (1.0x)
			expectedTier: "Moderate",
		},
		{
			name: "synthetic high risk profile",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.45},
				{SegmentID: "synth-02", RiskScore: 0.52},
				{SegmentID: "synth-03", RiskScore: 0.48},
			},
			hour:         12, // Day (1.0x)
			expectedTier: "High",
		},
		{
			name: "synthetic severe risk profile with critical segment",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.85},
				{SegmentID: "synth-02", RiskScore: 0.65},
			},
			hour:         12, // Day (1.0x)
			expectedTier: "Severe",
		},
		{
			name: "synthetic localized spike elevates overall tier to Severe",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.05},
				{SegmentID: "synth-02", RiskScore: 0.05},
				{SegmentID: "synth-03", RiskScore: 0.05},
				{SegmentID: "synth-04", RiskScore: 0.85}, // Critical hotspot spike
			},
			hour:         12, // Day (1.0x)
			expectedTier: "Severe",
		},
		{
			name: "synthetic diurnal escalation from moderate daytime to severe at dawn",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.35},
				{SegmentID: "synth-02", RiskScore: 0.40},
			},
			hour:         6, // Dawn (2.1x) -> peak reaches 0.40 * 2.1 = 0.84 (Severe)
			expectedTier: "Severe",
		},
		{
			name: "synthetic diurnal escalation to severe at dusk",
			segments: []HazardZone{
				{SegmentID: "synth-01", RiskScore: 0.30},
			},
			hour:         19, // Dusk (2.8x) -> 0.30 * 2.8 = 0.84 (Severe)
			expectedTier: "Severe",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report := EvaluateRouteRisk(tc.segments, tc.hour)
			if report.OverallRiskTier != tc.expectedTier {
				t.Errorf("expected overall tier %s, got %s (avg: %.3f, peak: %.3f)",
					tc.expectedTier, report.OverallRiskTier, report.AverageRiskScore, report.PeakRiskScore)
			}
		})
	}
}

// TestEvaluateRouteRisk_HazardZoneFilteringAndOrdering verifies that segments below
// the hazard threshold are excluded, and returned hazard zones are strictly sorted
// by severity (descending risk score).
func TestEvaluateRouteRisk_HazardZoneFilteringAndOrdering(t *testing.T) {
	// Synthetic test segments covering a wide distribution of baseline risk scores
	syntheticSegments := []HazardZone{
		{SegmentID: "seg-low-1", HighwayRef: "US-191", Milepost: 10.0, RiskScore: 0.10, State: "WY"},
		{SegmentID: "seg-high-1", HighwayRef: "US-191", Milepost: 10.5, RiskScore: 0.70, State: "WY", PrimaryRisk: "Ungulate Migration Corridor"},
		{SegmentID: "seg-crit-1", HighwayRef: "US-191", Milepost: 11.0, RiskScore: 0.90, State: "WY", PrimaryRisk: "Canyon Bottleneck"},
		{SegmentID: "seg-low-2", HighwayRef: "US-191", Milepost: 11.5, RiskScore: 0.20, State: "WY"},
		{SegmentID: "seg-mod-1", HighwayRef: "US-191", Milepost: 12.0, RiskScore: 0.50, State: "WY", PrimaryRisk: "Dense Canopy & Low Visibility"},
	}

	t.Run("daytime evaluation with 1.0x multiplier", func(t *testing.T) {
		report := EvaluateRouteRisk(syntheticSegments, 12)

		// Segments with score < 0.4 should be excluded: seg-low-1 (0.10) and seg-low-2 (0.20)
		if len(report.HazardZones) != 3 {
			t.Fatalf("expected 3 hazard zones, got %d", len(report.HazardZones))
		}

		// Verify ordering by severity descending: seg-crit-1 (0.90) > seg-high-1 (0.70) > seg-mod-1 (0.50)
		expectedIDs := []string{"seg-crit-1", "seg-high-1", "seg-mod-1"}
		expectedTiers := []string{"Critical", "High", "Moderate"}
		expectedScores := []float64{0.90, 0.70, 0.50}

		for i, hz := range report.HazardZones {
			if hz.SegmentID != expectedIDs[i] {
				t.Errorf("hazard zone %d: expected segment ID %s, got %s", i, expectedIDs[i], hz.SegmentID)
			}
			if hz.RiskTier != expectedTiers[i] {
				t.Errorf("hazard zone %d: expected tier %s, got %s", i, expectedTiers[i], hz.RiskTier)
			}
			if math.Abs(hz.RiskScore-expectedScores[i]) > 1e-4 {
				t.Errorf("hazard zone %d: expected score %.2f, got %.2f", i, expectedScores[i], hz.RiskScore)
			}
		}

		// Verify preserves metadata
		if report.HazardZones[0].HighwayRef != "US-191" || report.HazardZones[0].State != "WY" {
			t.Errorf("lost metadata on hazard zone: %+v", report.HazardZones[0])
		}
	})

	t.Run("dawn evaluation with 2.1x multiplier activates additional hazard zones", func(t *testing.T) {
		report := EvaluateRouteRisk(syntheticSegments, 6)

		// At Dawn (2.1x):
		// seg-low-1: 0.10 * 2.1 = 0.21 (below 0.4 threshold -> filtered out)
		// seg-low-2: 0.20 * 2.1 = 0.42 (>= 0.4 threshold -> included as Moderate!)
		// seg-mod-1: 0.50 * 2.1 = 1.05 (Critical)
		// seg-high-1: 0.70 * 2.1 = 1.47 (Critical)
		// seg-crit-1: 0.90 * 2.1 = 1.89 (Critical)
		if len(report.HazardZones) != 4 {
			t.Fatalf("expected 4 hazard zones at dawn, got %d", len(report.HazardZones))
		}

		expectedIDs := []string{"seg-crit-1", "seg-high-1", "seg-mod-1", "seg-low-2"}
		for i, hz := range report.HazardZones {
			if hz.SegmentID != expectedIDs[i] {
				t.Errorf("dawn hazard zone %d: expected %s, got %s", i, expectedIDs[i], hz.SegmentID)
			}
		}
		if report.HazardZones[3].RiskTier != "Moderate" {
			t.Errorf("expected seg-low-2 to be Moderate at dawn, got %s", report.HazardZones[3].RiskTier)
		}
	})
}

// TestEvaluateRouteRisk_EdgeCases tests boundary and empty conditions.
func TestEvaluateRouteRisk_EdgeCases(t *testing.T) {
	t.Run("empty segments slice", func(t *testing.T) {
		report := EvaluateRouteRisk([]HazardZone{}, 14)

		if report.TotalDistanceKm != 0.0 {
			t.Errorf("expected 0.0 distance, got %.2f", report.TotalDistanceKm)
		}
		if report.AverageRiskScore != 0.0 {
			t.Errorf("expected 0.0 average risk, got %.2f", report.AverageRiskScore)
		}
		if report.PeakRiskScore != 0.0 {
			t.Errorf("expected 0.0 peak risk, got %.2f", report.PeakRiskScore)
		}
		if report.OverallRiskTier != "Low" {
			t.Errorf("expected Low overall risk, got %s", report.OverallRiskTier)
		}
		if report.HazardZones == nil || len(report.HazardZones) != 0 {
			t.Errorf("expected non-nil empty hazard zones slice, got %+v", report.HazardZones)
		}
		if !strings.Contains(report.AdvisoryNotice, "No road segments analyzed") {
			t.Errorf("expected advisory for empty route, got %s", report.AdvisoryNotice)
		}
	})

	t.Run("nil segments slice", func(t *testing.T) {
		report := EvaluateRouteRisk(nil, 8)

		if report.HazardZones == nil {
			t.Error("expected non-nil HazardZones slice for JSON compatibility")
		}
		if report.OverallRiskTier != "Low" {
			t.Errorf("expected Low overall risk, got %s", report.OverallRiskTier)
		}
	})

	t.Run("single segment", func(t *testing.T) {
		single := []HazardZone{
			{SegmentID: "solo-1", HighwayRef: "US-89", Milepost: 45.2, RiskScore: 0.65, State: "WY", PrimaryRisk: "Elk Migration"},
		}
		report := EvaluateRouteRisk(single, 12)

		if len(report.HazardZones) != 1 {
			t.Fatalf("expected 1 hazard zone, got %d", len(report.HazardZones))
		}
		if report.OverallRiskTier != "High" {
			t.Errorf("expected High overall risk, got %s", report.OverallRiskTier)
		}
		if report.TotalDistanceKm <= 0 {
			t.Errorf("expected positive distance, got %.3f", report.TotalDistanceKm)
		}
		if !strings.Contains(report.AdvisoryNotice, "1 high-risk wildlife collision zone identified along US-89") {
			t.Errorf("expected singular zone phrasing in notice, got %s", report.AdvisoryNotice)
		}
	})

	t.Run("all zero risk segments", func(t *testing.T) {
		zeroSegs := []HazardZone{
			{SegmentID: "zero-1", HighwayRef: "I-80", RiskScore: 0.0},
			{SegmentID: "zero-2", HighwayRef: "I-80", RiskScore: 0.0},
			{SegmentID: "zero-3", HighwayRef: "I-80", RiskScore: 0.0},
		}
		report := EvaluateRouteRisk(zeroSegs, 6) // Dawn (2.1x)

		if report.AverageRiskScore != 0.0 {
			t.Errorf("expected 0.0 average risk, got %.2f", report.AverageRiskScore)
		}
		if report.PeakRiskScore != 0.0 {
			t.Errorf("expected 0.0 peak risk, got %.2f", report.PeakRiskScore)
		}
		if report.OverallRiskTier != "Low" {
			t.Errorf("expected Low overall risk, got %s", report.OverallRiskTier)
		}
		if len(report.HazardZones) != 0 {
			t.Errorf("expected 0 hazard zones for zero risk segments, got %d", len(report.HazardZones))
		}
		if !strings.Contains(report.AdvisoryNotice, "No elevated wildlife collision zones identified") {
			t.Errorf("expected clean advisory notice, got %s", report.AdvisoryNotice)
		}
	})

	t.Run("hour of day out of range wraps with mod 24", func(t *testing.T) {
		segs := []HazardZone{
			{SegmentID: "seg-1", HighwayRef: "US-191", RiskScore: 0.5},
		}

		repNeg := EvaluateRouteRisk(segs, -1) // -1 maps to 23 (Night: 1.5x)
		if math.Abs(repNeg.DiurnalMultiplier-1.5) > 1e-4 {
			t.Errorf("expected multiplier 1.5 for hour -1, got %.2f", repNeg.DiurnalMultiplier)
		}

		repOver := EvaluateRouteRisk(segs, 29) // 29 maps to 5 (Dawn: 2.1x)
		if math.Abs(repOver.DiurnalMultiplier-2.1) > 1e-4 {
			t.Errorf("expected multiplier 2.1 for hour 29, got %.2f", repOver.DiurnalMultiplier)
		}
		if !repOver.IsPeakWildlifeHour {
			t.Errorf("expected isPeak = true for hour 29")
		}
	})
}

// TestEvaluateRouteRisk_DriverAdvisoryNotice verifies plain-language driver advisory notice formatting.
func TestEvaluateRouteRisk_DriverAdvisoryNotice(t *testing.T) {
	// Synthetic 3-zone cluster along US-191 tested at dawn
	syntheticZones := []HazardZone{
		{SegmentID: "z1", HighwayRef: "US-191", Milepost: 14.1, RiskScore: 0.85, PrimaryRisk: "Ungulate Migration Corridor"},
		{SegmentID: "z2", HighwayRef: "US-191", Milepost: 14.2, RiskScore: 0.80, PrimaryRisk: "Canyon Section"},
		{SegmentID: "z3", HighwayRef: "US-191", Milepost: 14.3, RiskScore: 0.75, PrimaryRisk: "Dense Canopy & Low Visibility"},
	}

	t.Run("matches reference canyon and dawn advisory", func(t *testing.T) {
		report := EvaluateRouteRisk(syntheticZones, 6) // Dawn hour

		expectedNotice := "Caution: 3 high-risk wildlife collision zones identified along US-191. Dawn travel increases animal crossing activity by 2.1x. Reduce speed to 45 mph through canyon sections."
		if report.AdvisoryNotice != expectedNotice {
			t.Errorf("\nExpected: %q\nGot:      %q", expectedNotice, report.AdvisoryNotice)
		}
	})

	t.Run("multi-highway corridor notice", func(t *testing.T) {
		multiHwy := []HazardZone{
			{SegmentID: "z1", HighwayRef: "US-191", RiskScore: 0.7},
			{SegmentID: "z2", HighwayRef: "US-89", RiskScore: 0.75},
		}
		report := EvaluateRouteRisk(multiHwy, 20) // Dusk hour (2.8x)

		if !strings.Contains(report.AdvisoryNotice, "along US-191 and US-89") {
			t.Errorf("expected combined highway refs, got: %s", report.AdvisoryNotice)
		}
		if !strings.Contains(report.AdvisoryNotice, "Dusk travel increases animal crossing activity by 2.8x.") {
			t.Errorf("expected dusk 2.8x multiplier notice, got: %s", report.AdvisoryNotice)
		}
	})

	t.Run("multi-highway corridor notice with 3+ highways", func(t *testing.T) {
		threeHwys := []HazardZone{
			{SegmentID: "z1", HighwayRef: "US-191", RiskScore: 0.7},
			{SegmentID: "z2", HighwayRef: "US-89", RiskScore: 0.75},
			{SegmentID: "z3", HighwayRef: "WY-22", RiskScore: 0.65},
		}
		report := EvaluateRouteRisk(threeHwys, 12)

		if !strings.Contains(report.AdvisoryNotice, "along US-191, US-89, and WY-22") {
			t.Errorf("expected 3 combined highway refs with Oxford comma, got: %s", report.AdvisoryNotice)
		}
	})

	t.Run("canopy advice in moderate advisory", func(t *testing.T) {
		canopyZone := []HazardZone{
			{SegmentID: "z1", HighwayRef: "WY-22", RiskScore: 0.42, PrimaryRisk: "Dense Canopy & Low Visibility"},
		}
		report := EvaluateRouteRisk(canopyZone, 12) // Day hour (1.0x), Moderate

		if !strings.Contains(report.AdvisoryNotice, "Reduce speed through dense canopy sections") {
			t.Errorf("expected canopy recommendation, got: %s", report.AdvisoryNotice)
		}
	})
}

// TestClassifyHazardTier tests individual hazard tier score boundaries.
func TestClassifyHazardTier(t *testing.T) {
	cases := []struct {
		score float64
		tier  string
	}{
		{0.95, "Critical"},
		{0.80, "Critical"},
		{0.75, "High"},
		{0.60, "High"},
		{0.55, "Moderate"},
		{0.40, "Moderate"},
		{0.39, "Low"},
		{0.00, "Low"},
		{-0.1, "Low"},
	}

	for _, c := range cases {
		got := ClassifyHazardTier(c.score)
		if got != c.tier {
			t.Errorf("score %.2f: expected %s, got %s", c.score, c.tier, got)
		}
	}
}

// TestEvaluateRouteRisk_NegativeClampingAndTieBreaking verifies robustness against
// negative risk inputs and verifies deterministic sorting order for tied scores.
func TestEvaluateRouteRisk_NegativeClampingAndTieBreaking(t *testing.T) {
	segs := []HazardZone{
		{SegmentID: "b-seg", HighwayRef: "US-191", Milepost: 20.0, RiskScore: 0.8},
		{SegmentID: "a-seg", HighwayRef: "US-191", Milepost: 10.0, RiskScore: 0.8},
		{SegmentID: "neg-seg", HighwayRef: "US-191", Milepost: 5.0, RiskScore: -0.5},
	}

	report := EvaluateRouteRisk(segs, 12)

	if len(report.HazardZones) != 2 {
		t.Fatalf("expected 2 hazard zones (negative clamped to 0), got %d", len(report.HazardZones))
	}
	// Tie broken by Milepost ascending
	if report.HazardZones[0].SegmentID != "a-seg" || report.HazardZones[1].SegmentID != "b-seg" {
		t.Errorf("expected tie broken by milepost: got %s, %s",
			report.HazardZones[0].SegmentID, report.HazardZones[1].SegmentID)
	}
}

// TestWaypoint_DistanceCalculations verifies pure Go geodesic distance utilities.
func TestWaypoint_DistanceCalculations(t *testing.T) {
	// Jackson, WY (43.4799, -110.7624) to Moran, WY (43.8416, -110.5074) ~45.3 km
	jackson := Waypoint{Latitude: 43.4799, Longitude: -110.7624}
	moran := Waypoint{Latitude: 43.8416, Longitude: -110.5074}

	dist := HaversineDistanceKm(jackson, moran)
	if dist < 42.0 || dist > 48.0 {
		t.Errorf("expected Jackson to Moran distance ~45 km, got %.2f km", dist)
	}

	// Zero distance between identical waypoints
	if HaversineDistanceKm(jackson, jackson) != 0.0 {
		t.Errorf("expected 0 km for identical points")
	}

	// Polyline distance with multiple points
	polyDist := CalculatePolylineDistanceKm([]Waypoint{jackson, moran, jackson})
	if math.Abs(polyDist-dist*2) > 1e-6 {
		t.Errorf("expected polyline round-trip distance %.2f, got %.2f", dist*2, polyDist)
	}

	// Edge case: single waypoint polyline
	if CalculatePolylineDistanceKm([]Waypoint{jackson}) != 0.0 {
		t.Errorf("expected 0 km for single point polyline")
	}

	// Edge case: empty waypoints polyline
	if CalculatePolylineDistanceKm(nil) != 0.0 {
		t.Errorf("expected 0 km for nil waypoints polyline")
	}
}
