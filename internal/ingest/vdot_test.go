package ingest

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

// Synthetic ArcGIS JSON. Coordinates and identifiers are invented for the
// test and are not values returned by the live VDOT service. Route strings
// are coded domain values from the layer metadata.
const syntheticVDOTCarcassJSON = `{
  "features": [
    {
      "attributes": {
        "objectid": 501,
        "globalid": "{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}",
        "starttime": 1700000000000,
        "wildlife": "Deer",
        "wildlifeother": null,
        "rtenamedir": "I-81N",
        "latitude_d": 38.5,
        "longitude_d": -78.5,
        "Jurisdiction": "Synthetic",
        "District": "Staunton",
        "Residency": "Synthetic"
      }
    },
    {
      "attributes": {
        "objectid": 502,
        "globalid": "{bbbbbbbb-bbbb-cccc-dddd-eeeeeeeeeeee}",
        "starttime": 1700003600000,
        "wildlife": "Bear",
        "wildlifeother": "",
        "rtenamedir": "I-64E",
        "latitude_d": 38.25,
        "longitude_d": -79.125,
        "Jurisdiction": "Synthetic",
        "District": "Culpeper",
        "Residency": "Synthetic"
      }
    },
    {
      "attributes": {
        "objectid": 503,
        "globalid": "{cccccccc-bbbb-cccc-dddd-eeeeeeeeeeee}",
        "starttime": 1700007200000,
        "wildlife": "Other",
        "wildlifeother": "Bobcat",
        "rtenamedir": "I-66W",
        "latitude_d": 37.75,
        "longitude_d": -77.5,
        "Jurisdiction": "Synthetic",
        "District": "Richmond",
        "Residency": "Synthetic"
      }
    },
    {
      "attributes": {
        "objectid": 504,
        "globalid": "{dddddddd-bbbb-cccc-dddd-eeeeeeeeeeee}",
        "starttime": 1700010800000,
        "wildlife": "Notfound",
        "wildlifeother": null,
        "rtenamedir": "Rte37E",
        "latitude_d": 37.5,
        "longitude_d": -78.25,
        "Jurisdiction": "Synthetic",
        "District": "Hampton Roads",
        "Residency": "Synthetic"
      }
    }
  ]
}`

