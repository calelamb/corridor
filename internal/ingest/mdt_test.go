package ingest

import (
	"math"
	"strings"
	"testing"

	"corridor/internal/geo"
)

func TestMDTCarcassArtifact(t *testing.T) {
	artifact := MDTCarcassArtifact()
	if artifact.SourceID == "" {
		t.Error("expected non-empty SourceID")
	}
	if len(artifact.SHA256) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars: %s", len(artifact.SHA256), artifact.SHA256)
	}
	if !strings.Contains(artifact.URL, "mdt.mt.gov") {
		t.Errorf("expected mdt.mt.gov in URL, got %s", artifact.URL)
	}
	if artifact.RetrievedAt.IsZero() {
		t.Error("expected non-zero RetrievedAt")
	}
}

func TestMDTProjectionAccuracy_Benchmarks(t *testing.T) {
	// Official NOAA NGS NCAT Montana State Plane NAD83 (EPSG:32100) ground-truth benchmarks
	benchmarks := []struct {
		name     string
		easting  float64
		northing float64
		wantLat  float64
		wantLon  float64
	}{
		{
			name:     "Helena, MT",
			easting:  405557.683,
			northing: 263078.179,
			wantLat:  46.589100,
			wantLon:  -112.039100,
		},
		{
			name:     "Bozeman, MT",
			easting:  479836.847,
			northing: 159775.029,
			wantLat:  45.677000,
			wantLon:  -111.042900,
		},
		{
			name:     "Missoula, MT",
			easting:  257784.456,
			northing: 301187.545,
			wantLat:  46.872100,
			wantLon:  -113.994000,
		},
		{
			name:     "Billings, MT",
			easting:  677679.089,
			northing: 170897.584,
			wantLat:  45.783300,
			wantLon:  -108.500700,
		},
	}

	for _, bm := range benchmarks {
		t.Run(bm.name, func(t *testing.T) {
			lat, lon, err := geo.StatePlane32100ToWGS84(bm.easting, bm.northing)
			if err != nil {
				t.Fatalf("unexpected error converting state plane coordinates: %v", err)
			}

			// Sub-meter angular accuracy (~1e-5 degrees)
			if math.Abs(lat-bm.wantLat) > 1e-5 {
				t.Errorf("lat mismatch for %s: got %.6f, want %.6f (diff: %.8f)", bm.name, lat, bm.wantLat, math.Abs(lat-bm.wantLat))
			}
			if math.Abs(lon-bm.wantLon) > 1e-5 {
				t.Errorf("lon mismatch for %s: got %.6f, want %.6f (diff: %.8f)", bm.name, lon, bm.wantLon, math.Abs(lon-bm.wantLon))
			}

			// Forward projection round-trip
			roundEasting, roundNorthing, err := geo.WGS84ToStatePlane32100(lat, lon)
			if err != nil {
				t.Fatalf("unexpected error converting wgs84 coordinates: %v", err)
			}
			if math.Abs(roundEasting-bm.easting) > 0.05 {
				t.Errorf("easting roundtrip mismatch: got %.3f, want %.3f", roundEasting, bm.easting)
			}
			if math.Abs(roundNorthing-bm.northing) > 0.05 {
				t.Errorf("northing roundtrip mismatch: got %.3f, want %.3f", roundNorthing, bm.northing)
			}
		})
	}
}

