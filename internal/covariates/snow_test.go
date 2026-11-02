package covariates_test

import (
	"testing"

	"corridor/internal/covariates"
)

func TestMountainRegionDetection(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected covariates.MountainRegion
	}{
		{
			name:     "Tioga Pass (Yosemite, Sierra Nevada)",
			lat:      37.91,
			lon:      -119.26,
			expected: covariates.RegionSierraNevada,
		},
		{
			name:     "Yosemite Valley (Sierra Nevada)",
			lat:      37.74,
			lon:      -119.58,
			expected: covariates.RegionSierraNevada,
		},
		{
			name:     "Trail Ridge Road (Rocky Mountain, Central Rockies)",
			lat:      40.44,
			lon:      -105.75,
			expected: covariates.RegionCentralRockies,
		},
		{
			name:     "Logan Pass (Glacier, Northern Rockies)",
			lat:      48.69,
			lon:      -113.72,
			expected: covariates.RegionNorthernRockies,
		},
		{
			name:     "Jackson Hole (Grand Teton, Northern Rockies)",
			lat:      43.48,
			lon:      -110.76,
			expected: covariates.RegionNorthernRockies,
		},
		{
			name:     "Miami Florida (Outside Western US)",
			lat:      25.76,
			lon:      -80.19,
			expected: covariates.RegionOutside,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := covariates.GetMountainRegion(tc.lat, tc.lon)
			if actual != tc.expected {
				t.Errorf("GetMountainRegion(%f, %f) = %v; want %v", tc.lat, tc.lon, actual, tc.expected)
			}
		})
	}
}

func TestEstimateSnowDepth(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lon            float64
		elevationM     float64
		month          int
		minExpectedCm  float64
		maxExpectedCm  float64
		expectedZeroCm bool
	}{
		{
			name:          "Tioga Pass (Yosemite 3031m) January Snowpack",
			lat:           37.91,
			lon:           -119.26,
			elevationM:    3031.0,
			month:         1,
			minExpectedCm: 180.0,
			maxExpectedCm: 260.0,
		},
		{
			name:          "Tioga Pass (Yosemite 3031m) March Peak Snowpack",
			lat:           37.91,
			lon:           -119.26,
			elevationM:    3031.0,
			month:         3,
			minExpectedCm: 230.0,
			maxExpectedCm: 320.0,
		},
		{
			name:           "Tioga Pass (Yosemite 3031m) July Summer Meltout",
			lat:            37.91,
			lon:            -119.26,
			elevationM:     3031.0,
			month:          7,
			expectedZeroCm: true,
		},
		{
			name:          "Trail Ridge Road (Rocky Mtn 3713m) January Snowpack",
			lat:           40.44,
			lon:           -105.75,
			elevationM:    3713.0,
			month:         1,
			minExpectedCm: 200.0,
			maxExpectedCm: 300.0,
		},
		{
			name:          "Trail Ridge Road (Rocky Mtn 3713m) March Peak Snowpack",
			lat:           40.44,
			lon:           -105.75,
			elevationM:    3713.0,
			month:         3,
			minExpectedCm: 250.0,
			maxExpectedCm: 350.0,
		},
		{
			name:           "Trail Ridge Road (Rocky Mtn 3713m) August Summer",
			lat:            40.44,
			lon:            -105.75,
			elevationM:     3713.0,
			month:          8,
			expectedZeroCm: true,
		},
		{
			name:          "Logan Pass (Glacier 2026m) January Snowpack",
			lat:           48.69,
			lon:           -113.72,
			elevationM:    2026.0,
			month:         1,
			minExpectedCm: 120.0,
			maxExpectedCm: 200.0,
		},
		{
			name:          "Logan Pass (Glacier 2026m) March Peak Snowpack",
			lat:           48.69,
			lon:           -113.72,
			elevationM:    2026.0,
			month:         3,
			minExpectedCm: 160.0,
			maxExpectedCm: 250.0,
		},
		{
			name:           "Logan Pass (Glacier 2026m) July Summer",
			lat:            48.69,
			lon:            -113.72,
			elevationM:     2026.0,
			month:          7,
			expectedZeroCm: true,
		},
		{
			name:           "Yosemite Valley Floor (1200m) January (below snowline)",
			lat:            37.74,
			lon:            -119.58,
			elevationM:     1200.0,
			month:          1,
			expectedZeroCm: true,
		},
		{
			name:           "Yosemite Valley Floor (1200m) July Summer",
			lat:            37.74,
			lon:            -119.58,
			elevationM:     1200.0,
			month:          7,
			expectedZeroCm: true,
		},
		{
			name:          "Jackson Hole Valley Floor (1900m) January Snow",
			lat:           43.48,
			lon:           -110.76,
			elevationM:    1900.0,
			month:         1,
			minExpectedCm: 40.0,
			maxExpectedCm: 90.0,
		},
		{
			name:           "Jackson Hole Valley Floor (1900m) July Summer",
			lat:            43.48,
			lon:            -110.76,
			elevationM:     1900.0,
			month:          7,
			expectedZeroCm: true,
		},
		{
			name:           "Out of bounds month (month 13)",
			lat:            37.91,
			lon:            -119.26,
			elevationM:     3031.0,
			month:          13,
			expectedZeroCm: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			depth := covariates.EstimateSnowDepth(tc.lat, tc.lon, tc.elevationM, tc.month)
			if tc.expectedZeroCm {
				if depth != 0.0 {
					t.Errorf("EstimateSnowDepth() = %f; want 0.0", depth)
				}
			} else {
				if depth < tc.minExpectedCm || depth > tc.maxExpectedCm {
					t.Errorf("EstimateSnowDepth() = %f; want within [%f, %f]", depth, tc.minExpectedCm, tc.maxExpectedCm)
				}
			}
		})
	}
}

