package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"corridor/internal/geo"
)

func TestMDTCLI(t *testing.T) {
	const sensitiveEasting = 479836.847
	const sensitiveNorthing = 159775.029
	rawLat, rawLon, err := geo.StatePlane32100ToWGS84(sensitiveEasting, sensitiveNorthing)
	if err != nil {
		t.Fatalf("project sensitive fixture: %v", err)
	}

	csvData := `Collection Date,Route,Milepost,Animal Type,Gender,Quantity,SMS_X_CORD,SMS_Y_CORD
2020-06-15,us 191,35.0,Grizzly Bear,Male,1,479836.847,159775.029
2020-06-16,US-191,12.5,White-tailed Deer,Female,2,405557.683,263078.179
2020-07-01,I 90,10.0,Elk,Unknown,1,677679.089,170897.584
2020-08-02,us-191,40.0,Wolverine,Unknown,1,479836.847,159775.029
`
	path := filepath.Join(t.TempDir(), "mdt-synthetic.csv")
	if err := os.WriteFile(path, []byte(csvData), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = Run(context.Background(), []string{"ingest", "mdt", "--file", path, "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("ingest mdt: %v\nstderr: %s", err, stderr.String())
	}

	var got struct {
		Source        string           `json:"source"`
		RecordsRead   int              `json:"records_read"`
		Accepted      int              `json:"accepted"`
		TotalAnimals  int64            `json:"total_animals"`
		UniqueRoutes  int              `json:"unique_routes"`
		SpeciesCounts map[string]int64 `json:"species_counts"`
		SpatialRef    string           `json:"spatial_reference"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout.String())
	}

	if got.Source != "Montana Department of Transportation Statewide Carcass Collection" {
		t.Errorf("source = %q", got.Source)
	}
	if got.RecordsRead != 4 || got.Accepted != 4 {
		t.Errorf("records_read=%d accepted=%d, want 4 accepted records", got.RecordsRead, got.Accepted)
	}
	if got.TotalAnimals != 5 {
		t.Errorf("total_animals = %d, want 5", got.TotalAnimals)
	}
	// "us 191", "US-191", and "us-191" are one route; "I 90" is the second.
	if got.UniqueRoutes != 2 {
		t.Errorf("unique_routes = %d, want 2 after route normalization", got.UniqueRoutes)
	}
	if got.SpeciesCounts["grizzly_bear"] != 1 || got.SpeciesCounts["wolverine"] != 1 {
		t.Errorf("sensitive species counts = %#v", got.SpeciesCounts)
	}
	if got.SpeciesCounts["white_tailed_deer"] != 2 || got.SpeciesCounts["elk"] != 1 {
		t.Errorf("species counts = %#v", got.SpeciesCounts)
	}
	if !strings.Contains(got.SpatialRef, "EPSG:4326") {
		t.Errorf("spatial_reference = %q", got.SpatialRef)
	}

	body := stdout.String()
	secrets := []string{
		"479836.847",
		"159775.029",
		strconv.FormatFloat(rawLat, 'f', 5, 64),
		strconv.FormatFloat(rawLon, 'f', 5, 64),
		"raw_latitude",
		"raw_longitude",
		"sms_x_cord",
	}
	for _, secret := range secrets {
		if strings.Contains(body, secret) {
			t.Errorf("JSON stdout leaked precise coordinate %q:\n%s", secret, body)
		}
	}
}
