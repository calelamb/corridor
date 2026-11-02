package ingest

import (
	"strings"
	"testing"
)

func TestWSDOTSafetyArtifact(t *testing.T) {
	art := WSDOTSafetyArtifact()
	if art.SourceID == "" {
		t.Error("expected non-empty SourceID")
	}
	if !strings.Contains(art.Name, "WSDOT") {
		t.Errorf("expected WSDOT in name, got %q", art.Name)
	}
	if len(art.SHA256) != 64 {
		t.Errorf("expected 64-char sha256, got %d", len(art.SHA256))
	}
	if !strings.Contains(art.URL, "FeatureServer/3") {
		t.Errorf("expected FeatureServer/3 in URL, got %q", art.URL)
	}
}

func TestWSDOTRouteNormalization(t *testing.T) {
	tests := []struct {
		name      string
		rawRoute  string
		region    string
		beginARM  float64
		endARM    float64
		wantRoute string
	}{
		{
			name:      "Olympic US-101 via region",
			rawRoute:  "101",
			region:    "Olympic",
			beginARM:  200.0,
			endARM:    205.0,
			wantRoute: "US-101 Olympic NP",
		},
		{
			name:      "Olympic US-101 via ARM range",
			rawRoute:  "US-101",
			region:    "",
			beginARM:  230.5,
			endARM:    235.0,
			wantRoute: "US-101 Olympic NP",
		},
		{
			name:      "Olympic US-101 explicitly named",
			rawRoute:  "US-101 Olympic NP",
			region:    "",
			beginARM:  240.0,
			endARM:    245.0,
			wantRoute: "US-101 Olympic NP",
		},
		{
			name:      "I-90 Snoqualmie Pass via ARM",
			rawRoute:  "090",
			region:    "South Central",
			beginARM:  52.0,
			endARM:    62.0,
			wantRoute: "I-90 Snoqualmie Pass",
		},
		{
			name:      "I-90 Snoqualmie Pass explicitly named",
			rawRoute:  "I-90 Snoqualmie Pass",
			region:    "",
			beginARM:  54.0,
			endARM:    60.0,
			wantRoute: "I-90 Snoqualmie Pass",
		},
		{
			name:      "I-90 outside Snoqualmie Pass",
			rawRoute:  "090",
			region:    "Eastern",
			beginARM:  250.0,
			endARM:    260.0,
			wantRoute: "I-90",
		},
		{
			name:      "US-97 Highway",
			rawRoute:  "097",
			region:    "North Central",
			beginARM:  120.0,
			endARM:    125.0,
			wantRoute: "US-97",
		},
		{
			name:      "US-97 formatted directly",
			rawRoute:  "US-97",
			region:    "South Central",
			beginARM:  10.0,
			endARM:    15.0,
			wantRoute: "US-97",
		},
		{
			name:      "US-2 Stevens Pass",
			rawRoute:  "002",
			region:    "North Central",
			beginARM:  60.0,
			endARM:    70.0,
			wantRoute: "US-2",
		},
		{
			name:      "I-5 Interstate Corridor",
			rawRoute:  "005",
			region:    "Northwest",
			beginARM:  100.0,
			endARM:    105.0,
			wantRoute: "I-5",
		},
		{
			name:      "SR-20 State Route",
			rawRoute:  "020",
			region:    "North Central",
			beginARM:  130.0,
			endARM:    140.0,
			wantRoute: "SR-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeWSDOTRoute(tt.rawRoute, tt.region, tt.beginARM, tt.endARM)
			if got != tt.wantRoute {
				t.Errorf("NormalizeWSDOTRoute(%q, %q, %v, %v) = %q, want %q",
					tt.rawRoute, tt.region, tt.beginARM, tt.endARM, got, tt.wantRoute)
			}
		})
	}
}

