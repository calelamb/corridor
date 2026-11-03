// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// UDOT crash locations MapServer. Layer metadata checked 2026-11-03T00:38:00Z.
// Each year is a separate layer. This URL is the service root, not a carcass feed.
const UDOTCrashLocationsURL = "https://maps.udot.utah.gov/central/rest/services/TrafficAndSafety/Crash_Locations/MapServer"

// UDOTCrashLocationsArtifact is provenance for the public crash MapServer.
// No bulk extract was downloaded, so the checksum and object key are empty.
// The license text is the item disclaimer; redistribution remains unresolved.
func UDOTCrashLocationsArtifact() Artifact {
	return Artifact{
		SourceID:    "06c397a4-c162-42e4-bc29-28053f856def",
		Name:        "UDOT · Crash Locations",
		URL:         UDOTCrashLocationsURL,
		License:     "license_unresolved · UDOT disclaimer: informational use, field-verify before project use, no warranty · udotgis@utah.gov",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 38, 0, 0, time.UTC),
	}
}

// UDOTCrashRecord is one crash point from the UDOT Crash Locations MapServer.
// Attribute names match layer 14 (Crash Locations 2025) and layer 10
// (Crash Locations 2018). Fields present on the service and not listed here
// are ignored. LAT_UTM_Y and LONG_UTM_X are the published attribute names;
// the layer spatial reference is EPSG:26912, not WGS84.
type UDOTCrashRecord struct {
	ObjectID              int64
	CrashID               float64
	CrashDatetime         time.Time
	CountyName            string
	RegionName            string
	MainRoadName          string
	Route                 string
	StartAccum            float64
	RouteDirection        string
	LatUTMY               float64
	LongUTMX              float64
	CrashSeverity         string
	FirstHarmfulEvent     string
	WildAnimalRelated     string
	DomesticAnimalRelated string
	CurrentAsOf           time.Time
	GeometryX             float64
	GeometryY             float64
	HasGeometry           bool
}

// WildAnimal reports the verified flag value. Distinct values on layer 14
// were "Y" and "N".
func (r UDOTCrashRecord) WildAnimal() bool {
	return r.WildAnimalRelated == "Y"
}

// ParseUDOTCrashes reads an ArcGIS feature set or GeoJSON FeatureCollection.
// Dates are ArcGIS epoch milliseconds. The caller supplies one layer page;
// the service maxRecordCount is 2,000.
func ParseUDOTCrashes(r io.Reader) ([]UDOTCrashRecord, error) {
	if r == nil {
		return nil, errors.New("nil UDOT crash reader")
	}
	data, err := io.ReadAll(io.LimitReader(r, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("read UDOT crashes: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty UDOT crash source")
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("decode UDOT crash JSON: %w", err)
	}
	if errObj, ok := root["error"]; ok && errObj != nil {
		return nil, fmt.Errorf("UDOT crash service error: %v", errObj)
	}
	raw, ok := root["features"].([]any)
	if !ok {
		return nil, errors.New("UDOT crash JSON has no features array")
	}
	out := make([]UDOTCrashRecord, 0, len(raw))
	for _, item := range raw {
		feature, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, parseUDOTCrashFeature(feature))
	}
	return out, nil
}

func parseUDOTCrashFeature(feature map[string]any) UDOTCrashRecord {
	attrs := feature
	if a, ok := feature["attributes"].(map[string]any); ok {
		attrs = a
	} else if p, ok := feature["properties"].(map[string]any); ok {
		attrs = p
	}
	rec := UDOTCrashRecord{
		ObjectID:              int64(udotNumber(attrs["OBJECTID"])),
		CrashID:               udotNumber(attrs["CRASH_ID"]),
		CrashDatetime:         udotTime(attrs["CRASH_DATETIME"]),
		CountyName:            udotString(attrs["COUNTY_NAME"]),
		RegionName:            udotString(attrs["REGION_NAME"]),
		MainRoadName:          udotString(attrs["MAIN_ROAD_NAME"]),
		Route:                 udotString(attrs["ROUTE"]),
		StartAccum:            udotNumber(attrs["START_ACCUM"]),
		RouteDirection:        udotString(attrs["ROUTE_DIRECTION"]),
		LatUTMY:               udotNumber(attrs["LAT_UTM_Y"]),
		LongUTMX:              udotNumber(attrs["LONG_UTM_X"]),
		CrashSeverity:         udotString(attrs["CRASH_SEVERITY_DESC"]),
		FirstHarmfulEvent:     udotString(attrs["FIRST_HARMFUL_EVENT_DESC"]),
		WildAnimalRelated:     udotString(attrs["WILD_ANIMAL_RELATED"]),
		DomesticAnimalRelated: udotString(attrs["DOMESTIC_ANIMAL_RELATED"]),
		CurrentAsOf:           udotTime(attrs["CURRENT_AS_OF_DATE"]),
	}
	if g, ok := feature["geometry"].(map[string]any); ok {
		if _, hasX := g["x"]; hasX {
			rec.GeometryX = udotNumber(g["x"])
			rec.GeometryY = udotNumber(g["y"])
			rec.HasGeometry = true
		}
	}
	return rec
}

func udotString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func udotNumber(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func udotTime(v any) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch t := v.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(t))
		if err != nil {
			return time.Time{}
		}
		return parsed.UTC()
	default:
		ms := udotNumber(v)
		if ms == 0 {
			return time.Time{}
		}
		return time.UnixMilli(int64(ms)).UTC()
	}
}
