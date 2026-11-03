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

const vdotCarcassLayerURL = "https://services.arcgis.com/p5v98VHDX9Atv3l7/arcgis/rest/services/Wildlife_Carcass_Removal_Tracking/FeatureServer/0"

// VDOTCarcassArtifact is provenance for the VDOT Wildlife Carcass Removal
// Tracking layer. RetrievedAt is the 2026-11-03T00:39:59Z metadata check, not
// a file acquisition. No feature download was stored, so the SHA-256 and
// object key stay empty. The Virginia package notes cite 23 U.S.C. § 409,
// which those notes describe as a limit on using this crash information in
// discovery or as evidence. That citation is not a license grant.
// Redistribution stays license_unresolved. The layer notes describe contracted
// removals on VDOT interstates and select other roads, and say the dataset is
// incomplete. Coverage is claimed only for what those notes state.
func VDOTCarcassArtifact() Artifact {
	return Artifact{
		SourceID:    "57dd6e62-48a5-4725-8cf5-e8e92912d3ba",
		Name:        "VDOT · Wildlife Carcass Removal Tracking",
		URL:         vdotCarcassLayerURL,
		License:     "license_unresolved",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 39, 59, 0, time.UTC),
	}
}

// VDOTCarcassRecord is one removal from the Wildlife_Carcass_Removal_Tracking layer.
// Source attributes are objectid, globalid, starttime, wildlife,
// wildlifeother, rtenamedir, latitude_d, longitude_d, Jurisdiction, District,
// and Residency. latitude_d and longitude_d stay on the raw fields, which do
// not marshal. Public coordinates are the reported point for deer and the H3
// resolution 6 centroid for bear.
type VDOTCarcassRecord struct {
	ObjectID          int64     `json:"objectid"`
	GlobalID          string    `json:"globalid,omitempty"`
	StartTime         time.Time `json:"starttime"`
	Wildlife          string    `json:"wildlife,omitempty"`
	WildlifeOther     string    `json:"wildlifeother,omitempty"`
	RouteNameDir      string    `json:"rtenamedir,omitempty"`
	Latitude          float64   `json:"latitude,omitempty"`
	Longitude         float64   `json:"longitude,omitempty"`
	RawLatitude       float64   `json:"-"`
	RawLongitude      float64   `json:"-"`
	Jurisdiction      string    `json:"jurisdiction,omitempty"`
	District          string    `json:"district,omitempty"`
	Residency         string    `json:"residency,omitempty"`
	NormalizedSpecies string    `json:"normalized_species,omitempty"`
	IsSensitive       bool      `json:"is_sensitive"`
	H3R6              string    `json:"h3_r6,omitempty"`
	PublicAfter       time.Time `json:"public_after,omitzero"`
}