func TestWSDOTPriorityNormalization(t *testing.T) {
	tests := []struct {
		raw  string
		want WSDOTPriorityTier
	}{
		{"High", WSDOTPriorityHigh},
		{"HIGH", WSDOTPriorityHigh},
		{"Tier 1", WSDOTPriorityHigh},
		{"h", WSDOTPriorityHigh},
		{"Medium", WSDOTPriorityMedium},
		{"Moderate", WSDOTPriorityMedium},
		{"Tier 2", WSDOTPriorityMedium},
		{"med", WSDOTPriorityMedium},
		{"Low", WSDOTPriorityLow},
		{"LOW", WSDOTPriorityLow},
		{"Tier 3", WSDOTPriorityLow},
		{"None", WSDOTPriorityUnranked},
		{"", WSDOTPriorityUnranked},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got := NormalizeWSDOTPriority(tt.raw)
			if got != tt.want {
				t.Errorf("NormalizeWSDOTPriority(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestWSDOTTargetSpeciesExtraction(t *testing.T) {
	tests := []struct {
		name        string
		deerCount   int
		elkCount    int
		mooseCount  int
		deerCrash   int
		elkCrash    int
		deerHS      int
		elkHS       int
		explicit    []string
		wantSpecies []string
	}{
		{
			name:        "Both Elk and Mule Deer present",
			deerCount:   45,
			elkCount:    12,
			wantSpecies: []string{"Elk", "Mule Deer"},
		},
		{
			name:        "Elk only from crash count",
			elkCrash:    8,
			wantSpecies: []string{"Elk"},
		},
		{
			name:        "Mule Deer only from hotspot",
			deerHS:      1,
			wantSpecies: []string{"Mule Deer"},
		},
		{
			name:        "Moose with Mule Deer",
			deerCount:   10,
			mooseCount:  5,
			wantSpecies: []string{"Mule Deer", "Moose"},
		},
		{
			name:        "Explicit species list provided",
			explicit:    []string{"Elk", "Mule Deer", "Bighorn Sheep"},
			wantSpecies: []string{"Elk", "Mule Deer", "Bighorn Sheep"},
		},
		{
			name:        "No species counts recorded",
			wantSpecies: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractWSDOTTargetSpecies(
				tt.deerCount, tt.elkCount, tt.mooseCount,
				tt.deerCrash, tt.elkCrash, tt.deerHS, tt.elkHS,
				tt.explicit,
			)

			if len(got) != len(tt.wantSpecies) {
				t.Fatalf("ExtractWSDOTTargetSpecies len mismatch: got %v, want %v", got, tt.wantSpecies)
			}
			for i := range got {
				if got[i] != tt.wantSpecies[i] {
					t.Errorf("target species [%d] = %q, want %q", i, got[i], tt.wantSpecies[i])
				}
			}
		})
	}
}

func TestWSDOTParseGeoJSON(t *testing.T) {
	geojsonInput := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"id": 1,
				"properties": {
					"OBJECTID": 101,
					"RouteID": "101",
					"WSDOTRegion": "Olympic",
					"BARM": 240.0,
					"EARM": 245.5,
					"BSRMP": 240.0,
					"ESRMP": 245.5,
					"Safetyv10_Priority": "High",
					"WRSv10": 88.5,
					"DeerCount": 54,
					"ElkCount": 22,
					"DeerCrashCount": 30,
					"ElkCrashCount": 14,
					"AADT_1": 8500
				},
				"geometry": {
					"type": "LineString",
					"coordinates": [[-123.6, 48.0], [-123.5, 48.1]]
				}
			},
			{
				"type": "Feature",
				"id": 2,
				"properties": {
					"OBJECTID": 102,
					"RouteID": "090",
					"WSDOTRegion": "South Central",
					"BARM": 54.0,
					"EARM": 62.0,
					"Safetyv10_Priority": "High",
					"WRSv10": 94.2,
					"DeerCount": 110,
					"ElkCount": 45,
					"DeerCrashCount": 50,
					"ElkCrashCount": 25,
					"AADT_1": 28000
				},
				"geometry": {
					"type": "LineString",
					"coordinates": [[-121.4, 47.3], [-121.3, 47.2]]
				}
			},
			{
				"type": "Feature",
				"id": 3,
				"properties": {
					"OBJECTID": 103,
					"RouteID": "097",
					"WSDOTRegion": "North Central",
					"BARM": 140.0,
					"EARM": 145.0,
					"Safetyv10_Priority": "Medium",
					"WRSv10": 45.0,
					"DeerCount": 78,
					"ElkCount": 0,
					"DeerCrashCount": 35,
					"ElkCrashCount": 0,
					"AADT_1": 4200
				}
			}
		]
	}`

	rankings, err := ParseWSDOTSafetyRankings(strings.NewReader(geojsonInput))
	if err != nil {
		t.Fatalf("ParseWSDOTSafetyRankings failed: %v", err)
	}

	if len(rankings) != 3 {
		t.Fatalf("expected 3 rankings, got %d", len(rankings))
	}

	// 1. Olympic US-101 check
	r1 := rankings[0]
	if r1.Route != "US-101 Olympic NP" {
		t.Errorf("expected 'US-101 Olympic NP', got %q", r1.Route)
	}
	if r1.BeginARM != 240.0 || r1.EndARM != 245.5 {
		t.Errorf("expected ARM 240.0-245.5, got %v-%v", r1.BeginARM, r1.EndARM)
	}
	if r1.LengthMiles != 5.5 {
		t.Errorf("expected length 5.5, got %v", r1.LengthMiles)
	}
	if r1.PriorityTier != WSDOTPriorityHigh {
		t.Errorf("expected High priority tier, got %q", r1.PriorityTier)
	}
	if r1.CollisionRate != 88.5 {
		t.Errorf("expected collision rate 88.5, got %v", r1.CollisionRate)
	}
	if len(r1.TargetSpecies) != 2 || r1.TargetSpecies[0] != "Elk" || r1.TargetSpecies[1] != "Mule Deer" {
		t.Errorf("expected target species [Elk, Mule Deer], got %v", r1.TargetSpecies)
	}
	if r1.GeometryJSON == nil {
		t.Error("expected non-nil geometry JSON")
	}

	// 2. I-90 Snoqualmie Pass check
	r2 := rankings[1]
	if r2.Route != "I-90 Snoqualmie Pass" {
		t.Errorf("expected 'I-90 Snoqualmie Pass', got %q", r2.Route)
	}
	if r2.BeginARM != 54.0 || r2.EndARM != 62.0 {
		t.Errorf("expected ARM 54.0-62.0, got %v-%v", r2.BeginARM, r2.EndARM)
	}
	if r2.LengthMiles != 8.0 {
		t.Errorf("expected length 8.0, got %v", r2.LengthMiles)
	}
	if r2.PriorityTier != WSDOTPriorityHigh {
		t.Errorf("expected High priority, got %q", r2.PriorityTier)
	}
	if r2.AADT != 28000 {
		t.Errorf("expected AADT 28000, got %d", r2.AADT)
	}

	// 3. US-97 check
	r3 := rankings[2]
	if r3.Route != "US-97" {
		t.Errorf("expected 'US-97', got %q", r3.Route)
	}
	if r3.PriorityTier != WSDOTPriorityMedium {
		t.Errorf("expected Medium priority, got %q", r3.PriorityTier)
	}
	if len(r3.TargetSpecies) != 1 || r3.TargetSpecies[0] != "Mule Deer" {
		t.Errorf("expected target species [Mule Deer], got %v", r3.TargetSpecies)
	}
}

func TestWSDOTParseArcGISFeatureSet(t *testing.T) {
	arcgisInput := `{
		"objectIdFieldName": "OBJECTID",
		"features": [
			{
				"attributes": {
					"OBJECTID": 501,
					"RouteID": "101",
					"WSDOTRegion": "Olympic",
					"BARM": 245.0,
					"EARM": 240.0,
					"Safetyv10_Priority": "High",
					"WRSv10": 79.1,
					"DeerCount": 60,
					"ElkCount": 30,
					"DeerCrashCount": 25,
					"ElkCrashCount": 15
				},
				"geometry": {
					"paths": [[[-123.5, 48.0], [-123.4, 48.05]]]
				}
			},
			{
				"attributes": {
					"OBJECTID": 502,
					"RouteID": "097",
					"WSDOTRegion": "South Central",
					"BARM": 15.0,
					"EARM": 20.0,
					"Safetyv10_Priority": "Low",
					"WRSv10": 12.0,
					"DeerCount": 10,
					"ElkCount": 0
				}
			}
		]
	}`

	rankings, err := ParseWSDOTSafetyRankings(strings.NewReader(arcgisInput))
	if err != nil {
		t.Fatalf("ParseWSDOTSafetyRankings failed: %v", err)
	}

	if len(rankings) != 2 {
		t.Fatalf("expected 2 rankings, got %d", len(rankings))
	}

	// Verify swapping reversed BARM/EARM
	if rankings[0].BeginARM != 240.0 || rankings[0].EndARM != 245.0 {
		t.Errorf("expected normalized begin/end ARM 240-245, got %v-%v",
			rankings[0].BeginARM, rankings[0].EndARM)
	}
	if rankings[0].Route != "US-101 Olympic NP" {
		t.Errorf("expected 'US-101 Olympic NP', got %q", rankings[0].Route)
	}
	if rankings[1].Route != "US-97" {
		t.Errorf("expected 'US-97', got %q", rankings[1].Route)
	}
	if rankings[1].PriorityTier != WSDOTPriorityLow {
		t.Errorf("expected Low priority tier, got %q", rankings[1].PriorityTier)
	}
}

func TestWSDOTFilters(t *testing.T) {
	rankings := []WSDOTSafetyRanking{
		{
			Route:         "US-101 Olympic NP",
			PriorityTier:  WSDOTPriorityHigh,
			TargetSpecies: []string{"Elk", "Mule Deer"},
		},
		{
			Route:         "I-90 Snoqualmie Pass",
			PriorityTier:  WSDOTPriorityHigh,
			TargetSpecies: []string{"Elk", "Mule Deer"},
		},
		{
			Route:         "US-97",
			PriorityTier:  WSDOTPriorityMedium,
			TargetSpecies: []string{"Mule Deer"},
		},
		{
			Route:         "SR-20",
			PriorityTier:  WSDOTPriorityLow,
			TargetSpecies: []string{"Mule Deer"},
		},
	}

	// 1. Filter by Priority
	high := FilterWSDOTByPriority(rankings, WSDOTPriorityHigh)
	if len(high) != 2 {
		t.Errorf("expected 2 High priority rankings, got %d", len(high))
	}

	// 2. Filter by Route
	olympic := FilterWSDOTByRoute(rankings, "Olympic")
	if len(olympic) != 1 || olympic[0].Route != "US-101 Olympic NP" {
		t.Errorf("expected 1 Olympic route, got %v", olympic)
	}
	i90 := FilterWSDOTByRoute(rankings, "90")
	if len(i90) != 1 || i90[0].Route != "I-90 Snoqualmie Pass" {
		t.Errorf("expected 1 I-90 route, got %v", i90)
	}

	// 3. Filter by Species
	elk := FilterWSDOTBySpecies(rankings, "Elk")
	if len(elk) != 2 {
		t.Errorf("expected 2 Elk rankings, got %d", len(elk))
	}
	deer := FilterWSDOTBySpecies(rankings, "Mule Deer")
	if len(deer) != 4 {
		t.Errorf("expected 4 Mule Deer rankings, got %d", len(deer))
	}
}

func TestWSDOTParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nil reader", ""},
		{"empty json", ""},
		{"invalid json", "{bad:json}"},
		{"empty feature collection", `{"type":"FeatureCollection","features":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil reader" {
				_, err := ParseWSDOTSafetyRankings(nil)
				if err == nil {
					t.Error("expected error for nil reader")
				}
				return
			}
			_, err := ParseWSDOTSafetyRankings(strings.NewReader(tt.input))
			if err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}