func TestIsPassSeasonallyClosed(t *testing.T) {
	tests := []struct {
		name       string
		elevationM float64
		month      int
		wantClosed bool
	}{
		// Tioga Pass (3031m, Yosemite)
		{"Tioga Pass (3031m) January (Winter)", 3031.0, 1, true},
		{"Tioga Pass (3031m) February (Winter)", 3031.0, 2, true},
		{"Tioga Pass (3031m) March (Winter)", 3031.0, 3, true},
		{"Tioga Pass (3031m) April (Winter)", 3031.0, 4, true},
		{"Tioga Pass (3031m) May (Winter)", 3031.0, 5, true},
		{"Tioga Pass (3031m) June (Shoulder)", 3031.0, 6, false},
		{"Tioga Pass (3031m) July (Summer Open)", 3031.0, 7, false},
		{"Tioga Pass (3031m) August (Summer Open)", 3031.0, 8, false},
		{"Tioga Pass (3031m) September (Summer Open)", 3031.0, 9, false},
		{"Tioga Pass (3031m) October (Shoulder)", 3031.0, 10, false},
		{"Tioga Pass (3031m) November (Winter)", 3031.0, 11, true},
		{"Tioga Pass (3031m) December (Winter)", 3031.0, 12, true},

		// Trail Ridge Road (3713m, Rocky Mountain)
		{"Trail Ridge Road (3713m) January (Winter)", 3713.0, 1, true},
		{"Trail Ridge Road (3713m) July (Summer Open)", 3713.0, 7, false},

		// Logan Pass (2026m, Glacier)
		{"Logan Pass (2026m) January (Winter)", 2026.0, 1, true},
		{"Logan Pass (2026m) December (Winter)", 2026.0, 12, true},
		{"Logan Pass (2026m) July (Summer Open)", 2026.0, 7, false},

		// Valley floor roads (Must remain open year-round)
		{"Yosemite Valley Floor (1200m) January", 1200.0, 1, false},
		{"Yosemite Valley Floor (1200m) July", 1200.0, 7, false},
		{"Jackson Hole Valley Floor (1900m) January", 1900.0, 1, false},
		{"Jackson Hole Valley Floor (1900m) July", 1900.0, 7, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := covariates.IsPassSeasonallyClosed(tc.elevationM, tc.month)
			if got != tc.wantClosed {
				t.Errorf("IsPassSeasonallyClosed(%f, %d) = %v; want %v", tc.elevationM, tc.month, got, tc.wantClosed)
			}
		})
	}
}

