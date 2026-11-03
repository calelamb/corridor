package ingest

import (
	"strings"
	"testing"
	"time"
)

// syntheticUDOTCrashJSON is not a downloaded crash. Values are invented.
// Field names are the ones returned by Crash Locations layer 14 ?f=pjson.
const syntheticUDOTCrashJSON = `{
  "features": [
    {
      "attributes": {
        "OBJECTID": 1,
        "CRASH_ID": 900001,
        "CRASH_DATETIME": 1762041600000,
        "COUNTY_NAME": "SYNTHETIC COUNTY",
        "REGION_NAME": "SYNTHETIC REGION",
        "MAIN_ROAD_NAME": "SYNTHETIC ROAD",
        "ROUTE": "0099",
        "START_ACCUM": 12.5,
        "ROUTE_DIRECTION": "P",
        "LAT_UTM_Y": 2,
        "LONG_UTM_X": 1,
        "CRASH_SEVERITY_DESC": "No Injury",
        "FIRST_HARMFUL_EVENT_DESC": "Wild Animal",
        "WILD_ANIMAL_RELATED": "Y",
        "DOMESTIC_ANIMAL_RELATED": "N",
        "CURRENT_AS_OF_DATE": 1762128000000,
        "NOT_A_LIVE_FIELD": "ignored"
      },
      "geometry": {"x": 1, "y": 2}
    },
    {
      "attributes": {
        "OBJECTID": 2,
        "CRASH_ID": 900002,
        "ROUTE": "0099",
        "WILD_ANIMAL_RELATED": "N",
        "DOMESTIC_ANIMAL_RELATED": "Y"
      }
    }
  ]
}`

func TestUDOTCrashParseSynthetic(t *testing.T) {
	rows, err := ParseUDOTCrashes(strings.NewReader(syntheticUDOTCrashJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	wild := rows[0]
	if !wild.WildAnimal() {
		t.Fatal("synthetic Y row was not wild-animal")
	}
	if wild.Route != "0099" || wild.LongUTMX != 1 || wild.LatUTMY != 2 {
		t.Fatalf("attributes = %+v", wild)
	}
	if !wild.HasGeometry || wild.GeometryX != 1 || wild.GeometryY != 2 {
		t.Fatalf("geometry = %+v", wild)
	}
	want := time.Date(2025, 11, 2, 0, 0, 0, 0, time.UTC)
	if !wild.CrashDatetime.Equal(want) {
		t.Fatalf("crash time = %s", wild.CrashDatetime)
	}
	if rows[1].WildAnimal() {
		t.Fatal("synthetic N row was treated as wild-animal")
	}
	if rows[1].HasGeometry {
		t.Fatal("second synthetic row has no geometry")
	}
	if got := UDOTCrashLocationsArtifact(); got.URL != UDOTCrashLocationsURL || got.SHA256 != "" {
		t.Fatalf("artifact = %+v", got)
	}
}

func TestUDOTCrashRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseUDOTCrashes(strings.NewReader(`{"error":{"code":499}}`)); err == nil {
		t.Fatal("expected service error")
	}
	if _, err := ParseUDOTCrashes(strings.NewReader("")); err == nil {
		t.Fatal("expected empty source error")
	}
	if _, err := ParseUDOTCrashes(nil); err == nil {
		t.Fatal("expected nil reader error")
	}
}
