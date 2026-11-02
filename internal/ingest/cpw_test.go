package ingest

import (
	"strings"
	"testing"
)

func TestCPWCrossingsArtifact(t *testing.T) {
	art := CPWCrossingsArtifact()
	if art.SourceID == "" {
		t.Error("expected non-empty SourceID")
	}
	if !strings.Contains(art.Name, "CPW") {
		t.Errorf("expected CPW in name, got %q", art.Name)
	}
	if len(art.SHA256) != 64 {
		t.Errorf("expected 64-char sha256, got %d", len(art.SHA256))
	}
	if !strings.Contains(art.URL, "FeatureServer/0") {
		t.Errorf("expected FeatureServer/0 in URL, got %q", art.URL)
	}
}

func TestCPWStructureTypeNormalization(t *testing.T) {
	tests := []struct {
		raw      string
		wantType CPWStructureType
	}{
		{"Wildlife Overpass", CPWStructureOverpass},
		{"Overpass", CPWStructureOverpass},
		{"Land Bridge", CPWStructureOverpass},
		{"overcrossing", CPWStructureOverpass},
		{"Wildlife Underpass", CPWStructureUnderpass},
		{"Underpass", CPWStructureUnderpass},
		{"Bridge Underpass", CPWStructureUnderpass},
		{"Wildlife Span", CPWStructureUnderpass},
		{"Large Mammal Culvert", CPWStructureLargeMammalCulvert},
		{"Large Culvert", CPWStructureLargeMammalCulvert},
		{"Mammal Box Culvert", CPWStructureLargeMammalCulvert},
		{"Small Mammal Culvert", CPWStructureSmallMammalCulvert},
		{"Small Culvert", CPWStructureSmallMammalCulvert},
		{"Small Mammal Pipe", CPWStructureSmallMammalCulvert},
		{"Herptile Pipe", CPWStructureSmallMammalCulvert},
		{"Wildlife Jump-Out", CPWStructureWildlifeJumpOut},
		{"Jump-out", CPWStructureWildlifeJumpOut},
		{"Jump Out", CPWStructureWildlifeJumpOut},
		{"Escape Ramp", CPWStructureWildlifeJumpOut},
		{"8ft ungulate exclusion fence", CPWStructure8ftExclusionFence},
		{"8-Foot Wildlife Fence", CPWStructure8ftExclusionFence},
		{"8ft Wildlife Fencing", CPWStructure8ftExclusionFence},
		{"Wildlife Exclusion Fence", CPWStructure8ftExclusionFence},
		{"Exclusion Fence", CPWStructure8ftExclusionFence},
		{"Fencing", CPWStructure8ftExclusionFence},
		{"Unknown Structure", CPWStructureOther},
		{"", CPWStructureOther},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got := NormalizeCPWStructureType(tt.raw)
			if got != tt.wantType {
				t.Errorf("NormalizeCPWStructureType(%q) = %q, want %q", tt.raw, got, tt.wantType)
			}
		})
	}
}

func TestCPWRouteNormalization(t *testing.T) {
	tests := []struct {
		name         string
		rawRoute     string
		corridorDesc string
		lon          float64
		lat          float64
		wantRoute    string
	}{
		{
			name:         "US-34 Rocky Mountain via description",
			rawRoute:     "US-34",
			corridorDesc: "Rocky Mountain National Park East Entrance",
			lon:          -105.5,
			lat:          40.38,
			wantRoute:    "US-34 Rocky Mountain",
		},
		{
			name:         "US-34 Rocky Mountain via coordinates",
			rawRoute:     "34",
			corridorDesc: "Fall River Road Corridor",
			lon:          -105.55,
			lat:          40.40,
			wantRoute:    "US-34 Rocky Mountain",
		},
		{
			name:         "US-34 Rocky Mountain explicitly provided",
			rawRoute:     "US-34 Rocky Mountain",
			corridorDesc: "",
			lon:          0,
			lat:          0,
			wantRoute:    "US-34 Rocky Mountain",
		},
		{
			name:         "US-34 Eastern Plains outside Rocky Mountain",
			rawRoute:     "34",
			corridorDesc: "Yuma County",
			lon:          -102.7,
			lat:          40.1,
			wantRoute:    "US-34",
		},
		{
			name:         "I-70 Vail Pass",
			rawRoute:     "I-70",
			corridorDesc: "Vail Pass Wildlife Overpass Project",
			lon:          -106.2,
			lat:          39.5,
			wantRoute:    "I-70",
		},
		{
			name:         "I-70 raw number",
			rawRoute:     "70",
			corridorDesc: "Floyd Hill Mitigation",
			lon:          -105.4,
			lat:          39.7,
			wantRoute:    "I-70",
		},
		{
			name:         "US-550 Billy Creek",
			rawRoute:     "US-550",
			corridorDesc: "Billy Creek Wildlife Area",
			lon:          -107.7,
			lat:          38.3,
			wantRoute:    "US-550",
		},
		{
			name:         "US-550 raw number",
			rawRoute:     "550",
			corridorDesc: "Durango to NM Line",
			lon:          -107.8,
			lat:          37.2,
			wantRoute:    "US-550",
		},
		{
			name:         "SH-9 Green Mountain Reservoir",
			rawRoute:     "SH-9",
			corridorDesc: "Green Mountain Reservoir Crossing Project",
			lon:          -106.3,
			lat:          39.9,
			wantRoute:    "SH-9",
		},
		{
			name:         "US-160 Pagosa Springs",
			rawRoute:     "160",
			corridorDesc: "Yellow Jacket Pass",
			lon:          -107.3,
			lat:          37.2,
			wantRoute:    "US-160",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeCPWRoute(tt.rawRoute, tt.corridorDesc, tt.lon, tt.lat)
			if got != tt.wantRoute {
				t.Errorf("NormalizeCPWRoute(%q, %q, %v, %v) = %q, want %q",
					tt.rawRoute, tt.corridorDesc, tt.lon, tt.lat, got, tt.wantRoute)
			}
		})
	}
}