func TestMDTSpeciesNormalization(t *testing.T) {
	tests := []struct {
		input       string
		wantSpecies string
		isSensitive bool
	}{
		// Common ungulates (non-sensitive)
		{"White-tailed Deer", "white_tailed_deer", false},
		{"White Tailed Deer", "white_tailed_deer", false},
		{"white-tail deer", "white_tailed_deer", false},
		{"Whitetail", "white_tailed_deer", false},
		{"WT Deer", "white_tailed_deer", false},
		{"Mule Deer", "mule_deer", false},
		{"MD Deer", "mule_deer", false},
		{"Elk", "elk", false},
		{"Moose", "moose", false},
		{"Pronghorn", "pronghorn", false},
		{"Antelope", "pronghorn", false},
		{"Bighorn Sheep", "bighorn_sheep", false},
		{"Mountain Goat", "mountain_goat", false},
		{"Black Bear", "black_bear", false},

		// Sensitive species (must be generalized to H3 resolution 6)
		{"Grizzly Bear", "grizzly_bear", true},
		{"Grizzly", "grizzly_bear", true},
		{"Ursus arctos horribilis", "grizzly_bear", true},
		{"Canada Lynx", "lynx", true},
		{"Lynx", "lynx", true},
		{"Lynx canadensis", "lynx", true},
		{"Wolf", "gray_wolf", true},
		{"Gray Wolf", "gray_wolf", true},
		{"Grey Wolf", "gray_wolf", true},
		{"Timber Wolf", "gray_wolf", true},
		{"Wolverine", "wolverine", true},

		// Other fauna
		{"Coyote", "coyote", false},
		{"Mountain Lion", "mountain_lion", false},
		{"Bobcat", "bobcat", false},
		{"Wild Turkey", "wild_turkey", false},
		{"River Otter", "river_otter", false},
		{"Unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeMDTSpecies(tt.input)
			if got != tt.wantSpecies {
				t.Errorf("NormalizeMDTSpecies(%q) = %q, want %q", tt.input, got, tt.wantSpecies)
			}

			sens := IsMDTSensitiveSpecies(tt.input)
			if sens != tt.isSensitive {
				t.Errorf("IsMDTSensitiveSpecies(%q) = %v, want %v", tt.input, sens, tt.isSensitive)
			}
		})
	}
}

func TestMDTCarcassCSV_Basic(t *testing.T) {
	// Sample CSV with White-tailed Deer near Bozeman (479836.847, 159775.029)
	csvData := `Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD
10/14/2021,US-191,12.5,White-tailed Deer,Male,1,479836.847,159775.029
2022-05-18,I-90,305.2,Elk,Female,2,405557.683,263078.179
`

	records, err := ParseMDTCarcassCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Verify Record 1 (White-tailed deer near Bozeman)
	r1 := records[0]
	if r1.Route != "US-191" {
		t.Errorf("expected Route 'US-191', got %q", r1.Route)
	}
	if r1.Milepost != 12.5 {
		t.Errorf("expected Milepost 12.5, got %f", r1.Milepost)
	}
	if r1.AnimalType != "White-tailed Deer" {
		t.Errorf("expected AnimalType 'White-tailed Deer', got %q", r1.AnimalType)
	}
	if r1.NormalizedSpecies != "white_tailed_deer" {
		t.Errorf("expected NormalizedSpecies 'white_tailed_deer', got %q", r1.NormalizedSpecies)
	}
	if r1.Gender != "male" {
		t.Errorf("expected Gender 'male', got %q", r1.Gender)
	}
	if r1.Quantity != 1 {
		t.Errorf("expected Quantity 1, got %d", r1.Quantity)
	}
	if r1.IsSensitive {
		t.Errorf("White-tailed Deer must NOT be flagged sensitive")
	}

	// For non-sensitive species, Latitude and Longitude should equal RawLatitude and RawLongitude
	if math.Abs(r1.Latitude-45.6770) > 1e-4 {
		t.Errorf("expected lat ~45.6770, got %f", r1.Latitude)
	}
	if math.Abs(r1.Longitude-(-111.0429)) > 1e-4 {
		t.Errorf("expected lon ~ -111.0429, got %f", r1.Longitude)
	}
	if r1.Latitude != r1.RawLatitude || r1.Longitude != r1.RawLongitude {
		t.Errorf("non-sensitive species coordinates must match raw coordinates")
	}
	if r1.H3R6 == "" {
		t.Error("expected non-empty H3R6 cell")
	}

	// Verify Record 2 (Elk near Helena)
	r2 := records[1]
	if r2.Route != "I-90" {
		t.Errorf("expected Route 'I-90', got %q", r2.Route)
	}
	if r2.Milepost != 305.2 {
		t.Errorf("expected Milepost 305.2, got %f", r2.Milepost)
	}
	if r2.NormalizedSpecies != "elk" {
		t.Errorf("expected NormalizedSpecies 'elk', got %q", r2.NormalizedSpecies)
	}
	if r2.Gender != "female" {
		t.Errorf("expected Gender 'female', got %q", r2.Gender)
	}
	if r2.Quantity != 2 {
		t.Errorf("expected Quantity 2, got %d", r2.Quantity)
	}
	if r2.IsSensitive {
		t.Errorf("Elk must NOT be flagged sensitive")
	}
	if math.Abs(r2.Latitude-46.5891) > 1e-4 {
		t.Errorf("expected lat ~46.5891, got %f", r2.Latitude)
	}
}

