package ingest

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

const standardIRMAHeader = "Park Code,Observation Date,Common Name,Road/Route Name,Milepost/Location Description,Latitude,Longitude,Cause of Death,Observation ID,Comments\n"

func sampleIRMARows() string {
	return `YOSE,2024-06-15,Sierra Nevada bighorn sheep,Tioga Pass Road,MP 14.2,37.8821,-119.2612,Vehicle Strike,obs-yose-001,Adult ram struck near Tuolumne
YOSE,2024-07-20,Mule deer,Big Oak Flat Road,MP 3.5,37.7511,-119.7891,Hit by car,obs-yose-002,Found on road shoulder
YELL,2024-05-10,Grizzly bear,US-191,MP 22.0,44.8210,-111.1120,Vehicle Collision,obs-yell-001,Subadult bear on Highway 191
YELL,2024-08-01,Gray wolf,Grand Loop Road,MP 45.1,44.6200,-110.4500,Roadkill,obs-yell-002,Found dead on roadway
GRTE,2024-06-25,Rocky Mountain elk,US-89,MP 187.3,43.7800,-110.6500,Vehicle Strike,obs-grte-001,Cow elk carcass
GLAC,2024-07-12,Canada lynx,Going-to-the-Sun Road,MP 11.0,48.7100,-113.7800,Vehicle Strike,obs-glac-001,Carcass reported by ranger
OLYM,2024-04-18,Pacific fisher,US-101,MP 195.4,47.9500,-124.3200,Vehicle Collision,obs-olym-001,Rare carnivore road strike
ROMO,2024-09-02,Wolverine,Trail Ridge Road,MP 8.0,40.4200,-105.7500,Vehicle Strike,obs-romo-001,High elevation collision
YELL,2024-05-15,American bison,US-89,Gardiner Gateway,45.0300,-110.7000,Natural,obs-yell-003,Old age mortality
YOSE,2024-08-10,California black bear,Wawona Road,MP 9.1,37.5800,-119.6500,Poaching,obs-yose-003,Gunshot mortality investigated
`
}

func TestIRMACsvStreamParsing_ValidRows(t *testing.T) {
	csvData := standardIRMAHeader + sampleIRMARows()
	dataset, err := ParseIRMA(strings.NewReader(csvData), nil)
	if err != nil {
		t.Fatalf("unexpected error parsing IRMA stream: %v", err)
	}

	if dataset.QA.Read != 10 {
		t.Errorf("expected 10 rows read, got %d", dataset.QA.Read)
	}
	if dataset.QA.Accepted != 10 {
		t.Errorf("expected 10 rows accepted, got %d", dataset.QA.Accepted)
	}
	if dataset.QA.Rejected != 0 {
		t.Errorf("expected 0 rejected, got %d (%v)", dataset.QA.Rejected, dataset.QA.Reasons)
	}
	if dataset.QA.Duplicates != 0 {
		t.Errorf("expected 0 duplicates, got %d", dataset.QA.Duplicates)
	}
	if dataset.QA.RoadkillCount != 8 {
		t.Errorf("expected 8 roadkill records, got %d", dataset.QA.RoadkillCount)
	}
	if dataset.QA.SensitiveCount != 7 {
		t.Errorf("expected 7 sensitive records (bighorn, grizzly, wolf, lynx, fisher, wolverine, black bear), got %d", dataset.QA.SensitiveCount)
	}
	if len(dataset.Records) != 10 {
		t.Fatalf("expected 10 parsed records, got %d", len(dataset.Records))
	}

	// Verify specific record properties
	r0 := dataset.Records[0]
	if r0.ParkCode != "YOSE" {
		t.Errorf("record 0 expected park YOSE, got %s", r0.ParkCode)
	}
	if r0.CommonName != "Sierra Nevada bighorn sheep" {
		t.Errorf("record 0 expected common name Sierra Nevada bighorn sheep, got %s", r0.CommonName)
	}
	if !r0.IsSensitive {
		t.Errorf("record 0 should be marked sensitive")
	}
	if !r0.IsRoadkill {
		t.Errorf("record 0 should be marked roadkill")
	}
	if r0.H3R6 == "" {
		t.Errorf("record 0 missing H3-r6 cell")
	}
	cell, err := ParseH3Cell(r0.H3R6)
	if err != nil || cell.Resolution() != DefaultH3Resolution {
		t.Errorf("record 0 invalid H3-r6 cell %s: %v (res %d)", r0.H3R6, err, cell.Resolution())
	}
}

