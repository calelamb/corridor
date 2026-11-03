package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"
)

const idfgRoadkillLayerURL = "https://gisportal-idfg.idaho.gov/hosting/rest/services/Roadkill/Roadkill_Observations/FeatureServer/1"

// IDFGRoadkillArtifact is provenance for the anonymous IDFG roadkill layer
// checked on 2026-11-03. No feature download was stored. The portal license
// field was null and the terms page returned HTTP 503.
func IDFGRoadkillArtifact() Artifact {
	return Artifact{
		SourceID:    "d5184810-1f99-4df9-b6b7-94738830d62c",
		Name:        "IDFG · Roadkill Observations",
		URL:         idfgRoadkillLayerURL,
		License:     "license_unresolved",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 38, 2, 0, time.UTC),
	}
}

// IDFGRoadkillRecord is one observation from the IDFG Roadkill Observations
// layer. Field names are the ones returned by that layer's metadata.
type IDFGRoadkillRecord struct {
	ObjectID          int64     `json:"objectid"`
	Species           string    `json:"species"`
	NormalizedSpecies string    `json:"normalized_species"`
	Source            string    `json:"source,omitempty"`
	Observed          time.Time `json:"observed"`
	Highway           string    `json:"highway,omitempty"`
	Milepost          float64   `json:"milepost,omitempty"`
	County            string    `json:"county,omitempty"`
	Latitude          float64   `json:"latitude,omitempty"`
	Longitude         float64   `json:"longitude,omitempty"`
	RawLatitude       float64   `json:"-"`
	RawLongitude      float64   `json:"-"`
	IsSensitive       bool      `json:"is_sensitive"`
	H3R6              string    `json:"h3_r6,omitempty"`
	PublicAfter       time.Time `json:"public_after,omitempty"`
}

// ParseIDFGRoadkill reads an ArcGIS JSON query response for the IDFG layer.
func ParseIDFGRoadkill(r io.Reader) ([]IDFGRoadkillRecord, error) {
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
		return nil, fmt.Errorf("decode idfg roadkill: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("idfg roadkill service error %d: %s", payload.Error.Code, payload.Error.Message)
	}
	records := make([]IDFGRoadkillRecord, 0, len(payload.Features))
	for i, feature := range payload.Features {
		if feature.Attributes == nil {
			return nil, fmt.Errorf("idfg feature %d: missing attributes", i)
		}
		rec, err := idfgRecordFromAttributes(feature.Attributes)
		if err != nil {
			return nil, fmt.Errorf("idfg feature %d: %w", i, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func idfgRecordFromAttributes(attrs map[string]any) (IDFGRoadkillRecord, error) {
	observed, err := wydotTime(attrs["observed"])
	if err != nil {
		return IDFGRoadkillRecord{}, fmt.Errorf("observed: %w", err)
	}
	lat, err := wydotFloat(attrs["latitude"])
	if err != nil {
		return IDFGRoadkillRecord{}, fmt.Errorf("latitude: %w", err)
	}
	lon, err := wydotFloat(attrs["longitude"])
	if err != nil {
		return IDFGRoadkillRecord{}, fmt.Errorf("longitude: %w", err)
	}
	milepost, err := wydotFloat(attrs["milepost"])
	if err != nil {
		return IDFGRoadkillRecord{}, fmt.Errorf("milepost: %w", err)
	}
	species := wydotString(attrs["species"])
	normalized := NormalizeIDFGSpecies(species)
	rec := IDFGRoadkillRecord{
		ObjectID:          wydotInt64(attrs["OBJECTID"]),
		Species:           species,
		NormalizedSpecies: normalized,
		Source:            wydotString(attrs["source"]),
		Observed:          observed,
		Highway:           wydotString(attrs["highway"]),
		Milepost:          milepost,
		County:            wydotString(attrs["county"]),
		Latitude:          lat,
		Longitude:         lon,
		RawLatitude:       lat,
		RawLongitude:      lon,
		IsSensitive:       IsIDFGSensitiveSpecies(normalized),
	}
	if !rec.IsSensitive {
		return rec, nil
	}
	if math.Abs(lat) > 90 || math.Abs(lon) > 180 || (lat == 0 && lon == 0) {
		return IDFGRoadkillRecord{}, errors.New("sensitive species requires coordinates")
	}
	cell, centLat, centLon, err := ComputeH3R6(lat, lon)
	if err != nil {
		return IDFGRoadkillRecord{}, err
	}
	rec.H3R6 = cell
	rec.Latitude = centLat
	rec.Longitude = centLon
	if !observed.IsZero() {
		rec.PublicAfter = observed.Add(wydotSensitivePublicDelay)
	}
	return rec, nil
}

// PublicRecord withholds a sensitive observation until PublicAfter and never
// copies the reported coordinates.
func (r IDFGRoadkillRecord) PublicRecord(refTime time.Time) (IDFGRoadkillRecord, bool) {
	if r.IsSensitive && (r.PublicAfter.IsZero() || refTime.Before(r.PublicAfter)) {
		return IDFGRoadkillRecord{}, false
	}
	pub := r
	pub.RawLatitude = 0
	pub.RawLongitude = 0
	return pub, true
}

// NormalizeIDFGSpecies maps labels from the layer metadata and the distinct
// species response onto snake_case names.
func NormalizeIDFGSpecies(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mule deer", "odocoileus hemionus":
		return "mule_deer"
	case "white-tailed deer", "white tailed deer", "odocoileus virginianus":
		return "white_tailed_deer"
	case "elk", "cervus canadensis":
		return "elk"
	case "moose", "alces americanus":
		return "moose"
	case "deer", "odocoileus":
		return "deer"
	case "pronghorn", "antelope", "antilocapra americana":
		return "pronghorn"
	case "grizzly bear", "grizzly", "ursus arctos":
		return "grizzly_bear"
	case "gray wolf", "grey wolf", "wolf", "canis lupus":
		return "gray_wolf"
	case "wolverine", "gulo gulo":
		return "wolverine"
	case "canada lynx", "lynx", "lynx canadensis":
		return "canada_lynx"
	case "black bear", "ursus americanus":
		return "black_bear"
	case "bighorn sheep", "ovis canadensis":
		return "bighorn_sheep"
	case "mountain lion", "cougar", "puma concolor":
		return "mountain_lion"
	case "":
		return "unknown"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(raw)), " ", "_")
	}
}

// IsIDFGSensitiveSpecies reports taxa whose coordinates the 2026-11-03
// distinct-values response said must not be published at full precision.
func IsIDFGSensitiveSpecies(normalized string) bool {
	switch normalized {
	case "grizzly_bear", "gray_wolf", "wolverine", "canada_lynx", "black_bear", "bighorn_sheep", "mountain_lion":
		return true
	default:
		return false
	}
}