func TestMDTCarcassCSV_SensitiveSpeciesGeneralization(t *testing.T) {
	// Sample CSV with sensitive species: Grizzly Bear, Canada Lynx, Gray Wolf
	// at Bozeman benchmarks coordinates: Easting: 479836.847, Northing: 159775.029 (Raw: ~45.6770, -111.0429)
	csvData := `Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD
06/15/2023,US-191,35.0,Grizzly Bear,Male,1,479836.847,159775.029
08/20/2023,US-89,14.2,Canada Lynx,Female,1,479836.847,159775.029
11/04/2023,MT-83,55.8,Gray Wolf,Unknown,1,479836.847,159775.029
09/10/2023,US-191,42.0,Mule Deer,Female,1,479836.847,159775.029
`

	records, err := ParseMDTCarcassCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}

	// 1. Grizzly Bear
	grizzly := records[0]
	if !grizzly.IsSensitive {
		t.Errorf("Grizzly Bear must be flagged IsSensitive=true")
	}
	if grizzly.NormalizedSpecies != "grizzly_bear" {
		t.Errorf("expected normalized species 'grizzly_bear', got %q", grizzly.NormalizedSpecies)
	}
	if grizzly.H3R6 == "" {
		t.Fatalf("expected H3R6 index for grizzly bear, got empty")
	}
	// Compute expected H3-r6 cell and centroid for Bozeman coordinates:
	expectedCell, expectedCentLat, expectedCentLon, err := ComputeH3R6(grizzly.RawLatitude, grizzly.RawLongitude)
	if err != nil {
		t.Fatalf("failed to compute expected H3R6: %v", err)
	}
	if grizzly.H3R6 != expectedCell {
		t.Errorf("expected Bozeman H3 resolution 6 cell %q, got %q", expectedCell, grizzly.H3R6)
	}

	// Verify coordinates were generalized to the centroid of the H3-r6 cell
	if math.Abs(grizzly.Latitude-expectedCentLat) > 1e-6 {
		t.Errorf("generalized lat mismatch: got %f, want centroid %f", grizzly.Latitude, expectedCentLat)
	}
	if math.Abs(grizzly.Longitude-expectedCentLon) > 1e-6 {
		t.Errorf("generalized lon mismatch: got %f, want centroid %f", grizzly.Longitude, expectedCentLon)
	}

	// Verify raw coordinates are preserved in RawLatitude and RawLongitude
	if math.Abs(grizzly.RawLatitude-45.6770) > 1e-4 {
		t.Errorf("raw lat mismatch: got %f, want ~45.6770", grizzly.RawLatitude)
	}
	if math.Abs(grizzly.RawLongitude-(-111.0429)) > 1e-4 {
		t.Errorf("raw lon mismatch: got %f, want ~ -111.0429", grizzly.RawLongitude)
	}
	// Effective coordinates MUST differ from raw GPS to prevent poaching disclosure
	if grizzly.Latitude == grizzly.RawLatitude && grizzly.Longitude == grizzly.RawLongitude {
		t.Errorf("sensitive coordinates were NOT generalized (still equal to raw coordinates)")
	}

	// 2. Canada Lynx
	lynx := records[1]
	if !lynx.IsSensitive {
		t.Errorf("Canada Lynx must be flagged IsSensitive=true")
	}
	if lynx.NormalizedSpecies != "lynx" {
		t.Errorf("expected normalized species 'lynx', got %q", lynx.NormalizedSpecies)
	}
	if lynx.H3R6 != "86289230fffffff" {
		t.Errorf("expected H3 cell '86289230fffffff', got %q", lynx.H3R6)
	}
	if math.Abs(lynx.Latitude-expectedCentLat) > 1e-6 || math.Abs(lynx.Longitude-expectedCentLon) > 1e-6 {
		t.Errorf("lynx coordinates not generalized to H3 centroid")
	}

	// 3. Gray Wolf
	wolf := records[2]
	if !wolf.IsSensitive {
		t.Errorf("Gray Wolf must be flagged IsSensitive=true")
	}
	if wolf.NormalizedSpecies != "gray_wolf" {
		t.Errorf("expected normalized species 'gray_wolf', got %q", wolf.NormalizedSpecies)
	}
	if wolf.H3R6 != "86289230fffffff" {
		t.Errorf("expected H3 cell '86289230fffffff', got %q", wolf.H3R6)
	}
	if math.Abs(wolf.Latitude-expectedCentLat) > 1e-6 || math.Abs(wolf.Longitude-expectedCentLon) > 1e-6 {
		t.Errorf("wolf coordinates not generalized to H3 centroid")
	}

	// 4. Mule Deer (non-sensitive control in same CSV)
	muleDeer := records[3]
	if muleDeer.IsSensitive {
		t.Errorf("Mule Deer must NOT be flagged sensitive")
	}
	// For Mule deer, coordinates MUST match raw coordinates
	if muleDeer.Latitude != muleDeer.RawLatitude || muleDeer.Longitude != muleDeer.RawLongitude {
		t.Errorf("mule deer coordinates should not be blurred")
	}
}