func TestIRMAParkFilter(t *testing.T) {
	csvData := standardIRMAHeader + sampleIRMARows()

	tests := []struct {
		name          string
		parkCodes     []string
		expectedCount int
		expectedParks map[string]bool
	}{
		{
			name:          "Filter YOSE only",
			parkCodes:     []string{"YOSE"},
			expectedCount: 3,
			expectedParks: map[string]bool{"YOSE": true},
		},
		{
			name:          "Filter YELL only",
			parkCodes:     []string{"YELL"},
			expectedCount: 3,
			expectedParks: map[string]bool{"YELL": true},
		},
		{
			name:          "Filter YOSE and YELL",
			parkCodes:     []string{"YOSE", "YELL"},
			expectedCount: 6,
			expectedParks: map[string]bool{"YOSE": true, "YELL": true},
		},
		{
			name:          "Filter case-insensitive (yose, yell)",
			parkCodes:     []string{"yose", "yell"},
			expectedCount: 6,
			expectedParks: map[string]bool{"YOSE": true, "YELL": true},
		},
		{
			name:          "Filter GLAC only",
			parkCodes:     []string{"GLAC"},
			expectedCount: 1,
			expectedParks: map[string]bool{"GLAC": true},
		},
		{
			name:          "Filter Non-existent park (ZION)",
			parkCodes:     []string{"ZION"},
			expectedCount: 0,
			expectedParks: map[string]bool{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filter := &IRMAFilter{
				ParkCodes: tc.parkCodes,
			}
			ds, err := ParseIRMA(strings.NewReader(csvData), filter)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if len(ds.Records) != tc.expectedCount {
				t.Fatalf("expected %d records, got %d", tc.expectedCount, len(ds.Records))
			}
			if ds.QA.Accepted != int64(tc.expectedCount) {
				t.Errorf("expected %d QA accepted, got %d", tc.expectedCount, ds.QA.Accepted)
			}
			expectedOutside := int64(10 - tc.expectedCount)
			if ds.QA.OutsideScope != expectedOutside {
				t.Errorf("expected %d outside scope, got %d", expectedOutside, ds.QA.OutsideScope)
			}
			for _, rec := range ds.Records {
				if !tc.expectedParks[rec.ParkCode] {
					t.Errorf("unexpected park code %s in results", rec.ParkCode)
				}
			}
		})
	}
}

