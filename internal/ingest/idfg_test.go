package ingest

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

const syntheticIDFGRoadkillJSON = `{
  "features": [
    {
      "attributes": {
        "OBJECTID": 1,
        "species": "Odocoileus hemionus",
        "source": "synthetic",
        "observed": 1700000000000,
        "highway": "US-20",
        "milepost": 401.2,
        "county": "Fremont",
        "latitude": 44.4,
        "longitude": -111.3
      }
    },
    {
      "attributes": {
        "OBJECTID": 2,
        "species": "GRAY WOLF",
        "source": "synthetic",
        "observed": 1700000000000,
        "highway": "US-20",
        "milepost": 402,
        "county": "Fremont",
        "latitude": 44.41,
        "longitude": -111.35
      }
    }
  ]
}`

func TestParseIDFGRoadkillSyntheticSchema(t *testing.T) {
	records, err := ParseIDFGRoadkill(strings.NewReader(syntheticIDFGRoadkillJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("count = %d", len(records))
	}
	deer := records[0]
	if deer.NormalizedSpecies != "mule_deer" || deer.IsSensitive || deer.Highway != "US-20" {
		t.Fatalf("deer = %+v", deer)
	}
	if deer.Latitude != 44.4 || deer.Longitude != -111.3 {
		t.Fatalf("deer coordinates changed")
	}
	wolf := records[1]
	if wolf.NormalizedSpecies != "gray_wolf" || !wolf.IsSensitive {
		t.Fatalf("wolf = %+v", wolf)
	}
	if math.Abs(wolf.Latitude-44.41) < 1e-6 && math.Abs(wolf.Longitude-(-111.35)) < 1e-6 {
		t.Fatal("wolf published the reported coordinates")
	}
	body, err := json.Marshal(wolf)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "raw_longitude") || strings.Contains(string(body), "-111.35") {
		t.Fatalf("JSON leaked the reported longitude: %s", body)
	}
	if _, ok := wolf.PublicRecord(wolf.Observed); ok {
		t.Fatal("wolf was public on the observation day")
	}
	if _, ok := wolf.PublicRecord(wolf.PublicAfter); !ok {
		t.Fatal("wolf stayed withheld after the delay")
	}
}

func TestIDFGRoadkillArtifactUnresolved(t *testing.T) {
	a := IDFGRoadkillArtifact()
	if a.SHA256 != "" || a.License != "license_unresolved" {
		t.Fatalf("artifact claims a cleared file: %+v", a)
	}
	if !a.RetrievedAt.Equal(time.Date(2026, 11, 3, 0, 38, 2, 0, time.UTC)) {
		t.Fatalf("retrieved = %s", a.RetrievedAt)
	}
}