func TestIsPassSeasonallyClosedAt(t *testing.T) {
	tests := []struct {
		name       string
		lat        float64
		lon        float64
		elevationM float64
		month      int
		wantClosed bool
	}{
		{"Tioga Pass (Yosemite 37.91N, 3031m) Jan", 37.91, -119.26, 3031.0, 1, true},
		{"Trail Ridge Road (Rocky Mtn 40.44N, 3713m) Jan", 40.44, -105.75, 3713.0, 1, true},
		{"Logan Pass (Glacier 48.69N, 2026m) Jan", 48.69, -113.72, 2026.0, 1, true},
		{"Yosemite Valley Floor (37.74N, 1200m) Jan", 37.74, -119.58, 1200.0, 1, false},
		{"Jackson Hole Valley Floor (43.48N, 1900m) Jan", 43.48, -110.76, 1900.0, 1, false},
		{"Tioga Pass (Yosemite 37.91N, 3031m) Jul", 37.91, -119.26, 3031.0, 7, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := covariates.IsPassSeasonallyClosedAt(tc.lat, tc.lon, tc.elevationM, tc.month)
			if got != tc.wantClosed {
				t.Errorf("IsPassSeasonallyClosedAt(%f, %f, %f, %d) = %v; want %v",
					tc.lat, tc.lon, tc.elevationM, tc.month, got, tc.wantClosed)
			}
		})
	}
}

func TestAdjustTrafficForSnow(t *testing.T) {
	tests := []struct {
		name       string
		baseAADT   float64
		elevationM float64
		month      int
		expected   float64
	}{
		{
			name:       "Tioga Pass closed in January (Traffic scaled to zero)",
			baseAADT:   3500.0,
			elevationM: 3031.0,
			month:      1,
			expected:   0.0,
		},
		{
			name:       "Tioga Pass June shoulder month (50% reduction)",
			baseAADT:   3500.0,
			elevationM: 3031.0,
			month:      6,
			expected:   1750.0,
		},
		{
			name:       "Tioga Pass October shoulder month (40% reduction)",
			baseAADT:   3500.0,
			elevationM: 3031.0,
			month:      10,
			expected:   2100.0,
		},
		{
			name:       "Tioga Pass July summer peak (100% traffic)",
			baseAADT:   3500.0,
			elevationM: 3031.0,
			month:      7,
			expected:   3500.0,
		},
		{
			name:       "Trail Ridge Road closed in December",
			baseAADT:   4500.0,
			elevationM: 3713.0,
			month:      12,
			expected:   0.0,
		},
		{
			name:       "Logan Pass closed in February",
			baseAADT:   2800.0,
			elevationM: 2026.0,
			month:      2,
			expected:   0.0,
		},
		{
			name:       "Logan Pass open in August",
			baseAADT:   2800.0,
			elevationM: 2026.0,
			month:      8,
			expected:   2800.0,
		},
		{
			name:       "Yosemite Valley floor road open in January",
			baseAADT:   8000.0,
			elevationM: 1200.0,
			month:      1,
			expected:   8000.0,
		},
		{
			name:       "Jackson Hole valley floor road open in January",
			baseAADT:   12000.0,
			elevationM: 1900.0,
			month:      1,
			expected:   12000.0,
		},
		{
			name:       "Zero baseline AADT remains zero",
			baseAADT:   0.0,
			elevationM: 1200.0,
			month:      1,
			expected:   0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := covariates.AdjustTrafficForSnow(tc.baseAADT, tc.elevationM, tc.month)
			if got != tc.expected {
				t.Errorf("AdjustTrafficForSnow(%f, %f, %d) = %f; want %f",
					tc.baseAADT, tc.elevationM, tc.month, got, tc.expected)
			}
		})
	}
}