func TestIRMASensitiveSpeciesDetectionAndPrivacy(t *testing.T) {
	// Table of sensitive species from config/sensitive_species.yaml
	sensitiveTests := []struct {
		inputName      string
		expectedCommon string
		expectedSci    string
		park           string
	}{
		{"Sierra Nevada bighorn sheep", "Sierra Nevada bighorn sheep", "Ovis canadensis sierrae", "YOSE"},
		{"sierra bighorn sheep", "Sierra Nevada bighorn sheep", "Ovis canadensis sierrae", "YOSE"},
		{"bighorn sheep", "Sierra Nevada bighorn sheep", "Ovis canadensis sierrae", "YOSE"},
		{"Ovis canadensis sierrae", "Sierra Nevada bighorn sheep", "Ovis canadensis sierrae", "YOSE"},
		{"Pacific fisher", "Pacific fisher", "Pekania pennanti", "OLYM"},
		{"fisher", "Pacific fisher", "Pekania pennanti", "OLYM"},
		{"Pekania pennanti", "Pacific fisher", "Pekania pennanti", "OLYM"},
		{"California black bear", "California black bear", "Ursus americanus californiensis", "YOSE"},
		{"black bear", "California black bear", "Ursus americanus californiensis", "YOSE"},
		{"Ursus americanus", "California black bear", "Ursus americanus californiensis", "YOSE"},
		{"Grizzly bear", "Grizzly bear", "Ursus arctos horribilis", "YELL"},
		{"grizzly", "Grizzly bear", "Ursus arctos horribilis", "YELL"},
		{"Ursus arctos horribilis", "Grizzly bear", "Ursus arctos horribilis", "YELL"},
		{"Canada lynx", "Canada lynx", "Lynx canadensis", "GLAC"},
		{"lynx", "Canada lynx", "Lynx canadensis", "GLAC"},
		{"Lynx canadensis", "Canada lynx", "Lynx canadensis", "GLAC"},
		{"Gray wolf", "Gray wolf", "Canis lupus", "YELL"},
		{"grey wolf", "Gray wolf", "Canis lupus", "YELL"},
		{"wolf", "Gray wolf", "Canis lupus", "YELL"},
		{"Canis lupus", "Gray wolf", "Canis lupus", "YELL"},
		{"Wolverine", "Wolverine", "Gulo gulo luscus", "ROMO"},
		{"gulo gulo", "Wolverine", "Gulo gulo luscus", "ROMO"},
	}

	for _, tc := range sensitiveTests {
		t.Run(tc.inputName, func(t *testing.T) {
			obsDateStr := "2024-06-01"
			row := fmt.Sprintf("%s,%s,%s,Main Road,MP 1.0,44.5000,-110.5000,Vehicle Strike,id-123,\n", tc.park, obsDateStr, tc.inputName)
			csvData := standardIRMAHeader + row

			obsDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
			refTimeDuringQuarantine := obsDate.Add(10 * 24 * time.Hour) // 10 days after (in quarantine)
			refTimeAfterQuarantine := obsDate.Add(40 * 24 * time.Hour)  // 40 days after (released)

			filter := &IRMAFilter{ReferenceTime: &refTimeDuringQuarantine}
			ds, err := ParseIRMA(strings.NewReader(csvData), filter)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if len(ds.Records) != 1 {
				t.Fatalf("expected 1 record, got %d", len(ds.Records))
			}
			rec := ds.Records[0]

			if !rec.IsSensitive {
				t.Errorf("species %q should be sensitive", tc.inputName)
			}
			if rec.CommonName != tc.expectedCommon {
				t.Errorf("expected common name %q, got %q", tc.expectedCommon, rec.CommonName)
			}
			if rec.ScientificName != tc.expectedSci {
				t.Errorf("expected scientific name %q, got %q", tc.expectedSci, rec.ScientificName)
			}

			// Verify Uber H3 resolution 6 cell assignment
			if rec.H3R6 == "" {
				t.Fatalf("expected non-empty H3-r6 cell")
			}
			cell, err := ParseH3Cell(rec.H3R6)
			if err != nil {
				t.Errorf("H3-r6 cell %s is invalid: %v", rec.H3R6, err)
			}
			if cell.Resolution() != 6 {
				t.Errorf("expected H3 resolution 6, got %d", cell.Resolution())
			}

			// Verify H3 centroid is within reasonable bounds of original coordinates
			if math.Abs(rec.GeneralizedLat-rec.Latitude) > 0.5 || math.Abs(rec.GeneralizedLon-rec.Longitude) > 0.5 {
				t.Errorf("generalized coordinates (%.4f, %.4f) too far from raw (%.4f, %.4f)",
					rec.GeneralizedLat, rec.GeneralizedLon, rec.Latitude, rec.Longitude)
			}

			// Verify 30-day quarantine duration
			expectedQuarantineUntil := obsDate.Add(30 * 24 * time.Hour)
			if !rec.QuarantineUntil.Equal(expectedQuarantineUntil) {
				t.Errorf("expected quarantine until %v, got %v", expectedQuarantineUntil, rec.QuarantineUntil)
			}
			if !rec.InQuarantine {
				t.Errorf("expected record to be in quarantine at 10 days post-observation")
			}

			// Test PublicRecord during quarantine -> must be withheld
			if _, ok := rec.PublicRecord(refTimeDuringQuarantine); ok {
				t.Errorf("public record must be withheld during 30-day quarantine")
			}

			// Test PublicRecord after quarantine -> must be published with generalized coordinates
			pub, ok := rec.PublicRecord(refTimeAfterQuarantine)
			if !ok {
				t.Fatalf("public record should be available after 30-day quarantine")
			}
			if pub.Latitude != rec.GeneralizedLat || pub.Longitude != rec.GeneralizedLon {
				t.Errorf("public record must use generalized H3 centroid coords (got %.4f, %.4f, expected %.4f, %.4f)",
					pub.Latitude, pub.Longitude, rec.GeneralizedLat, rec.GeneralizedLon)
			}
		})
	}
}

