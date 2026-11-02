package ingest

import (
	"math"
	"strings"
	"testing"

	"corridor/internal/db"
	"github.com/google/uuid"
)

// syntheticAADTGeoJSON contains synthetic test records for CA-120 and CA-140 in Yosemite.
const syntheticAADTGeoJSON = `{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "id": 5682,
      "geometry": {
        "type": "Point",
        "coordinates": [-120.066013, 37.813834]
      },
      "properties": {
        "OBJECTID": 5682,
        "DISTRICT": "10",
        "RTE": "120",
        "RTE_SFX": "",
        "CNTY": "MPA",
        "PM_PFX": "R",
        "PM": "43.09",
        "PM_SFX": "",
        "DESCRIPTION": "BUCK MEADOWS",
        "BACK_PEAK_HOUR": "1400",
        "BACK_PEAK_MADT": "3900",
        "BACK_AADT": "3750",
        "AHEAD_PEAK_HOUR": "1400",
        "AHEAD_PEAK_MADT": "9000",
        "AHEAD_AADT": "6200"
      }
    },
    {
      "type": "Feature",
      "id": 5683,
      "geometry": {
        "type": "Point",
        "coordinates": [-119.257000, 37.910800]
      },
      "properties": {
        "OBJECTID": 5683,
        "DISTRICT": "10",
        "RTE": "120",
        "RTE_SFX": "",
        "CNTY": "TUO",
        "PM_PFX": "",
        "PM": "75.00",
        "PM_SFX": "",
        "DESCRIPTION": "TIOGA PASS ENTRANCE",
        "BACK_PEAK_HOUR": "350",
        "BACK_PEAK_MADT": "1500",
        "BACK_AADT": "1200",
        "AHEAD_PEAK_HOUR": "250",
        "AHEAD_PEAK_MADT": "1000",
        "AHEAD_AADT": "800"
      }
    },
    {
      "type": "Feature",
      "id": 971,
      "geometry": {
        "type": "Point",
        "coordinates": [-120.912000, 37.289000]
      },
      "properties": {
        "OBJECTID": 971,
        "DISTRICT": "10",
        "RTE": "140",
        "RTE_SFX": "",
        "CNTY": "MER",
        "PM_PFX": "",
        "PM": "0",
        "PM_SFX": "",
        "DESCRIPTION": "JCT. RTE. 5",
        "BACK_PEAK_HOUR": "",
        "BACK_PEAK_MADT": "",
        "BACK_AADT": "",
        "AHEAD_PEAK_HOUR": "110",
        "AHEAD_PEAK_MADT": "1900",
        "AHEAD_AADT": "1400"
      }
    },
    {
      "type": "Feature",
      "id": 972,
      "geometry": {
        "type": "Point",
        "coordinates": [-119.784000, 37.674000]
      },
      "properties": {
        "OBJECTID": 972,
        "DISTRICT": "10",
        "RTE": "140",
        "RTE_SFX": "",
        "CNTY": "MPA",
        "PM_PFX": "",
        "PM": "28.50",
        "PM_SFX": "",
        "DESCRIPTION": "EL PORTAL ROAD ENTRANCE",
        "BACK_PEAK_HOUR": "300",
        "BACK_PEAK_MADT": "3000",
        "BACK_AADT": "2200",
        "AHEAD_PEAK_HOUR": "350",
        "AHEAD_PEAK_MADT": "3400",
        "AHEAD_AADT": "2600"
      }
    }
  ]
}`

// syntheticAADTCSV standard Caltrans GIS export format.
const syntheticAADTCSV = `OBJECTID,DISTRICT,RTE,RTE_SFX,CNTY,PM_PFX,PM,PM_SFX,DESCRIPTION,BACK_PEAK_HOUR,BACK_PEAK_MADT,BACK_AADT,AHEAD_PEAK_HOUR,AHEAD_PEAK_MADT,AHEAD_AADT,x,y
5682,10,120,,MPA,R,43.09,,BUCK MEADOWS,1400,3900,"3,750",1400,9000,"6,200",-120.066013,37.813834
5683,10,120,,TUO,,75.00,,TIOGA PASS,350,1500,1200,250,1000,800,-119.257000,37.910800
971,10,140,,MER,,0.00,,JCT RTE 5,,,,110,1900,1400,-120.912000,37.289000
972,10,140,,MPA,,28.50,,EL PORTAL,300,3000,2200,350,3400,2600,-119.784000,37.674000
`

