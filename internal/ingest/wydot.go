package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const wydotRoadkillLayerURL = "https://gisservices.wyoroad.info/arcgis/rest/services/Roadkill/Roadkill_Reports/FeatureServer/0"

// wydotSensitivePublicDelay is the public embargo for sensitive species records.
const wydotSensitivePublicDelay = 30 * 24 * time.Hour

// WYDOTRoadkillArtifact returns provenance for the anonymous WYDOT Roadkill
// Reports layer checked on 2026-11-03. No raw feature download was stored, so
// the checksum and object key stay empty. The layer copyright text was empty.
func WYDOTRoadkillArtifact() Artifact {
	return Artifact{
		SourceID:    "c229c363-360a-45d8-b3fa-81d2ce3ea9a7",
		Name:        "WYDOT · Roadkill Reports (511 app)",
		URL:         wydotRoadkillLayerURL,
		License:     "license_unresolved",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 38, 2, 0, time.UTC),
	}
}

// WYDOTRoadkillRecord is one roadside report from the WYDOT Roadkill Reports
// feature layer. Field names follow that layer's metadata.
type WYDOTRoadkillRecord struct {
	PK                int64     `json:"pk"`
	Received          time.Time `json:"received_ts"`
	Status            string    `json:"status"`
	Location          string    `json:"location"`
	Route             string    `json:"route"`
	Highway           string    `json:"highway"`
	Milepost          float64   `json:"milepost"`
	CardinalDirection string    `json:"cardinal_direction"`
	County            string    `json:"county"`
	District          int       `json:"district"`
	Longitude         float64   `json:"longitude"`
	Latitude          float64   `json:"latitude"`
	RawLongitude      float64   `json:"raw_longitude,omitempty"`
	RawLatitude       float64   `json:"raw_latitude,omitempty"`
	Species           string    `json:"species"`
	OtherSpecies      string    `json:"other_species"`
	NormalizedSpecies string    `json:"normalized_species"`
	ProhibitedSpecies int       `json:"prohibited_species"`
	Collected         int       `json:"collected"`
	RoadBlockage      int       `json:"road_blockage"`
	InExcludedZone    int       `json:"in_excluded_zone"`
	IsSensitive       bool      `json:"is_sensitive"`
	H3R6              string    `json:"h3_r6,omitempty"`
	PublicAfter       time.Time `json:"public_after,omitempty"`
}