func TestIRMANonSensitiveSpecies(t *testing.T) {
	nonSensitive := []string{
		"Mule deer",
		"Odocoileus hemionus",
		"Rocky Mountain elk",
		"Cervus canadensis",
		"Pronghorn",
		"American bison",
		"Coyote",
		"Bobcat",
		"Mountain lion",
	}

	for _, sp := range nonSensitive {
		t.Run(sp, func(t *testing.T) {
			row := fmt.Sprintf("YELL,2024-06-01,%s,US-191,MP 5.0,44.5000,-110.5000,Vehicle Strike,id-ns,\n", sp)
			ds, err := ParseIRMA(strings.NewReader(standardIRMAHeader+row), nil)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if len(ds.Records) != 1 {
				t.Fatalf("expected 1 record, got %d", len(ds.Records))
			}
			rec := ds.Records[0]
			if rec.IsSensitive {
				t.Errorf("species %q should NOT be sensitive", sp)
			}
			if rec.InQuarantine {
				t.Errorf("non-sensitive species should not be in quarantine")
			}

			// Public record should immediately retain coordinates
			now := time.Now().UTC()
			pub, ok := rec.PublicRecord(now)
			if !ok {
				t.Fatalf("non-sensitive species public record must be available")
			}
			if pub.Latitude != rec.Latitude || pub.Longitude != rec.Longitude {
				t.Errorf("non-sensitive species should preserve precise coordinates in public record")
			}
		})
	}
}