// ParseVDOTCarcass reads an ArcGIS JSON feature set for the carcass-removal
// layer. Bear rows publish the H3 resolution 6 centroid. Deer rows keep the
// reported point.
func ParseVDOTCarcass(r io.Reader) ([]VDOTCarcassRecord, error) {
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
		return nil, fmt.Errorf("decode vdot carcass: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("vdot carcass service error %d: %s", payload.Error.Code, payload.Error.Message)
	}
	records := make([]VDOTCarcassRecord, 0, len(payload.Features))
	for i, feature := range payload.Features {
		if feature.Attributes == nil {
			return nil, fmt.Errorf("vdot feature %d: missing attributes", i)
		}
		rec, err := vdotRecordFromAttributes(feature.Attributes)
		if err != nil {
			return nil, fmt.Errorf("vdot feature %d: %w", i, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func vdotRecordFromAttributes(attrs map[string]any) (VDOTCarcassRecord, error) {
	start, err := wydotTime(vdotAttr(attrs, "starttime"))
	if err != nil {
		return VDOTCarcassRecord{}, fmt.Errorf("starttime: %w", err)
	}
	wildlife := wydotString(vdotAttr(attrs, "wildlife"))
	other := wydotString(vdotAttr(attrs, "wildlifeother"))
	normalized := NormalizeVDOTSpecies(wildlife, other)
	sensitive := vdotSensitive(normalized, other)
	lat, lon, err := vdotCoordinates(attrs, sensitive)
	if err != nil {
		return VDOTCarcassRecord{}, err
	}
	rec := VDOTCarcassRecord{
		ObjectID:          wydotInt64(vdotAttr(attrs, "objectid")),
		GlobalID:          wydotString(vdotAttr(attrs, "globalid")),
		StartTime:         start,
		Wildlife:          wildlife,
		WildlifeOther:     other,
		RouteNameDir:      wydotString(vdotAttr(attrs, "rtenamedir")),
		Latitude:          lat,
		Longitude:         lon,
		RawLatitude:       lat,
		RawLongitude:      lon,
		Jurisdiction:      wydotString(vdotAttr(attrs, "Jurisdiction")),
		District:          wydotString(vdotAttr(attrs, "District")),
		Residency:         wydotString(vdotAttr(attrs, "Residency")),
		NormalizedSpecies: normalized,
		IsSensitive:       sensitive,
	}
	if !sensitive {
		return rec, nil
	}
	cell, centLat, centLon, err := ComputeH3R6(lat, lon)
	if err != nil {
		return VDOTCarcassRecord{}, err
	}
	rec.H3R6 = cell
	rec.Latitude = centLat
	rec.Longitude = centLon
	if !start.IsZero() {
		rec.PublicAfter = start.Add(SensitiveQuarantineDuration)
	}
	return rec, nil
}

// vdotAttr reads an attribute by exact name, then case-insensitively, because
// ArcGIS services differ in field-name casing.
func vdotAttr(attrs map[string]any, name string) any {
	if v, ok := attrs[name]; ok {
		return v
	}
	for k, v := range attrs {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return nil
}

func vdotPresent(v any) bool {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) != ""
	}
	return v != nil
}

// vdotCoordinates parses latitude_d and longitude_d. Sensitive records need
// both. Other records may omit both, but never just one, and any supplied
// pair must be finite and in range.
func vdotCoordinates(attrs map[string]any, sensitive bool) (float64, float64, error) {
	rawLat, rawLon := vdotAttr(attrs, "latitude_d"), vdotAttr(attrs, "longitude_d")
	hasLat, hasLon := vdotPresent(rawLat), vdotPresent(rawLon)
	if !hasLat && !hasLon && !sensitive {
		return 0, 0, nil
	}
	if !hasLat || !hasLon {
		if sensitive {
			return 0, 0, errors.New("sensitive species requires coordinates")
		}
		return 0, 0, errors.New("latitude_d and longitude_d must both be present")
	}
	lat, err := wydotFloat(rawLat)
	if err != nil {
		return 0, 0, fmt.Errorf("latitude_d: %w", err)
	}
	lon, err := wydotFloat(rawLon)
	if err != nil {
		return 0, 0, fmt.Errorf("longitude_d: %w", err)
	}
	if math.IsNaN(lat) || math.IsNaN(lon) || math.Abs(lat) > 90 || math.Abs(lon) > 180 {
		return 0, 0, errors.New("coordinates out of range")
	}
	if sensitive && lat == 0 && lon == 0 {
		return 0, 0, errors.New("sensitive species requires coordinates")
	}
	return lat, lon, nil
}

// vdotSensitive fails closed: a bear named in either wildlife or the
// free-text wildlifeother makes the record sensitive.
func vdotSensitive(normalized, other string) bool {
	if IsVDOTSensitiveSpecies(normalized) {
		return true
	}
	return strings.TrimSpace(other) != "" && IsVDOTSensitiveSpecies(normalizeVDOTOther(other))
}

// PublicRecord returns the feed copy. A bear record is withheld until 30 days
// after starttime, and also when starttime is unknown or the point was not
// generalized to H3 resolution 6. Deer is not sensitive and is returned
// immediately. Raw coordinates are cleared on the returned copy.
func (r VDOTCarcassRecord) PublicRecord(refTime time.Time) (VDOTCarcassRecord, bool) {
	if r.IsSensitive && !r.releasable(refTime) {
		return VDOTCarcassRecord{}, false
	}
	pub := r
	pub.RawLatitude = 0
	pub.RawLongitude = 0
	return pub, true
}

func (r VDOTCarcassRecord) releasable(refTime time.Time) bool {
	if r.StartTime.IsZero() || r.H3R6 == "" {
		return false
	}
	return !refTime.Before(r.StartTime.Add(SensitiveQuarantineDuration))
}

// NormalizeVDOTSpecies maps the wildlife coded value, and wildlifeother when
// wildlife is Other, onto a snake_case name. The Bear code is black bear.
// Distinct responses also spelled NotFound as Notfound.
func NormalizeVDOTSpecies(wildlife, wildlifeOther string) string {
	switch strings.ToLower(strings.TrimSpace(wildlife)) {
	case "deer":
		return "deer"
	case "bear", "black bear", "black_bear":
		return "black_bear"
	case "notfound", "not found", "not_found":
		return "not_found"
	case "other":
		return normalizeVDOTOther(wildlifeOther)
	case "":
		if strings.TrimSpace(wildlifeOther) == "" {
			return "unknown"
		}
		return normalizeVDOTOther(wildlifeOther)
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(wildlife)), " ", "_")
	}
}

func normalizeVDOTOther(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "other":
		return "other"
	case "coyote":
		return "coyote"
	case "raccoon":
		return "raccoon"
	case "bird":
		return "bird"
	case "bobcat":
		return "bobcat"
	case "squirrel":
		return "squirrel"
	case "fox":
		return "fox"
	case "skunk":
		return "skunk"
	case "turtle":
		return "turtle"
	case "opossum":
		return "opossum"
	case "groundhog":
		return "groundhog"
	case "dog":
		return "dog"
	case "cat":
		return "cat"
	case "farmanimal", "farm animal":
		return "farm_animal"
	case "otheranimal", "other animal":
		return "other_animal"
	case "bear", "black bear", "black_bear":
		return "black_bear"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(raw)), " ", "_")
	}
}

// IsVDOTSensitiveSpecies reports any bear. It matches whole underscore
// tokens, so free-text values such as "bear_cub" are sensitive while
// "bearded_dragon" is not. Deer and other common wildlifeother values are not
// sensitive on this layer.
func IsVDOTSensitiveSpecies(normalized string) bool {
	for _, token := range strings.FieldsFunc(strings.ToLower(normalized), func(r rune) bool {
		return r == '_' || r == ' ' || r == '-'
	}) {
		if token == "bear" || token == "bears" {
			return true
		}
	}
	return false
}
