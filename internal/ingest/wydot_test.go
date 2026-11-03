package ingest

import (
	"math"
	"strings"
	"testing"
	"time"
)

// Synthetic ArcGIS JSON. Coordinates are invented for the test and are not
// values returned by the live WYDOT service.
const syntheticWYDOTRoadkillJSON = `{
  "features": [
    {
      "attributes": {
        "pk": 1001,
        "received_ts": 1700000000000,
        "status": "received",
        "location": "synthetic marker",
        "route": "ML80B",
        "highway": "I-80",
        "milepost": 12.5,
        "cardinal_direction": "E",
        "county": "Albany",
        "district": 1,
        "longitude": -107.0,
        "latitude": 43.0,
        "species": "DEER",
        "other_species": "",
        "prohibited_species": 0,
        "collected": 1,
        "road_blockage": 0,
        "in_excluded_zone": 0
      }
    },
    {
      "attributes": {
        "pk": 1002,
        "received_ts": 1700003600000,
        "status": "received",
        "location": "synthetic marker",
        "route": "US191",
        "highway": "US-191",
        "milepost": 40,
        "cardinal_direction": "N",
        "county": "Teton",
        "district": 3,
        "longitude": -109.25,
        "latitude": 44.25,
        "species": "GRIZZLY BEAR",
        "other_species": null,
        "prohibited_species": 1,
        "collected": 0,
        "road_blockage": 0,
        "in_excluded_zone": 1
      }
    },
    {
      "attributes": {
        "pk": 1003,
        "received_ts": 1700007200000,
        "status": "received",
        "location": "synthetic marker",
        "route": "US26",
        "highway": "US-26",
        "milepost": 8.25,
        "cardinal_direction": "W",
        "county": "Fremont",
        "district": 5,
        "longitude": -110.5,
        "latitude": 43.75,
        "species": "WOLF",
        "other_species": "",
        "prohibited_species": 1,
        "collected": 0,
        "road_blockage": 1,
        "in_excluded_zone": 0
      }
    },
    {
      "attributes": {
        "pk": 1004,
        "received_ts": 1700010800000,
        "status": "received",
        "location": "synthetic marker",
        "route": "US189",
        "highway": "US-189",
        "milepost": 22,
        "cardinal_direction": "S",
        "county": "Sublette",
        "district": 3,
        "longitude": -108.1,
        "latitude": 42.8,
        "species": "MOUNTAIN LION",
        "other_species": "",
        "prohibited_species": 0,
        "collected": 0,
        "road_blockage": 0,
        "in_excluded_zone": 0
      }
    }
  ]
}`

func TestParseWYDOTRoadkillSyntheticSchema(t *testing.T) {
	records, err := ParseWYDOTRoadkill(strings.NewReader(syntheticWYDOTRoadkillJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("record count = %d, want 4", len(records))
	}

	deer := records[0]
	if deer.PK != 1001 || deer.Highway != "I-80" || deer.Route != "ML80B" {
		t.Fatalf("deer identity = pk %d highway %q route %q", deer.PK, deer.Highway, deer.Route)
	}
	if deer.Milepost != 12.5 || deer.District != 1 || deer.Collected != 1 {
		t.Fatalf("deer measures = milepost %v district %d collected %d", deer.Milepost, deer.District, deer.Collected)
	}
	if deer.NormalizedSpecies != "deer" || deer.IsSensitive {
		t.Fatalf("deer species = %q sensitive %v", deer.NormalizedSpecies, deer.IsSensitive)
	}
	if deer.Latitude != 43.0 || deer.Longitude != -107.0 {
		t.Fatalf("deer published coordinates changed")
	}
	if !deer.PublicAfter.IsZero() {
		t.Fatalf("non-sensitive record has a public delay")
	}
	wantDeerTime := time.UnixMilli(1700000000000).UTC()
	if !deer.Received.Equal(wantDeerTime) {
		t.Fatalf("received = %s, want %s", deer.Received, wantDeerTime)
	}

	for _, rec := range records[1:] {
		if !rec.IsSensitive {
			t.Fatalf("pk %d species %q should be sensitive", rec.PK, rec.Species)
		}
		if rec.Latitude == rec.RawLatitude && rec.Longitude == rec.RawLongitude {
			t.Fatalf("pk %d published the synthetic input coordinates", rec.PK)
		}
		cell, centLat, centLon, err := ComputeH3R6(rec.RawLatitude, rec.RawLongitude)
		if err != nil {
			t.Fatalf("h3: %v", err)
		}
		if rec.H3R6 != cell {
			t.Fatalf("pk %d h3 = %q, want %q", rec.PK, rec.H3R6, cell)
		}
		if math.Abs(rec.Latitude-centLat) > 1e-8 || math.Abs(rec.Longitude-centLon) > 1e-8 {
			t.Fatalf("pk %d published point is not the H3 resolution 6 centroid", rec.PK)
		}
		wantHold := rec.Received.Add(wydotSensitivePublicDelay)
		if !rec.PublicAfter.Equal(wantHold) {
			t.Fatalf("pk %d public-after = %s, want %s", rec.PK, rec.PublicAfter, wantHold)
		}
	}

	if records[1].NormalizedSpecies != "grizzly_bear" || records[1].ProhibitedSpecies != 1 || records[1].InExcludedZone != 1 {
		t.Fatalf("grizzly flags = species %q prohibited %d excluded %d", records[1].NormalizedSpecies, records[1].ProhibitedSpecies, records[1].InExcludedZone)
	}
	if records[2].NormalizedSpecies != "gray_wolf" || records[2].RoadBlockage != 1 {
		t.Fatalf("wolf = species %q blockage %d", records[2].NormalizedSpecies, records[2].RoadBlockage)
	}
	if records[3].NormalizedSpecies != "mountain_lion" {
		t.Fatalf("lion species = %q", records[3].NormalizedSpecies)
	}
}

func TestParseWYDOTRoadkillRejectsServiceError(t *testing.T) {
	body := `{"error":{"code":400,"message":"Unable to complete operation."}}`
	if _, err := ParseWYDOTRoadkill(strings.NewReader(body)); err == nil {
		t.Fatal("expected an error for an ArcGIS error payload")
	}
}

func TestWYDOTRoadkillArtifactUnresolvedLicense(t *testing.T) {
	a := WYDOTRoadkillArtifact()
	if a.URL != wydotRoadkillLayerURL {
		t.Fatalf("url = %q", a.URL)
	}
	if a.License != "license_unresolved" {
		t.Fatalf("license = %q", a.License)
	}
	if !a.RetrievedAt.Equal(time.Date(2026, 11, 3, 0, 38, 2, 0, time.UTC)) {
		t.Fatalf("retrieved = %s", a.RetrievedAt)
	}
	if a.SHA256 != "" || a.ObjectKey != "" {
		t.Fatal("artifact records a hash or object key without an acquired file")
	}
}