func TestIRMAMissingCoordinates(t *testing.T) {
	rows := standardIRMAHeader + `YOSE,2024-06-15,Mule deer,Tioga Road,MP 12.0,,,Vehicle Strike,obs-mc-1,No GPS coordinates recorded
YELL,2024-06-16,Elk,US-191,Near West Entrance,44.6000,-111.0000,Vehicle Strike,obs-mc-2,With GPS
YELL,2024-06-17,Grizzly bear,Grand Loop,MP 20.0,,,Vehicle Strike,obs-mc-3,Sensitive without GPS
`
	// 1. Default mode: RequireCoordinates = false
	ds, err := ParseIRMA(strings.NewReader(rows), &IRMAFilter{RequireCoordinates: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds.QA.Accepted != 3 {
		t.Errorf("expected all 3 rows accepted, got %d", ds.QA.Accepted)
	}
	if ds.QA.MissingCoords != 2 {
		t.Errorf("expected 2 records with missing coords, got %d", ds.QA.MissingCoords)
	}

	rec0 := ds.Records[0]
	if rec0.HasCoordinates {
		t.Errorf("record 0 should have HasCoordinates=false")
	}
	if rec0.Latitude != 0 || rec0.Longitude != 0 {
		t.Errorf("record 0 should have zero coordinates, got %f, %f", rec0.Latitude, rec0.Longitude)
	}
	if rec0.H3R6 != "" {
		t.Errorf("record 0 without coordinates should have empty H3-r6 cell")
	}

	rec2 := ds.Records[2]
	if !rec2.IsSensitive {
		t.Errorf("record 2 (Grizzly bear) should still be flagged sensitive even without GPS")
	}

	// 2. Strict mode: RequireCoordinates = true
	dsStrict, err := ParseIRMA(strings.NewReader(rows), &IRMAFilter{RequireCoordinates: true})
	if err != nil {
		t.Fatalf("unexpected error in strict mode: %v", err)
	}
	if dsStrict.QA.Accepted != 1 {
		t.Errorf("expected 1 accepted in strict mode, got %d", dsStrict.QA.Accepted)
	}
	if dsStrict.QA.Rejected != 2 {
		t.Errorf("expected 2 rejected in strict mode, got %d", dsStrict.QA.Rejected)
	}
	if dsStrict.QA.Reasons["missing coordinates"] != 2 {
		t.Errorf("expected 2 rejections for 'missing coordinates', got %d", dsStrict.QA.Reasons["missing coordinates"])
	}
}

func TestIRMAUnknownSpecies(t *testing.T) {
	rows := standardIRMAHeader + `YOSE,2024-06-15,Unknown small mammal,Tioga Road,MP 5.0,37.8000,-119.5000,Vehicle Strike,obs-unk-1,Carcass damaged
GLAC,2024-07-01,Sciurus griseus,Sun Road,MP 2.0,48.5000,-113.5000,Vehicle Strike,obs-unk-2,Western gray squirrel
YELL,2024-08-12,Mystery Carnivore,Highway 89,MP 3.0,45.1000,-110.8000,Vehicle Strike,obs-unk-3,Unidentifiable specimen
`
	ds, err := ParseIRMA(strings.NewReader(rows), nil)
	if err != nil {
		t.Fatalf("unexpected error parsing unknown species: %v", err)
	}

	if ds.QA.Accepted != 3 {
		t.Errorf("expected 3 accepted unknown species rows, got %d", ds.QA.Accepted)
	}
	if ds.QA.Rejected != 0 {
		t.Errorf("expected 0 rejected, got %d", ds.QA.Rejected)
	}

	rec0 := ds.Records[0]
	if rec0.CommonName != "Unknown small mammal" {
		t.Errorf("expected CommonName to retain 'Unknown small mammal', got %q", rec0.CommonName)
	}
	if rec0.IsSensitive {
		t.Errorf("unknown species should not be flagged sensitive")
	}
}

func TestIRMADateRangeFilter(t *testing.T) {
	rows := standardIRMAHeader + `YOSE,2024-05-01,Mule deer,CA-120,MP 1.0,37.8,-119.5,Vehicle Strike,d1,
YOSE,2024-06-15,Mule deer,CA-120,MP 2.0,37.8,-119.5,Vehicle Strike,d2,
YOSE,2024-07-31,Mule deer,CA-120,MP 3.0,37.8,-119.5,Vehicle Strike,d3,
YOSE,2024-09-01,Mule deer,CA-120,MP 4.0,37.8,-119.5,Vehicle Strike,d4,
`
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)

	filter := &IRMAFilter{
		StartDate: &start,
		EndDate:   &end,
	}

	ds, err := ParseIRMA(strings.NewReader(rows), filter)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if ds.QA.Accepted != 2 {
		t.Fatalf("expected 2 records in date range [June, July], got %d", ds.QA.Accepted)
	}
	if ds.QA.OutsideScope != 2 {
		t.Errorf("expected 2 outside date range, got %d", ds.QA.OutsideScope)
	}
	if ds.Records[0].RawID != "d2" || ds.Records[1].RawID != "d3" {
		t.Errorf("unexpected records returned: %s, %s", ds.Records[0].RawID, ds.Records[1].RawID)
	}
}

func TestIRMAMortalityCauseFilter(t *testing.T) {
	rows := standardIRMAHeader + `YELL,2024-06-01,Elk,US-191,MP 1.0,44.5,-111.0,Vehicle Strike,c1,
YELL,2024-06-02,Elk,US-191,MP 2.0,44.5,-111.0,Chronic Wasting Disease,c2,
YELL,2024-06-03,Elk,US-191,MP 3.0,44.5,-111.0,Hit by car,c3,
YELL,2024-06-04,Elk,US-191,MP 4.0,44.5,-111.0,Poaching gunshot,c4,
YELL,2024-06-05,Elk,US-191,MP 5.0,44.5,-111.0,Roadkill on highway,c5,
`
	// 1. OnlyRoadkill: true
	dsRoadkill, err := ParseIRMA(strings.NewReader(rows), &IRMAFilter{OnlyRoadkill: true})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if dsRoadkill.QA.Accepted != 3 {
		t.Errorf("expected 3 roadkill records accepted, got %d", dsRoadkill.QA.Accepted)
	}
	if dsRoadkill.QA.OutsideScope != 2 {
		t.Errorf("expected 2 non-roadkill filtered out, got %d", dsRoadkill.QA.OutsideScope)
	}

	// 2. MortalityCause: "disease"
	dsDisease, err := ParseIRMA(strings.NewReader(rows), &IRMAFilter{MortalityCause: "disease"})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if dsDisease.QA.Accepted != 1 {
		t.Errorf("expected 1 disease record accepted, got %d", dsDisease.QA.Accepted)
	}
	if dsDisease.Records[0].RawID != "c2" {
		t.Errorf("expected record c2, got %s", dsDisease.Records[0].RawID)
	}
}

func TestIRMARejectsMalformedRows(t *testing.T) {
	tests := []struct {
		name           string
		csvRow         string
		expectedReason string
	}{
		{
			name:           "Missing park code",
			csvRow:         ",2024-06-15,Mule deer,Tioga Road,MP 1.0,37.8,-119.5,Vehicle Strike,id1,\n",
			expectedReason: "missing park code",
		},
		{
			name:           "Empty species",
			csvRow:         "YOSE,2024-06-15,,Tioga Road,MP 1.0,37.8,-119.5,Vehicle Strike,id2,\n",
			expectedReason: "empty species",
		},
		{
			name:           "Invalid date format",
			csvRow:         "YOSE,not-a-date,Mule deer,Tioga Road,MP 1.0,37.8,-119.5,Vehicle Strike,id3,\n",
			expectedReason: "unrecognized observation date format",
		},
		{
			name:           "Invalid date year out of range",
			csvRow:         "YOSE,1850-05-10,Mule deer,Tioga Road,MP 1.0,37.8,-119.5,Vehicle Strike,id4,\n",
			expectedReason: "observation date year out of range",
		},
		{
			name:           "Latitude > 90",
			csvRow:         "YOSE,2024-06-15,Mule deer,Tioga Road,MP 1.0,95.5,-119.5,Vehicle Strike,id5,\n",
			expectedReason: "invalid coordinates out of bounds",
		},
		{
			name:           "Longitude < -180",
			csvRow:         "YOSE,2024-06-15,Mule deer,Tioga Road,MP 1.0,37.8,-185.0,Vehicle Strike,id6,\n",
			expectedReason: "invalid coordinates out of bounds",
		},
		{
			name:           "Non-numeric coordinate (NaN)",
			csvRow:         "YOSE,2024-06-15,Mule deer,Tioga Road,MP 1.0,NaN,-119.5,Vehicle Strike,id7,\n",
			expectedReason: "invalid coordinates",
		},
		{
			name:           "Non-numeric coordinate (text)",
			csvRow:         "YOSE,2024-06-15,Mule deer,Tioga Road,MP 1.0,North,-119.5,Vehicle Strike,id8,\n",
			expectedReason: "invalid coordinates",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			csvData := standardIRMAHeader + tc.csvRow
			ds, err := ParseIRMA(strings.NewReader(csvData), nil)
			if err != nil {
				t.Fatalf("ParseIRMA returned unexpected top-level error: %v", err)
			}
			if ds.QA.Rejected != 1 {
				t.Errorf("expected 1 rejected row, got %d", ds.QA.Rejected)
			}
			if len(ds.Records) != 0 {
				t.Errorf("expected 0 accepted records, got %d", len(ds.Records))
			}
			if ds.QA.Reasons[tc.expectedReason] != 1 {
				t.Errorf("expected reason %q, got reasons: %+v", tc.expectedReason, ds.QA.Reasons)
			}
		})
	}
}