func TestCPWEffectivenessTier(t *testing.T) {
	tests := []struct {
		name          string
		explicitTier  string
		structureType CPWStructureType
		wantTier      CPWEffectivenessTier
	}{
		{
			name:          "Overpass default tier",
			explicitTier:  "",
			structureType: CPWStructureOverpass,
			wantTier:      CPWEffectivenessTierHigh,
		},
		{
			name:          "Underpass default tier",
			explicitTier:  "",
			structureType: CPWStructureUnderpass,
			wantTier:      CPWEffectivenessTierHigh,
		},
		{
			name:          "8ft Exclusion Fence default tier",
			explicitTier:  "",
			structureType: CPWStructure8ftExclusionFence,
			wantTier:      CPWEffectivenessTierHigh,
		},
		{
			name:          "Large Mammal Culvert default tier",
			explicitTier:  "",
			structureType: CPWStructureLargeMammalCulvert,
			wantTier:      CPWEffectivenessTierMedium,
		},
		{
			name:          "Small Mammal Culvert default tier",
			explicitTier:  "",
			structureType: CPWStructureSmallMammalCulvert,
			wantTier:      CPWEffectivenessTierLow,
		},
		{
			name:          "Wildlife Jump-Out default tier",
			explicitTier:  "",
			structureType: CPWStructureWildlifeJumpOut,
			wantTier:      CPWEffectivenessTierEscape,
		},
		{
			name:          "Explicit High Tier overrides default",
			explicitTier:  "High",
			structureType: CPWStructureLargeMammalCulvert,
			wantTier:      CPWEffectivenessTierHigh,
		},
		{
			name:          "Explicit Tier 2 overrides default",
			explicitTier:  "Tier 2",
			structureType: CPWStructureOverpass,
			wantTier:      CPWEffectivenessTierMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineCPWEffectivenessTier(tt.explicitTier, tt.structureType)
			if got != tt.wantTier {
				t.Errorf("DetermineCPWEffectivenessTier(%q, %q) = %q, want %q",
					tt.explicitTier, tt.structureType, got, tt.wantTier)
			}
		})
	}
}

