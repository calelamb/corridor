package ingest

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// syntheticNCDOTAnimalCrashJSON is an invented ArcGIS feature page.
// Coordinates are 1,2 and 3,4 so they cannot be read as a real county.
// GIS_Milepo covers a number, a blank string, null, an omitted field,
// whitespace, and an explicit zero. Explicit zero must stay zero.
const syntheticNCDOTAnimalCrashJSON = `{
  "features": [
    {
      "attributes": {
        "FID": 1,
        "Crash_ID": 900001,
        "Date": 1762041600000,
        "Time": "14:30",
        "Severity": "Property Damage",
        "CrashType": "ANIMAL",
        "CountyName": "SYNTHETIC COUNTY",
        "City": "SYN CITY",
        "OnRd": "SYNTHETIC RD",
        "Distance": 0.5,
        "Direction": "N",
        "FromRd": "SYN FROM",
        "TowardRd": "SYN TOWARD",
        "GIS_RteTxt": "SYN-ROUTE",
        "GIS_Milepo": 12.5,
        "Latitude": 1,
        "Longitude": 2,
        "Source_ID": "synthetic-src",
        "MapMethod": "SYN",
        "CrashRepor": "SYN-REPORT",
        "LOC_ERROR": "SYN-LOC",
        "NOT_A_LIVE_FIELD": "ignored"
      },
      "geometry": {"x": 500000, "y": 4000000}
    },
    {
      "attributes": {
        "FID": 2,
        "Crash_ID": "900002",
        "Date": "2025-06-15T12:00:00Z",
        "Time": "09:00",
        "CrashType": "ANIMAL",
        "CountyName": "SYNTHETIC COUNTY",
        "OnRd": "SYNTHETIC RD",
        "GIS_RteTxt": "SYN-ROUTE",
        "GIS_Milepo": "",
        "Latitude": 3,
        "Longitude": 4
      }
    },
    {
      "attributes": {
        "FID": 3,
        "Crash_ID": 900003,
        "Date": "1762041600000",
        "CrashType": "ANIMAL",
        "GIS_Milepo": null
      }
    },
    {
      "attributes": {
        "FID": 4,
        "Crash_ID": 900004,
        "Date": "2025-06-15",
        "CrashType": "ANIMAL"
      }
    },
    {
      "attributes": {
        "FID": 5,
        "Crash_ID": 900005,
        "CrashType": "ANIMAL",
        "GIS_Milepo": "   "
      }
    },
    {
      "attributes": {
        "FID": 6,
        "Crash_ID": 900006,
        "CrashType": "ANIMAL",
        "GIS_Milepo": 0,
        "Latitude": 1,
        "Longitude": 2
      }
    }
  ]
}`

func TestNCDOTAnimalCrashParseSynthetic(t *testing.T) {
	rows, err := ParseNCDOTAnimalCrashes(strings.NewReader(syntheticNCDOTAnimalCrashJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6 {
		t.Fatalf("rows = %d", len(rows))
	}

	first := rows[0]
	if first.FID != 1 || first.CrashID != "900001" {
		t.Fatalf("identity = fid %d crash %q", first.FID, first.CrashID)
	}
	if first.CrashType != "ANIMAL" {
		t.Fatalf("crash type = %q", first.CrashType)
	}
	wantDate := time.Date(2025, 11, 2, 0, 0, 0, 0, time.UTC)
	if !first.Date.Equal(wantDate) {
		t.Fatalf("epoch date = %s", first.Date)
	}
	if first.Time != "14:30" || first.Severity != "Property Damage" || first.OnRd != "SYNTHETIC RD" {
		t.Fatalf("road fields = %+v", first)
	}
	if first.CountyName != "SYNTHETIC COUNTY" || first.City != "SYN CITY" || first.Direction != "N" {
		t.Fatalf("place fields = %+v", first)
	}
	if first.Distance != "0.5" || first.FromRd != "SYN FROM" || first.TowardRd != "SYN TOWARD" {
		t.Fatalf("offset fields = %+v", first)
	}
	if first.GISRteTxt != "SYN-ROUTE" || first.SourceID != "synthetic-src" {
		t.Fatalf("route fields = %+v", first)
	}
	if first.MapMethod != "SYN" || first.CrashRepor != "SYN-REPORT" || first.LocError != "SYN-LOC" {
		t.Fatalf("method fields = %+v", first)
	}
	if first.GISMilepo == nil || *first.GISMilepo != 12.5 {
		t.Fatalf("milepost = %v", first.GISMilepo)
	}
	if first.Latitude == nil || first.Longitude == nil || *first.Latitude != 1 || *first.Longitude != 2 {
		t.Fatalf("point = lat %v lon %v", first.Latitude, first.Longitude)
	}

	stringDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if rows[1].CrashID != "900002" || !rows[1].Date.Equal(stringDate) {
		t.Fatalf("string date row = %+v", rows[1])
	}
	if rows[1].GISMilepo != nil {
		t.Fatalf("blank milepost string = %v", *rows[1].GISMilepo)
	}
	if rows[1].Latitude == nil || *rows[1].Latitude != 3 || *rows[1].Longitude != 4 {
		t.Fatalf("second point = %+v", rows[1])
	}

	if !rows[2].Date.Equal(wantDate) {
		t.Fatalf("epoch string date = %s", rows[2].Date)
	}
	if rows[2].GISMilepo != nil || rows[2].Latitude != nil || rows[2].Longitude != nil {
		t.Fatalf("null milepost row published a number or a point: %+v", rows[2])
	}

	dateOnly := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if !rows[3].Date.Equal(dateOnly) || rows[3].GISMilepo != nil {
		t.Fatalf("omitted milepost row = %+v", rows[3])
	}
	if rows[4].GISMilepo != nil {
		t.Fatalf("whitespace milepost = %v", *rows[4].GISMilepo)
	}
	if rows[5].GISMilepo == nil || *rows[5].GISMilepo != 0 {
		t.Fatalf("explicit zero milepost = %v", rows[5].GISMilepo)
	}

	pub := first.PublicRecord()
	body, err := json.Marshal(pub)
	if err != nil {
		t.Fatal(err)
	}
	assertNCDOTPublicCrash(t, body, 1, 2)
	if !strings.Contains(string(body), `"gis_milepo":12.5`) {
		t.Fatalf("public JSON dropped the reported milepost: %s", body)
	}

	blankBody, err := json.Marshal(rows[1].PublicRecord())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blankBody), "gis_milepo") {
		t.Fatalf("blank milepost was published as a number: %s", blankBody)
	}
	assertNCDOTPublicCrash(t, blankBody, 3, 4)

	zeroBody, err := json.Marshal(rows[5].PublicRecord())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(zeroBody), `"gis_milepo":0`) {
		t.Fatalf("explicit zero milepost was omitted: %s", zeroBody)
	}
}