func TestIRMADuplicateDetection(t *testing.T) {
	row := "YOSE,2024-06-15,Mule deer,Tioga Road,MP 1.0,37.8,-119.5,Vehicle Strike,dup-001,\n"
	csvData := standardIRMAHeader + row + row + row

	ds, err := ParseIRMA(strings.NewReader(csvData), nil)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if ds.QA.Read != 3 {
		t.Errorf("expected 3 rows read, got %d", ds.QA.Read)
	}
	if ds.QA.Accepted != 1 {
		t.Errorf("expected 1 record accepted, got %d", ds.QA.Accepted)
	}
	if ds.QA.Duplicates != 2 {
		t.Errorf("expected 2 duplicates detected, got %d", ds.QA.Duplicates)
	}
}

func TestIRMAHeaderVariations(t *testing.T) {
	headers := []string{
		"Park Code,Observation Date,Common Name,Road/Route Name,Milepost/Location Description,Latitude,Longitude,Cause of Death,Observation ID",
		"ParkCode,ObservationDate,Species,Road,Milepost,Latitude,Longitude,CauseOfDeath,RecordID",
		"Unit_Code,Obs_Date,Animal,Route,Location_Description,Lat,Lon,Mortality_Type,ID",
		"park,date,scientific_name,road_name,milepost,y,x,mortality_cause,objectid",
	}

	for i, h := range headers {
		t.Run(fmt.Sprintf("HeaderFormat_%d", i), func(t *testing.T) {
			row := "YOSE,2024-06-15,Elk,CA-120,MP 5.0,37.8,-119.5,Vehicle Strike,rec-01\n"
			csvData := h + "\n" + row
			ds, err := ParseIRMA(strings.NewReader(csvData), nil)
			if err != nil {
				t.Fatalf("failed with header %q: %v", h, err)
			}
			if len(ds.Records) != 1 {
				t.Fatalf("expected 1 record accepted, got %d", len(ds.Records))
			}
			rec := ds.Records[0]
			if rec.ParkCode != "YOSE" || rec.NormalizedSpecies != "elk" {
				t.Errorf("mismatched parsed values: %+v", rec)
			}
		})
	}
}