func TestVDOTCarcassSyntheticPrivacy(t *testing.T) {
	records, err := ParseVDOTCarcass(strings.NewReader(syntheticVDOTCarcassJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("record count = %d, want 4", len(records))
	}

	deer := records[0]
	if deer.ObjectID != 501 || deer.RouteNameDir != "I-81N" || deer.District != "Staunton" {
		t.Fatalf("deer identity = id %d route %q district %q", deer.ObjectID, deer.RouteNameDir, deer.District)
	}
	if deer.Wildlife != "Deer" || deer.NormalizedSpecies != "deer" || deer.IsSensitive {
		t.Fatalf("deer species = %q normalized %q sensitive %v", deer.Wildlife, deer.NormalizedSpecies, deer.IsSensitive)
	}
	if deer.Latitude != 38.5 || deer.Longitude != -78.5 {
		t.Fatalf("deer published coordinates changed: lat %v lon %v", deer.Latitude, deer.Longitude)
	}
	if deer.H3R6 != "" || !deer.PublicAfter.IsZero() {
		t.Fatalf("deer was generalized: h3 %q public-after %s", deer.H3R6, deer.PublicAfter)
	}
	deerBody, err := json.Marshal(deer)
	if err != nil {
		t.Fatal(err)
	}
	var deerPayload map[string]any
	if err := json.Unmarshal(deerBody, &deerPayload); err != nil {
		t.Fatal(err)
	}
	deerLon, ok := deerPayload["longitude"].(float64)
	if !ok || deerLon != -78.5 {
		t.Fatalf("deer JSON dropped the synthetic longitude: %s", deerBody)
	}
	if _, ok := deer.PublicRecord(deer.StartTime); !ok {
		t.Fatal("deer was withheld at its start time")
	}

	bear := records[1]
	if bear.Wildlife != "Bear" || bear.NormalizedSpecies != "black_bear" || !bear.IsSensitive {
		t.Fatalf("bear species = %q normalized %q sensitive %v", bear.Wildlife, bear.NormalizedSpecies, bear.IsSensitive)
	}
	if bear.RouteNameDir != "I-64E" || bear.District != "Culpeper" {
		t.Fatalf("bear route = %q district %q", bear.RouteNameDir, bear.District)
	}
	if bear.RawLatitude != 38.25 || bear.RawLongitude != -79.125 {
		t.Fatalf("bear raw point = %v, %v", bear.RawLatitude, bear.RawLongitude)
	}
	if bear.Latitude == bear.RawLatitude && bear.Longitude == bear.RawLongitude {
		t.Fatal("bear published the synthetic input coordinates")
	}
	cell, centLat, centLon, err := ComputeH3R6(bear.RawLatitude, bear.RawLongitude)
	if err != nil {
		t.Fatalf("h3: %v", err)
	}
	if bear.H3R6 != cell {
		t.Fatalf("bear h3 = %q, want %q", bear.H3R6, cell)
	}
	if math.Abs(bear.Latitude-centLat) > 1e-8 || math.Abs(bear.Longitude-centLon) > 1e-8 {
		t.Fatal("bear published point is not the H3 resolution 6 centroid")
	}
	wantHold := bear.StartTime.Add(SensitiveQuarantineDuration)
	if !bear.PublicAfter.Equal(wantHold) {
		t.Fatalf("public-after = %s, want %s", bear.PublicAfter, wantHold)
	}
	bearBody, err := json.Marshal(bear)
	if err != nil {
		t.Fatal(err)
	}
	var bearPayload map[string]any
	if err := json.Unmarshal(bearBody, &bearPayload); err != nil {
		t.Fatal(err)
	}
	if _, present := bearPayload["raw_longitude"]; present {
		t.Fatalf("bear JSON included raw_longitude: %s", bearBody)
	}
	for key, value := range bearPayload {
		number, isNumber := value.(float64)
		if !isNumber {
			continue
		}
		if math.Abs(number-bear.RawLongitude) < 1e-8 {
			t.Fatalf("bear JSON field %s marshaled the input longitude: %s", key, bearBody)
		}
	}
	if _, ok := bear.PublicRecord(bear.StartTime); ok {
		t.Fatal("PublicRecord returned a bear record at its start time")
	}
	pub, ok := bear.PublicRecord(bear.PublicAfter)
	if !ok {
		t.Fatal("bear stayed withheld at the end of the 30-day delay")
	}
	if pub.RawLongitude != 0 || math.Abs(pub.Longitude-bear.RawLongitude) < 1e-8 {
		t.Fatal("public bear kept the reported longitude")
	}

	bobcat := records[2]
	if bobcat.NormalizedSpecies != "bobcat" || bobcat.IsSensitive || bobcat.WildlifeOther != "Bobcat" {
		t.Fatalf("bobcat = species %q other %q sensitive %v", bobcat.NormalizedSpecies, bobcat.WildlifeOther, bobcat.IsSensitive)
	}
	if bobcat.Latitude != 37.75 || bobcat.Longitude != -77.5 {
		t.Fatal("bobcat published coordinates changed")
	}

	missing := records[3]
	if missing.Wildlife != "Notfound" || missing.NormalizedSpecies != "not_found" || missing.IsSensitive {
		t.Fatalf("not-found = %q normalized %q sensitive %v", missing.Wildlife, missing.NormalizedSpecies, missing.IsSensitive)
	}
	if missing.Latitude != 37.5 || missing.Longitude != -78.25 || missing.RouteNameDir != "Rte37E" {
		t.Fatalf("not-found point changed: route %q lat %v lon %v", missing.RouteNameDir, missing.Latitude, missing.Longitude)
	}
}

func TestVDOTWildlifeCodes(t *testing.T) {
	codes := []struct {
		wildlife  string
		other     string
		want      string
		sensitive bool
	}{
		{wildlife: "Deer", want: "deer"},
		{wildlife: "Bear", want: "black_bear", sensitive: true},
		{wildlife: "NotFound", want: "not_found"},
		{wildlife: "Other", other: "Coyote", want: "coyote"},
		{wildlife: "Other", other: "Raccoon", want: "raccoon"},
		{wildlife: "Other", other: "Bird", want: "bird"},
		{wildlife: "Other", other: "Bobcat", want: "bobcat"},
		{wildlife: "Other", other: "Squirrel", want: "squirrel"},
		{wildlife: "Other", other: "Fox", want: "fox"},
		{wildlife: "Other", other: "Skunk", want: "skunk"},
		{wildlife: "Other", other: "Turtle", want: "turtle"},
		{wildlife: "Other", other: "Opossum", want: "opossum"},
		{wildlife: "Other", other: "Groundhog", want: "groundhog"},
		{wildlife: "Other", other: "Dog", want: "dog"},
		{wildlife: "Other", other: "Cat", want: "cat"},
		{wildlife: "Other", other: "Farmanimal", want: "farm_animal"},
		{wildlife: "Other", other: "Otheranimal", want: "other_animal"},
	}
	for _, code := range codes {
		got := NormalizeVDOTSpecies(code.wildlife, code.other)
		if got != code.want || IsVDOTSensitiveSpecies(got) != code.sensitive {
			t.Fatalf("wildlife %q other %q = %q sensitive %v, want %q sensitive %v",
				code.wildlife, code.other, got, IsVDOTSensitiveSpecies(got), code.want, code.sensitive)
		}
	}
}

func TestVDOTCarcassRejectsServiceError(t *testing.T) {
	body := `{"error":{"code":400,"message":"Unable to complete operation."}}`
	if _, err := ParseVDOTCarcass(strings.NewReader(body)); err == nil {
		t.Fatal("expected an error for an ArcGIS error payload")
	}
}

func TestVDOTCarcassBearNeedsCoordinates(t *testing.T) {
	body := `{"features":[{"attributes":{"objectid":9,"wildlife":"Bear","starttime":1700000000000,"latitude_d":null,"longitude_d":null}}]}`
	if _, err := ParseVDOTCarcass(strings.NewReader(body)); err == nil {
		t.Fatal("expected an error for a bear row without coordinates")
	}
}

func TestVDOTCarcassArtifactUnresolvedLicense(t *testing.T) {
	a := VDOTCarcassArtifact()
	if a.URL != vdotCarcassLayerURL {
		t.Fatalf("url = %q", a.URL)
	}
	if a.License != "license_unresolved" {
		t.Fatalf("license = %q", a.License)
	}
	if a.SHA256 != "" || a.ObjectKey != "" {
		t.Fatal("artifact records a hash or object key without an acquired file")
	}
	if !a.RetrievedAt.Equal(time.Date(2026, 11, 3, 0, 39, 59, 0, time.UTC)) {
		t.Fatalf("retrieved = %s", a.RetrievedAt)
	}
	if a.SourceID == "" || a.Name == "" {
		t.Fatalf("artifact identity is empty: %+v", a)
	}
}

// vdotRow builds one synthetic feature. Values are invented for tests.
func vdotRow(wildlife, other, lat, lon, start string) string {
	return `{"features":[{"attributes":{"objectid":7,"starttime":` + start +
		`,"wildlife":` + wildlife + `,"wildlifeother":` + other +
		`,"latitude_d":` + lat + `,"longitude_d":` + lon + `}}]}`
}

func TestVDOTBearInFreeTextIsSensitive(t *testing.T) {
	for _, other := range []string{`"Black Bear"`, `"Bear cub"`, `"bear"`} {
		body := vdotRow(`"Other"`, other, "38.4", "-78.6", "1700000000000")
		records, err := ParseVDOTCarcass(strings.NewReader(body))
		if err != nil {
			t.Fatalf("%s: %v", other, err)
		}
		if !records[0].IsSensitive || records[0].H3R6 == "" || records[0].Longitude == -78.6 {
			t.Fatalf("%s was not generalized: %+v", other, records[0])
		}
	}
	if IsVDOTSensitiveSpecies("bearded_dragon") {
		t.Fatal("token match flagged bearded_dragon")
	}
}

func TestVDOTPublicRecordBoundary(t *testing.T) {
	body := vdotRow(`"Bear"`, "null", "38.4", "-78.6", "1700000000000")
	records, err := ParseVDOTCarcass(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	bear := records[0]
	release := bear.StartTime.Add(30 * 24 * time.Hour)
	if _, ok := bear.PublicRecord(release.Add(-time.Nanosecond)); ok {
		t.Fatal("released before starttime plus 30 days")
	}
	if _, ok := bear.PublicRecord(release); !ok {
		t.Fatal("withheld at starttime plus 30 days")
	}
	// A tampered PublicAfter must not shorten the delay.
	bear.PublicAfter = bear.StartTime
	if _, ok := bear.PublicRecord(bear.StartTime.Add(time.Hour)); ok {
		t.Fatal("PublicRecord trusted a shortened PublicAfter")
	}
}

func TestVDOTSensitiveFailsClosed(t *testing.T) {
	ref := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := map[string]VDOTCarcassRecord{
		"unknown start":  {IsSensitive: true, H3R6: "x"},
		"not generalize": {IsSensitive: true, StartTime: ref.Add(-365 * 24 * time.Hour)},
	}
	for name, rec := range cases {
		if _, ok := rec.PublicRecord(ref); ok {
			t.Fatalf("%s: sensitive record was published", name)
		}
	}
}

func TestVDOTDeerWithoutStartOrCoordinates(t *testing.T) {
	body := vdotRow(`"Deer"`, "null", "null", "null", "null")
	records, err := ParseVDOTCarcass(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := records[0].PublicRecord(time.Time{}); !ok {
		t.Fatal("deer withheld")
	}
	out, _ := json.Marshal(records[0])
	if strings.Contains(string(out), "public_after") {
		t.Fatalf("zero public_after serialized: %s", out)
	}
}

func TestVDOTRejectsBadCoordinates(t *testing.T) {
	cases := [][2]string{
		{"38.4", "null"}, {"null", "-78.6"}, {"95", "-78.6"}, {"38.4", "-181"}, {`"NaN"`, "-78.6"},
	}
	for _, wildlife := range []string{`"Deer"`, `"Bear"`} {
		for _, c := range cases {
			body := vdotRow(wildlife, "null", c[0], c[1], "1700000000000")
			if _, err := ParseVDOTCarcass(strings.NewReader(body)); err == nil {
				t.Fatalf("%s lat %s lon %s: expected error", wildlife, c[0], c[1])
			}
		}
	}
}

func TestVDOTAttributeNamesCaseInsensitive(t *testing.T) {
	body := `{"features":[{"attributes":{"OBJECTID":3,"STARTTIME":1700000000000,"WILDLIFE":"Deer",` +
		`"jurisdiction":"Augusta","district":"Staunton","residency":"Verona","LATITUDE_D":38.1,"LONGITUDE_D":-78.9}}]}`
	records, err := ParseVDOTCarcass(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	r := records[0]
	if r.ObjectID != 3 || r.Jurisdiction != "Augusta" || r.District != "Staunton" || r.Residency != "Verona" || r.Latitude != 38.1 {
		t.Fatalf("case-insensitive lookup failed: %+v", r)
	}
}

func TestVDOTArtifactDoesNotNameUncoveredRoutes(t *testing.T) {
	src, err := os.ReadFile("vdot.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, banned := range []string{"Skyline", "US-33", "US-211"} {
		if strings.Contains(string(src), banned) {
			t.Fatalf("vdot.go mentions %s", banned)
		}
	}
}