func TestCPWConstructionYearParsing(t *testing.T) {
	tests := []struct {
		input any
		want  int
	}{
		{2016, 2016},
		{2021.0, 2021},
		{"2016", 2016},
		{"Built in 2018", 2018},
		{"2020-07-15", 2020},
		{"Installed Oct 2022", 2022},
		{"N/A", 0},
		{nil, 0},
	}

	for _, tt := range tests {
		got := ParseConstructionYear(tt.input)
		if got != tt.want {
			t.Errorf("ParseConstructionYear(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCPWParseGeoJSON(t *testing.T) {
	geojsonInput := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"id": "cpw-overpass-1",
				"geometry": {
					"type": "Point",
					"coordinates": [-105.52, 40.37]
				},
				"properties": {
					"OBJECTID": 101,
					"Structure_Type": "Wildlife Overpass",
					"Highway": "US-34",
					"Corridor": "Rocky Mountain National Park Corridor",
					"Construction_Year": 2020,
					"Effectiveness_Tier": "High",
					"Milepost": 64.2,
					"Target_Species": "Elk, Mule Deer, Bighorn Sheep"
				}
			},
			{
				"type": "Feature",
				"id": "cpw-underpass-2",
				"geometry": {
					"type": "Point",
					"coordinates": [-106.22, 39.52]
				},
				"properties": {
					"OBJECTID": 102,
					"Structure_Type": "Bridge Underpass",
					"Highway": "I-70",
					"Corridor": "Vail Pass East",
					"Construction_Year": 2021,
					"Milepost": 192.5
				}
			},
			{
				"type": "Feature",
				"id": "cpw-culvert-3",
				"geometry": {
					"type": "Point",
					"coordinates": [-107.71, 38.31]
				},
				"properties": {
					"OBJECTID": 103,
					"Structure_Type": "Large Mammal Culvert",
					"Highway": "US-550",
					"Corridor": "Billy Creek Wildlife Area",
					"Construction_Year": 2017,
					"Milepost": 115.0
				}
			},
			{
				"type": "Feature",
				"id": "cpw-small-culvert-4",
				"geometry": {
					"type": "Point",
					"coordinates": [-106.31, 39.92]
				},
				"properties": {
					"OBJECTID": 104,
					"Structure_Type": "Small Mammal Culvert",
					"Highway": "SH-9",
					"Corridor": "Green Mountain",
					"Construction_Year": 2016
				}
			},
			{
				"type": "Feature",
				"id": "cpw-jumpout-5",
				"geometry": {
					"type": "Point",
					"coordinates": [-106.32, 39.93]
				},
				"properties": {
					"OBJECTID": 105,
					"Structure_Type": "Wildlife Jump-Out",
					"Highway": "SH-9",
					"Construction_Year": 2016
				}
			},
			{
				"type": "Feature",
				"id": "cpw-fence-6",
				"geometry": {
					"type": "LineString",
					"coordinates": [[-106.33, 39.90], [-106.32, 39.95]]
				},
				"properties": {
					"OBJECTID": 106,
					"Structure_Type": "8ft ungulate exclusion fence",
					"Highway": "SH-9",
					"Construction_Year": 2016,
					"Fence_Height_Ft": 8.0,
					"Length_M": 16700.0
				}
			}
		]
	}`

	records, err := ParseCPWCrossings(strings.NewReader(geojsonInput))
	if err != nil {
		t.Fatalf("ParseCPWCrossings failed: %v", err)
	}

	if len(records) != 6 {
		t.Fatalf("expected 6 records, got %d", len(records))
	}

	// 1. US-34 Rocky Mountain Overpass
	r1 := records[0]
	if r1.StructureType != CPWStructureOverpass {
		t.Errorf("expected overpass structure type, got %q", r1.StructureType)
	}
	if r1.HighwayRoute != "US-34 Rocky Mountain" {
		t.Errorf("expected 'US-34 Rocky Mountain', got %q", r1.HighwayRoute)
	}
	if r1.Longitude != -105.52 || r1.Latitude != 40.37 {
		t.Errorf("expected coordinates [-105.52, 40.37], got [%v, %v]", r1.Longitude, r1.Latitude)
	}
	if r1.ConstructionYear != 2020 {
		t.Errorf("expected construction year 2020, got %d", r1.ConstructionYear)
	}
	if r1.EffectivenessTier != CPWEffectivenessTierHigh {
		t.Errorf("expected High effectiveness tier, got %q", r1.EffectivenessTier)
	}
	if len(r1.TargetSpecies) != 3 {
		t.Errorf("expected 3 target species, got %v", r1.TargetSpecies)
	}

	// 2. I-70 Underpass
	r2 := records[1]
	if r2.StructureType != CPWStructureUnderpass {
		t.Errorf("expected underpass, got %q", r2.StructureType)
	}
	if r2.HighwayRoute != "I-70" {
		t.Errorf("expected 'I-70', got %q", r2.HighwayRoute)
	}
	if r2.ConstructionYear != 2021 {
		t.Errorf("expected construction year 2021, got %d", r2.ConstructionYear)
	}
	if r2.EffectivenessTier != CPWEffectivenessTierHigh {
		t.Errorf("expected High effectiveness tier, got %q", r2.EffectivenessTier)
	}

	// 3. US-550 Large Mammal Culvert
	r3 := records[2]
	if r3.StructureType != CPWStructureLargeMammalCulvert {
		t.Errorf("expected large mammal culvert, got %q", r3.StructureType)
	}
	if r3.HighwayRoute != "US-550" {
		t.Errorf("expected 'US-550', got %q", r3.HighwayRoute)
	}
	if r3.EffectivenessTier != CPWEffectivenessTierMedium {
		t.Errorf("expected Medium effectiveness tier, got %q", r3.EffectivenessTier)
	}

	// 4. SH-9 Small Mammal Culvert
	r4 := records[3]
	if r4.StructureType != CPWStructureSmallMammalCulvert {
		t.Errorf("expected small mammal culvert, got %q", r4.StructureType)
	}
	if r4.EffectivenessTier != CPWEffectivenessTierLow {
		t.Errorf("expected Low effectiveness tier, got %q", r4.EffectivenessTier)
	}

	// 5. SH-9 Wildlife Jump-Out
	r5 := records[4]
	if r5.StructureType != CPWStructureWildlifeJumpOut {
		t.Errorf("expected wildlife jump-out, got %q", r5.StructureType)
	}
	if r5.EffectivenessTier != CPWEffectivenessTierEscape {
		t.Errorf("expected Escape effectiveness tier, got %q", r5.EffectivenessTier)
	}

	// 6. SH-9 8ft Ungulate Exclusion Fence
	r6 := records[5]
	if r6.StructureType != CPWStructure8ftExclusionFence {
		t.Errorf("expected 8ft ungulate exclusion fence, got %q", r6.StructureType)
	}
	if r6.FenceHeightFeet != 8.0 {
		t.Errorf("expected fence height 8.0 ft, got %v", r6.FenceHeightFeet)
	}
	if r6.EffectivenessTier != CPWEffectivenessTierHigh {
		t.Errorf("expected High effectiveness tier, got %q", r6.EffectivenessTier)
	}
}

func TestCPWParseArcGISFeatureSet(t *testing.T) {
	arcgisInput := `{
		"features": [
			{
				"attributes": {
					"OBJECTID": 201,
					"Structure_Type": "Wildlife Overpass",
					"Highway": "US-34",
					"Corridor": "Estes Park",
					"Year_Built": "2019",
					"Longitude": -105.51,
					"Latitude": 40.38
				},
				"geometry": {
					"x": -105.51,
					"y": 40.38
				}
			},
			{
				"attributes": {
					"OBJECTID": 202,
					"Structure_Type": "Large Mammal Culvert",
					"Highway": "US-550",
					"Year_Built": "2018",
					"Longitude": -107.72,
					"Latitude": 38.32
				}
			}
		]
	}`

	records, err := ParseCPWCrossings(strings.NewReader(arcgisInput))
	if err != nil {
		t.Fatalf("ParseCPWCrossings failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	if records[0].HighwayRoute != "US-34 Rocky Mountain" {
		t.Errorf("expected 'US-34 Rocky Mountain', got %q", records[0].HighwayRoute)
	}
	if records[0].ConstructionYear != 2019 {
		t.Errorf("expected construction year 2019, got %d", records[0].ConstructionYear)
	}
	if records[1].HighwayRoute != "US-550" {
		t.Errorf("expected 'US-550', got %q", records[1].HighwayRoute)
	}
}

func TestCPWFilters(t *testing.T) {
	records := []CPWCrossingRecord{
		{
			ID:                "1",
			StructureType:     CPWStructureOverpass,
			HighwayRoute:      "US-34 Rocky Mountain",
			ConstructionYear:  2020,
			EffectivenessTier: CPWEffectivenessTierHigh,
		},
		{
			ID:                "2",
			StructureType:     CPWStructureUnderpass,
			HighwayRoute:      "I-70",
			ConstructionYear:  2021,
			EffectivenessTier: CPWEffectivenessTierHigh,
		},
		{
			ID:                "3",
			StructureType:     CPWStructureLargeMammalCulvert,
			HighwayRoute:      "US-550",
			ConstructionYear:  2017,
			EffectivenessTier: CPWEffectivenessTierMedium,
		},
		{
			ID:                "4",
			StructureType:     CPWStructureWildlifeJumpOut,
			HighwayRoute:      "SH-9",
			ConstructionYear:  2016,
			EffectivenessTier: CPWEffectivenessTierEscape,
		},
	}

	// 1. Filter by Type
	overpasses := FilterCPWByType(records, CPWStructureOverpass)
	if len(overpasses) != 1 || overpasses[0].HighwayRoute != "US-34 Rocky Mountain" {
		t.Errorf("expected 1 overpass on US-34, got %v", overpasses)
	}

	// 2. Filter by Route
	i70 := FilterCPWByRoute(records, "70")
	if len(i70) != 1 || i70[0].HighwayRoute != "I-70" {
		t.Errorf("expected 1 I-70 record, got %v", i70)
	}

	// 3. Filter by Effectiveness
	highTier := FilterCPWByEffectiveness(records, CPWEffectivenessTierHigh)
	if len(highTier) != 2 {
		t.Errorf("expected 2 High tier records, got %d", len(highTier))
	}

	// 4. Filter by Year Range
	recent := FilterCPWByYearRange(records, 2020, 2025)
	if len(recent) != 2 {
		t.Errorf("expected 2 records built 2020-2025, got %d", len(recent))
	}
}

func TestCPWParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nil reader", ""},
		{"empty json", ""},
		{"invalid json", "{bad-json}"},
		{"empty feature collection", `{"type":"FeatureCollection","features":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil reader" {
				_, err := ParseCPWCrossings(nil)
				if err == nil {
					t.Error("expected error for nil reader")
				}
				return
			}
			_, err := ParseCPWCrossings(strings.NewReader(tt.input))
			if err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}