func TestIRMAToCorridorRecord(t *testing.T) {
	csvData := standardIRMAHeader +
		"YOSE,2024-06-15,Sierra Nevada bighorn sheep,Tioga Road,MP 14.0,37.8821,-119.2612,Vehicle Strike,rec-sens,\n" +
		"YELL,2024-07-20,Mule deer,US-191,MP 22.0,44.8210,-111.1120,Vehicle Strike,rec-norm,\n"

	ds, err := ParseIRMA(strings.NewReader(csvData), nil)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ds.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(ds.Records))
	}

	// 1. Sensitive record conversion
	sensCorridor := ds.Records[0].ToRecord()
	if sensCorridor.ID != "rec-sens" {
		t.Errorf("expected ID rec-sens, got %s", sensCorridor.ID)
	}
	if sensCorridor.Species != "Ovis canadensis sierrae" {
		t.Errorf("expected scientific species Ovis canadensis sierrae, got %s", sensCorridor.Species)
	}
	if sensCorridor.Uncertainty != 6000.0 {
		t.Errorf("expected 6000m uncertainty for generalized sensitive record, got %f", sensCorridor.Uncertainty)
	}
	if sensCorridor.Survey != "NPS-IRMA-YOSE" {
		t.Errorf("expected Survey NPS-IRMA-YOSE, got %s", sensCorridor.Survey)
	}

	// 2. Normal record conversion
	normCorridor := ds.Records[1].ToRecord()
	if normCorridor.ID != "rec-norm" {
		t.Errorf("expected ID rec-norm, got %s", normCorridor.ID)
	}
	if normCorridor.Species != "Odocoileus hemionus" {
		t.Errorf("expected species Odocoileus hemionus, got %s", normCorridor.Species)
	}
	if normCorridor.Uncertainty != 50.0 {
		t.Errorf("expected 50m uncertainty for standard GPS record, got %f", normCorridor.Uncertainty)
	}

	// 3. InQuarantineAt on non-sensitive record
	if ds.Records[1].InQuarantineAt(time.Now()) {
		t.Errorf("non-sensitive record should never be in quarantine")
	}

	// 4. Month and Year precision in ToRecord
	monthRec := ds.Records[1]
	monthRec.DatePrecision = "month"
	monthCorridor := monthRec.ToRecord()
	if monthCorridor.End != monthRec.ObservationDate.AddDate(0, 1, 0) {
		t.Errorf("expected month precision end date, got %v", monthCorridor.End)
	}

	yearRec := ds.Records[1]
	yearRec.DatePrecision = "year"
	yearCorridor := yearRec.ToRecord()
	if yearCorridor.End != yearRec.ObservationDate.AddDate(1, 0, 0) {
		t.Errorf("expected year precision end date, got %v", yearCorridor.End)
	}

	noCoordRec := ds.Records[1]
	noCoordRec.HasCoordinates = false
	noCoordCorridor := noCoordRec.ToRecord()
	if noCoordCorridor.Uncertainty != 50000.0 {
		t.Errorf("expected 50000m uncertainty for record without coordinates, got %f", noCoordCorridor.Uncertainty)
	}
}