// ParseWYDOTRoadkill reads an ArcGIS JSON query response whose attributes use
// the Roadkill Reports field names. Sensitive species keep raw coordinates on
// the record and publish the H3 resolution 6 centroid.
func ParseWYDOTRoadkill(r io.Reader) ([]WYDOTRoadkillRecord, error) {
	var payload struct {
		Features []struct {
			Attributes map[string]any `json:"attributes"`
		} `json:"features"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode wydot roadkill: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("wydot roadkill service error %d: %s", payload.Error.Code, payload.Error.Message)
	}

	records := make([]WYDOTRoadkillRecord, 0, len(payload.Features))
	for i, feature := range payload.Features {
		if feature.Attributes == nil {
			return nil, fmt.Errorf("wydot feature %d: missing attributes", i)
		}
		rec, err := wydotRecordFromAttributes(feature.Attributes)
		if err != nil {
			return nil, fmt.Errorf("wydot feature %d: %w", i, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func wydotRecordFromAttributes(attrs map[string]any) (WYDOTRoadkillRecord, error) {
	received, err := wydotTime(attrs["received_ts"])
	if err != nil {
		return WYDOTRoadkillRecord{}, fmt.Errorf("received_ts: %w", err)
	}
	lat, err := wydotFloat(attrs["latitude"])
	if err != nil {
		return WYDOTRoadkillRecord{}, fmt.Errorf("latitude: %w", err)
	}
	lon, err := wydotFloat(attrs["longitude"])
	if err != nil {
		return WYDOTRoadkillRecord{}, fmt.Errorf("longitude: %w", err)
	}
	milepost, err := wydotFloat(attrs["milepost"])
	if err != nil {
		return WYDOTRoadkillRecord{}, fmt.Errorf("milepost: %w", err)
	}
	species := wydotString(attrs["species"])
	normalized := NormalizeWYDOTSpecies(species)
	rec := WYDOTRoadkillRecord{
		PK:                wydotInt64(attrs["pk"]),
		Received:          received,
		Status:            wydotString(attrs["status"]),
		Location:          wydotString(attrs["location"]),
		Route:             wydotString(attrs["route"]),
		Highway:           wydotString(attrs["highway"]),
		Milepost:          milepost,
		CardinalDirection: wydotString(attrs["cardinal_direction"]),
		County:            wydotString(attrs["county"]),
		District:          int(wydotInt64(attrs["district"])),
		Longitude:         lon,
		Latitude:          lat,
		RawLongitude:      lon,
		RawLatitude:       lat,
		Species:           species,
		OtherSpecies:      wydotString(attrs["other_species"]),
		NormalizedSpecies: normalized,
		ProhibitedSpecies: int(wydotInt64(attrs["prohibited_species"])),
		Collected:         int(wydotInt64(attrs["collected"])),
		RoadBlockage:      int(wydotInt64(attrs["road_blockage"])),
		InExcludedZone:    int(wydotInt64(attrs["in_excluded_zone"])),
		IsSensitive:       IsWYDOTSensitiveSpecies(normalized),
	}
	if !rec.IsSensitive {
		return rec, nil
	}
	if math.Abs(lat) > 90 || math.Abs(lon) > 180 || (lat == 0 && lon == 0) {
		return WYDOTRoadkillRecord{}, errors.New("sensitive species requires coordinates")
	}
	cell, centLat, centLon, err := ComputeH3R6(lat, lon)
	if err != nil {
		return WYDOTRoadkillRecord{}, err
	}
	rec.H3R6 = cell
	rec.Latitude = centLat
	rec.Longitude = centLon
	if !received.IsZero() {
		rec.PublicAfter = received.Add(wydotSensitivePublicDelay)
	}
	return rec, nil
}

// NormalizeWYDOTSpecies maps labels observed on the Roadkill Reports layer
// onto snake_case names.
func NormalizeWYDOTSpecies(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "antelope", "pronghorn":
		return "pronghorn"
	case "bighorn sheep", "bighorn":
		return "bighorn_sheep"
	case "black bear":
		return "black_bear"
	case "coyote":
		return "coyote"
	case "deer":
		return "deer"
	case "elk":
		return "elk"
	case "grizzly bear", "grizzly":
		return "grizzly_bear"
	case "moose":
		return "moose"
	case "mountain goat":
		return "mountain_goat"
	case "mountain lion", "cougar", "puma":
		return "mountain_lion"
	case "wild bison", "bison":
		return "wild_bison"
	case "wild turkey", "turkey":
		return "wild_turkey"
	case "wolf", "gray wolf", "grey wolf":
		return "gray_wolf"
	case "":
		return "unknown"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(raw)), " ", "_")
	}
}

// IsWYDOTSensitiveSpecies reports taxa whose precise coordinates stay off
// public outputs. The live layer returned grizzly bear, wolf, mountain lion,
// black bear, and bighorn sheep together with latitude and longitude fields.
func IsWYDOTSensitiveSpecies(normalized string) bool {
	switch normalized {
	case "grizzly_bear", "gray_wolf", "mountain_lion", "black_bear", "bighorn_sheep":
		return true
	default:
		return false
	}
}

func wydotString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func wydotFloat(v any) (float64, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case float64:
		return t, nil
	case json.Number:
		return t.Float64()
	case string:
		if strings.TrimSpace(t) == "" {
			return 0, nil
		}
		return strconv.ParseFloat(strings.TrimSpace(t), 64)
	default:
		return 0, fmt.Errorf("unsupported number %T", v)
	}
}

func wydotInt64(v any) int64 {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return int64(t)
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			f, ferr := t.Float64()
			if ferr != nil {
				return 0
			}
			return int64(f)
		}
		return n
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func wydotTime(v any) (time.Time, error) {
	switch t := v.(type) {
	case nil:
		return time.Time{}, nil
	case float64:
		return time.UnixMilli(int64(t)).UTC(), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return time.Time{}, err
		}
		return time.UnixMilli(n).UTC(), nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return time.Time{}, nil
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return time.UnixMilli(n).UTC(), nil
		}
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, err
		}
		return parsed.UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time %T", v)
	}
}