func TestMDTCarcassCSV_BOMAndHeaderVariations(t *testing.T) {
	// Test UTF-8 BOM, snake_case headers, extra spaces, mixed case
	csvWithBOM := "\uFEFFcollection_date,route,milepost,animal_type,gender,quantity,sms_x_cord,sms_y_cord\n" +
		"2023-01-15,US 191,10.0,Elk,M,1,405557.683,263078.179\n"

	records, err := ParseMDTCarcassCSV(strings.NewReader(csvWithBOM))
	if err != nil {
		t.Fatalf("failed to parse CSV with BOM and snake_case headers: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Route != "US 191" {
		t.Errorf("expected Route 'US 191', got %q", records[0].Route)
	}
	if records[0].NormalizedSpecies != "elk" {
		t.Errorf("expected NormalizedSpecies 'elk', got %q", records[0].NormalizedSpecies)
	}
}

func TestMDTCarcassCSV_MissingRequiredHeaders(t *testing.T) {
	tests := []struct {
		name    string
		csvData string
	}{
		{
			name:    "Missing Route",
			csvData: "Collection Date,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD\n10/14/2021,12.5,Deer,M,1,405557.683,263078.179\n",
		},
		{
			name:    "Missing SMS_X_CORD",
			csvData: "Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_Y_CORD\n10/14/2021,US-191,12.5,Deer,M,1,263078.179\n",
		},
		{
			name:    "Missing SMS_Y_CORD",
			csvData: "Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD\n10/14/2021,US-191,12.5,Deer,M,1,405557.683\n",
		},
		{
			name:    "Missing Animal Type",
			csvData: "Collection Date,Route,Milepost,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD\n10/14/2021,US-191,12.5,M,1,405557.683,263078.179\n",
		},
		{
			name:    "Missing Collection Date",
			csvData: "Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD\nUS-191,12.5,Deer,M,1,405557.683,263078.179\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMDTCarcassCSV(strings.NewReader(tt.csvData))
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestMDTCarcassCSV_InvalidRows(t *testing.T) {
	// Empty CSV
	_, err := ParseMDTCarcassCSV(strings.NewReader(""))
	if err == nil {
		t.Error("expected error on empty CSV")
	}

	// Nil reader
	_, err = ParseMDTCarcassCSV(nil)
	if err == nil {
		t.Error("expected error on nil reader")
	}

	// Malformed coordinate
	badCoords := "Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD\n" +
		"10/14/2021,US-191,12.5,Deer,M,1,NOT_A_NUMBER,263078.179\n"
	_, err = ParseMDTCarcassCSV(strings.NewReader(badCoords))
	if err == nil {
		t.Error("expected error on non-numeric SMS_X_CORD")
	}

	// Malformed date
	badDate := "Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD\n" +
		"INVALID_DATE,US-191,12.5,Deer,M,1,405557.683,263078.179\n"
	_, err = ParseMDTCarcassCSV(strings.NewReader(badDate))
	if err == nil {
		t.Error("expected error on invalid date")
	}
}