func TestIRMATaxonomyAndH3EdgeCases(t *testing.T) {
	reg := getTaxonomyRegistry()

	// Empty string normalization
	emptyTaxon := reg.Normalize("", "YOSE")
	if emptyTaxon.CommonName != "Unknown" {
		t.Errorf("expected Unknown for empty species, got %s", emptyTaxon.CommonName)
	}

	// Slash-separated name
	slashTaxon := reg.Normalize("Elk / Cervus canadensis", "YELL")
	if slashTaxon.CommonName != "Rocky Mountain elk" {
		t.Errorf("expected Rocky Mountain elk, got %s", slashTaxon.CommonName)
	}

	// ParseH3Cell invalid strings
	if _, err := ParseH3Cell("not-hex"); err == nil {
		t.Errorf("expected error for non-hex H3 string")
	}
	if _, err := ParseH3Cell("0"); err == nil {
		t.Errorf("expected error for invalid H3 cell index")
	}

	// IsVehicleStrike false conditions
	if IsVehicleStrike("") {
		t.Errorf("empty cause should not be vehicle strike")
	}
	if IsVehicleStrike("Disease / Avian Flu") {
		t.Errorf("disease should not be vehicle strike")
	}
	if !IsVehicleStrike("Highway kill on road") {
		t.Errorf("highway kill should be recognized as vehicle strike")
	}

	// Missing header columns
	if _, err := ParseIRMA(strings.NewReader("col1,col2\n1,2\n"), nil); err == nil {
		t.Errorf("expected error for missing required columns")
	}
}

func TestIRMAArtifact(t *testing.T) {
	a := IRMAArtifact()
	if a.SourceID == "" {
		t.Errorf("expected non-empty SourceID")
	}
	if !strings.Contains(a.Name, "2317362") {
		t.Errorf("expected reference 2317362 in name, got %s", a.Name)
	}
	if !strings.Contains(a.URL, "2317362") {
		t.Errorf("expected reference 2317362 in URL, got %s", a.URL)
	}
	if !strings.Contains(a.License, "Public Domain") {
		t.Errorf("expected Public Domain in license, got %s", a.License)
	}
}