// syntheticAADTCSVFriendly alternative column names format.
const syntheticAADTCSVFriendly = `District,Route,County,Postmile,BackAADT,AheadAADT,BackPeakHour,AheadPeakHour,Longitude,Latitude
10,CA-120,MPA,43.09,3750,6200,1400,1400,-120.066013,37.813834
10,CA-140,MPA,28.50,2200,2600,300,350,-119.784000,37.674000
`

func TestCaltrans_ParseGeoJSON(t *testing.T) {
	records, err := ParseCaltransAADT(strings.NewReader(syntheticAADTGeoJSON))
	if err != nil {
		t.Fatalf("ParseCaltransAADT failed: %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}

	// 1. CA-120 at Buck Meadows
	rec0 := records[0]
	if rec0.District != "10" {
		t.Errorf("record 0: expected District '10', got %q", rec0.District)
	}
	if rec0.DistrictNum() != 10 {
		t.Errorf("record 0: expected DistrictNum 10, got %d", rec0.DistrictNum())
	}
	if rec0.Route != "120" {
		t.Errorf("record 0: expected Route '120', got %q", rec0.Route)
	}
	if rec0.County != "MPA" {
		t.Errorf("record 0: expected County 'MPA', got %q", rec0.County)
	}
	if rec0.Postmile != 43.09 {
		t.Errorf("record 0: expected Postmile 43.09, got %f", rec0.Postmile)
	}
	if rec0.PostmilePrefix != "R" {
		t.Errorf("record 0: expected PostmilePrefix 'R', got %q", rec0.PostmilePrefix)
	}
	if rec0.BackAADT != 3750 {
		t.Errorf("record 0: expected BackAADT 3750, got %f", rec0.BackAADT)
	}
	if rec0.AheadAADT != 6200 {
		t.Errorf("record 0: expected AheadAADT 6200, got %f", rec0.AheadAADT)
	}
	if rec0.BackPeakHour != 1400 {
		t.Errorf("record 0: expected BackPeakHour 1400, got %f", rec0.BackPeakHour)
	}
	if rec0.AheadPeakHour != 1400 {
		t.Errorf("record 0: expected AheadPeakHour 1400, got %f", rec0.AheadPeakHour)
	}
	if math.Abs(rec0.Longitude - -120.066013) > 1e-5 || math.Abs(rec0.Latitude-37.813834) > 1e-5 {
		t.Errorf("record 0: coordinates mismatch: [%f, %f]", rec0.Longitude, rec0.Latitude)
	}
	expectedAvg0 := (3750.0 + 6200.0) / 2.0 // 4975.0
	if rec0.AverageBidirectionalAADT() != expectedAvg0 {
		t.Errorf("record 0: expected AverageBidirectionalAADT %.1f, got %.1f", expectedAvg0, rec0.AverageBidirectionalAADT())
	}

	// 2. CA-120 at Tioga Pass
	rec1 := records[1]
	if rec1.Route != "120" || rec1.County != "TUO" || rec1.Postmile != 75.0 {
		t.Errorf("record 1: route/county/postmile mismatch: %v", rec1)
	}
	expectedAvg1 := (1200.0 + 800.0) / 2.0 // 1000.0
	if rec1.AverageBidirectionalAADT() != expectedAvg1 {
		t.Errorf("record 1: expected AverageBidirectionalAADT %.1f, got %.1f", expectedAvg1, rec1.AverageBidirectionalAADT())
	}
	if rec1.AveragePeakHour() != 300.0 {
		t.Errorf("record 1: expected AveragePeakHour 300, got %f", rec1.AveragePeakHour())
	}

	// 3. CA-140 at Route Start (Jct Rte 5) - Ahead only
	rec2 := records[2]
	if rec2.Route != "140" || rec2.County != "MER" || rec2.Postmile != 0.0 {
		t.Errorf("record 2: route/county/postmile mismatch: %v", rec2)
	}
	if rec2.BackAADT != 0.0 {
		t.Errorf("record 2: expected BackAADT 0, got %f", rec2.BackAADT)
	}
	if rec2.AheadAADT != 1400.0 {
		t.Errorf("record 2: expected AheadAADT 1400, got %f", rec2.AheadAADT)
	}
	if rec2.AverageBidirectionalAADT() != 1400.0 {
		t.Errorf("record 2: expected AheadAADT fallback 1400, got %f", rec2.AverageBidirectionalAADT())
	}

	// 4. CA-140 at El Portal
	rec3 := records[3]
	if rec3.Route != "140" || rec3.County != "MPA" || rec3.Postmile != 28.5 {
		t.Errorf("record 3: route/county/postmile mismatch: %v", rec3)
	}
	expectedAvg3 := (2200.0 + 2600.0) / 2.0 // 2400.0
	if rec3.AverageBidirectionalAADT() != expectedAvg3 {
		t.Errorf("record 3: expected AverageBidirectionalAADT %.1f, got %.1f", expectedAvg3, rec3.AverageBidirectionalAADT())
	}
}

func TestCaltrans_ParseCSV(t *testing.T) {
	t.Run("Standard Caltrans GIS Header", func(t *testing.T) {
		records, err := ParseCaltransAADT(strings.NewReader(syntheticAADTCSV))
		if err != nil {
			t.Fatalf("ParseCaltransAADT CSV failed: %v", err)
		}
		if len(records) != 4 {
			t.Fatalf("expected 4 records, got %d", len(records))
		}

		// Verify record with comma thousands separator: "3,750" and "6,200"
		if records[0].BackAADT != 3750 || records[0].AheadAADT != 6200 {
			t.Errorf("record 0: BackAADT/AheadAADT parse error: got %f / %f", records[0].BackAADT, records[0].AheadAADT)
		}
		if records[0].AverageBidirectionalAADT() != 4975.0 {
			t.Errorf("record 0: expected AADT 4975, got %f", records[0].AverageBidirectionalAADT())
		}
		if math.Abs(records[0].Longitude - -120.066013) > 1e-5 {
			t.Errorf("record 0: longitude expected -120.066013, got %f", records[0].Longitude)
		}

		// Verify route start record (no BackAADT)
		if records[2].Route != "140" || records[2].AheadAADT != 1400 || records[2].BackAADT != 0 {
			t.Errorf("record 2: unexpected values: %+v", records[2])
		}
	})

	t.Run("Friendly Headers", func(t *testing.T) {
		records, err := ParseCaltransAADT(strings.NewReader(syntheticAADTCSVFriendly))
		if err != nil {
			t.Fatalf("ParseCaltransAADT Friendly CSV failed: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 records, got %d", len(records))
		}
		if records[0].Route != "CA-120" || records[0].AverageBidirectionalAADT() != 4975.0 {
			t.Errorf("record 0 mismatch: %+v", records[0])
		}
		if records[1].Route != "CA-140" || records[1].AverageBidirectionalAADT() != 2400.0 {
			t.Errorf("record 1 mismatch: %+v", records[1])
		}
	})
}

func TestCaltrans_MalformedInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Nil Reader",
			input:   "", // handled specially
			wantErr: "nil reader",
		},
		{
			name:    "Empty input",
			input:   "   \n\t  ",
			wantErr: "empty input",
		},
		{
			name:    "Malformed JSON Syntax",
			input:   `{"type": "FeatureCollection", "features": [ { broken `,
			wantErr: "decode caltrans geojson",
		},
		{
			name:    "Invalid GeoJSON Document Type",
			input:   `{"type": "Topology", "objects": {}}`,
			wantErr: "invalid caltrans GeoJSON",
		},
		{
			name: "GeoJSON Feature Missing Route",
			input: `{
				"type": "FeatureCollection",
				"features": [{
					"type": "Feature",
					"properties": {"PM": "10.0"}
				}]
			}`,
			wantErr: "missing route",
		},
		{
			name: "GeoJSON Feature Missing Postmile",
			input: `{
				"type": "FeatureCollection",
				"features": [{
					"type": "Feature",
					"properties": {"RTE": "120"}
				}]
			}`,
			wantErr: "missing postmile",
		},
		{
			name: "GeoJSON Feature Invalid Postmile Number",
			input: `{
				"type": "FeatureCollection",
				"features": [{
					"type": "Feature",
					"properties": {"RTE": "120", "PM": "NOT_A_NUMBER"}
				}]
			}`,
			wantErr: "invalid postmile",
		},
		{
			name:    "CSV Missing Route Column",
			input:   "District,County,Postmile\n10,MPA,43.09\n",
			wantErr: "missing required columns",
		},
		{
			name:    "CSV Missing Postmile Column",
			input:   "District,Route,County\n10,120,MPA\n",
			wantErr: "missing required columns",
		},
		{
			name:    "CSV Row With Invalid Postmile",
			input:   "Route,Postmile\n120,bad_postmile\n",
			wantErr: "invalid postmile",
		},
		{
			name:    "CSV Row With Invalid AADT",
			input:   "Route,Postmile,BackAADT\n120,43.09,corrupted_number\n",
			wantErr: "invalid BackAADT",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.name == "Nil Reader" {
				_, err = ParseCaltransAADT(nil)
			} else {
				_, err = ParseCaltransAADT(strings.NewReader(tc.input))
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.wantErr)) {
				t.Errorf("error %q does not contain expected substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCaltrans_MapAADTToSegments(t *testing.T) {
	records, err := ParseCaltransAADT(strings.NewReader(syntheticAADTGeoJSON))
	if err != nil {
		t.Fatalf("ParseCaltransAADT failed: %v", err)
	}

	id120Buck := uuid.New()
	id120Tioga := uuid.New()
	id140Portal := uuid.New()
	id140Start := uuid.New()
	id395 := uuid.New()
	id120NoCoords := uuid.New()

	ref120 := "CA-120"
	ref120Alt := "SR 120"
	ref140 := "CA-140"
	ref140Alt := "140"
	ref395 := "US-395"
	stateCA := "CA"

	segments := []db.SegmentSummary{
		// 1. CA-120 segment near Buck Meadows station (-120.066013, 37.813834)
		{
			ID:          id120Buck,
			HighwayRef:  &ref120,
			State:       &stateCA,
			Coordinates: [][]float64{{-120.0655, 37.8139}, {-120.0662, 37.8138}},
		},
		// 2. CA-120 segment near Tioga Pass station (-119.257000, 37.910800)
		{
			ID:          id120Tioga,
			HighwayRef:  &ref120Alt,
			State:       &stateCA,
			Coordinates: [][]float64{{-119.2565, 37.9109}, {-119.2575, 37.9107}},
		},
		// 3. CA-140 segment near El Portal entrance (-119.784000, 37.674000)
		{
			ID:          id140Portal,
			HighwayRef:  &ref140,
			State:       &stateCA,
			Coordinates: [][]float64{{-119.7838, 37.6742}, {-119.7845, 37.6738}},
		},
		// 4. CA-140 segment near Jct Rte 5 (-120.912000, 37.289000)
		{
			ID:          id140Start,
			HighwayRef:  &ref140Alt,
			State:       &stateCA,
			Coordinates: [][]float64{{-120.9115, 37.2891}, {-120.9125, 37.2889}},
		},
		// 5. US-395 segment (no matching AADT records in our dataset)
		{
			ID:          id395,
			HighwayRef:  &ref395,
			State:       &stateCA,
			Coordinates: [][]float64{{-119.000, 38.000}},
		},
		// 6. CA-120 segment with no coordinates (should average all CA-120 stations)
		{
			ID:         id120NoCoords,
			HighwayRef: &ref120,
			State:      &stateCA,
		},
	}

	aadtMap := MapAADTToSegments(records, segments)

	// Check CA-120 Buck Meadows segment -> Buck Meadows AADT = 4975.0
	if val, ok := aadtMap[id120Buck]; !ok || val != 4975.0 {
		t.Errorf("seg id120Buck: expected AADT 4975.0, got ok=%v, val=%f", ok, val)
	}

	// Check CA-120 Tioga Pass segment -> Tioga Pass AADT = 1000.0
	if val, ok := aadtMap[id120Tioga]; !ok || val != 1000.0 {
		t.Errorf("seg id120Tioga: expected AADT 1000.0, got ok=%v, val=%f", ok, val)
	}

	// Check CA-140 El Portal segment -> El Portal AADT = 2400.0
	if val, ok := aadtMap[id140Portal]; !ok || val != 2400.0 {
		t.Errorf("seg id140Portal: expected AADT 2400.0, got ok=%v, val=%f", ok, val)
	}

	// Check CA-140 Start segment -> AheadAADT = 1400.0
	if val, ok := aadtMap[id140Start]; !ok || val != 1400.0 {
		t.Errorf("seg id140Start: expected AADT 1400.0, got ok=%v, val=%f", ok, val)
	}

	// Check US-395 segment -> not mapped
	if _, ok := aadtMap[id395]; ok {
		t.Errorf("seg id395: expected omitted from map, but got %v", aadtMap[id395])
	}

	// Check CA-120 without coordinates -> average of (4975 + 1000) / 2 = 2987.5
	expectedAvg120 := (4975.0 + 1000.0) / 2.0
	if val, ok := aadtMap[id120NoCoords]; !ok || math.Abs(val-expectedAvg120) > 1e-4 {
		t.Errorf("seg id120NoCoords: expected average %f, got ok=%v, val=%f", expectedAvg120, ok, val)
	}

	// Test empty cases
	if m := MapAADTToSegments(nil, segments); len(m) != 0 {
		t.Errorf("expected empty map for nil records, got len %d", len(m))
	}
	if m := MapAADTToSegments(records, nil); len(m) != 0 {
		t.Errorf("expected empty map for nil segments, got len %d", len(m))
	}
}

func TestCaltrans_Postmiles(t *testing.T) {
	t.Run("FormatPostmile", func(t *testing.T) {
		pmStr := FormatPostmile("10", "MPA", "120", "R", 43.09, "")
		if pmStr != "10-MPA-120-R43.09" {
			t.Errorf("expected '10-MPA-120-R43.09', got %q", pmStr)
		}

		pmStr2 := FormatPostmile("", "TUO", "120", "", 75.0, "A")
		if pmStr2 != "TUO-120-75A" {
			t.Errorf("expected 'TUO-120-75A', got %q", pmStr2)
		}
	})

	t.Run("ParsePostmileReference", func(t *testing.T) {
		ref, err := ParsePostmileReference("10-MPA-120-R43.09")
		if err != nil {
			t.Fatalf("ParsePostmileReference failed: %v", err)
		}
		if ref.District != "10" || ref.County != "MPA" || ref.Route != "120" || ref.Prefix != "R" || ref.Postmile != 43.09 {
			t.Errorf("unexpected parsed reference: %+v", ref)
		}

		// Simple format
		ref2, err := ParsePostmileReference("CA-120 PM 43.09")
		if err != nil {
			t.Fatalf("ParsePostmileReference simple failed: %v", err)
		}
		if ref2.Route != "120" || ref2.Postmile != 43.09 {
			t.Errorf("unexpected parsed simple reference: %+v", ref2)
		}

		// Invalid format
		_, err = ParsePostmileReference("completely-invalid-postmile-string-xyz")
		if err == nil {
			t.Errorf("expected error for invalid postmile string, got nil")
		}
	})

	t.Run("ParseCaltransPostmiles GeoJSON", func(t *testing.T) {
		syntheticPostmileGeoJSON := `{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"id": 863,
					"geometry": {
						"type": "LineString",
						"coordinates": [
							[-119.091004, 37.886924],
							[-119.082952, 37.890870]
						]
					},
					"properties": {
						"OBJECTID": 863,
						"RouteId": "MNO120...L",
						"County": "MNO",
						"RouteNum": "120",
						"RouteSuffix": "",
						"PMPrefix": "",
						"PMSuffix": "",
						"Alignment": "L",
						"LineId": "120._L",
						"LineOrder": 1400,
						"Shape__Length": 91271.48
					}
				}
			]
		}`

		records, err := ParseCaltransPostmiles(strings.NewReader(syntheticPostmileGeoJSON))
		if err != nil {
			t.Fatalf("ParseCaltransPostmiles failed: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		rec := records[0]
		if rec.ObjectID != 863 || rec.RouteNum != "120" || rec.County != "MNO" || len(rec.Coordinates) != 2 {
			t.Errorf("unexpected postmile record values: %+v", rec)
		}
	})
}

func TestCaltrans_Artifacts(t *testing.T) {
	aadtArt := CaltransAADTArtifact()
	if aadtArt.URL != CaltransAADTFeatureServerURL {
		t.Errorf("expected AADT URL %q, got %q", CaltransAADTFeatureServerURL, aadtArt.URL)
	}
	if aadtArt.License == "" {
		t.Error("expected non-empty license for AADT artifact")
	}

	pmArt := CaltransPostmilesArtifact()
	if pmArt.URL != CaltransPostmilesFeatureServerURL {
		t.Errorf("expected Postmiles URL %q, got %q", CaltransPostmilesFeatureServerURL, pmArt.URL)
	}
	if pmArt.License == "" {
		t.Error("expected non-empty license for Postmiles artifact")
	}
}