func TestNCDOTAnimalCrashPropertiesPage(t *testing.T) {
	page := `{
	  "type": "FeatureCollection",
	  "features": [
	    {"properties": {"FID": 7, "CrashType": "ANIMAL", "Date": "6/15/2025", "GIS_Milepo": "8.5", "Latitude": 1, "Longitude": 2}}
	  ]
	}`
	rows, err := ParseNCDOTAnimalCrashes(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].CrashType != "ANIMAL" || rows[0].FID != 7 {
		t.Fatalf("row = %+v", rows)
	}
	want := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if !rows[0].Date.Equal(want) {
		t.Fatalf("us date = %s", rows[0].Date)
	}
	if rows[0].GISMilepo == nil || *rows[0].GISMilepo != 8.5 {
		t.Fatalf("string milepost = %v", rows[0].GISMilepo)
	}
}

func TestNCDOTAnimalCrashArtifactUnresolved(t *testing.T) {
	a := NCDOTAnimalCrashesArtifact()
	if a.URL != NCDOTAnimalCrashes2025URL {
		t.Fatalf("url = %q", a.URL)
	}
	if a.License != "license_unresolved" {
		t.Fatalf("license = %q", a.License)
	}
	if a.SHA256 != "" || a.ObjectKey != "" {
		t.Fatal("artifact records a hash or object key without an acquired file")
	}
	want := time.Date(2026, 11, 3, 0, 41, 30, 0, time.UTC)
	if !a.RetrievedAt.Equal(want) {
		t.Fatalf("retrieved = %s", a.RetrievedAt)
	}
	if strings.Contains(strings.ToLower(a.Name), "carcass") {
		t.Fatalf("name labels a carcass inventory: %q", a.Name)
	}
	if a.SourceID == "" {
		t.Fatal("missing source id")
	}
}

func TestNCDOTAnimalCrashRejectsInvalidJSON(t *testing.T) {
	cases := []string{
		"",
		"{",
		`{"error":{"code":499,"message":"Token Required"}}`,
		`{"features":null}`,
		`{"features":[{"attributes":{"FID":1,"GIS_Milepo":"not-a-mile","CrashType":"ANIMAL"}}]}`,
		`{"features":[{"attributes":{"FID":1,"Date":"not-a-date","CrashType":"ANIMAL"}}]}`,
		`{"features":[{"nope":true}]}`,
	}
	for _, body := range cases {
		if _, err := ParseNCDOTAnimalCrashes(strings.NewReader(body)); err == nil {
			t.Fatalf("expected error for %s", body)
		}
	}
	if _, err := ParseNCDOTAnimalCrashes(nil); err == nil {
		t.Fatal("expected nil reader error")
	}
	rows, err := ParseNCDOTAnimalCrashes(strings.NewReader(`{"features":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("empty page = %d", len(rows))
	}
}

func assertNCDOTPublicCrash(t *testing.T, body []byte, lat, lon float64) {
	t.Helper()
	text := strings.ToLower(string(body))
	if strings.Contains(text, "deer") || strings.Contains(text, "carcass") || strings.Contains(text, "species") {
		t.Fatalf("public JSON invented a species or a carcass inventory: %s", body)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["crash_type"] != "ANIMAL" {
		t.Fatalf("crash type = %#v", payload["crash_type"])
	}
	if payload["latitude"] != lat || payload["longitude"] != lon {
		t.Fatalf("public point = %s", body)
	}
}
