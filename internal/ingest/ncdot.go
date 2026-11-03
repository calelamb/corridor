package ingest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

// NCDOTAnimalCrashes2025URL is the 2025 reported-animal-crash FeatureServer layer.
// The metadata check time is 2026-11-03T00:41:30Z. That timestamp is not a file hash.
// CrashType "ANIMAL" is a flag, not a species. This layer is not a carcass inventory.
const NCDOTAnimalCrashes2025URL = "https://services.arcgis.com/NuWFvHYDMVmmxMeM/arcgis/rest/services/AnimalCrashes2025_Mapped/FeatureServer/0"

// ncdotAnimalCrashesWebMapItem is the public web map cited by the safety-maps page.
// ArcGIS item ids are UUIDs without hyphens.
const ncdotAnimalCrashesWebMapItem = "7bf2a5ae-f68a-4719-bbc3-3b994e368330"

// NCDOTAnimalCrashesArtifact is provenance for the 2025 layer metadata check.
// No feature file was stored, so the checksum and object key stay empty.
// The web-map item carries a liability disclaimer and no reuse grant.
func NCDOTAnimalCrashesArtifact() Artifact {
	return Artifact{
		SourceID:    ncdotAnimalCrashesWebMapItem,
		Name:        "NCDOT · Reported Animal Crashes 2025",
		URL:         NCDOTAnimalCrashes2025URL,
		License:     "license_unresolved",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 41, 30, 0, time.UTC),
	}
}

// NCDOTAnimalCrashRecord is one point from an NCDOT AnimalCrashes feature page.
// Field names follow the 2025 layer. CrashType is the service flag; the value
// observed on that layer was "ANIMAL", which does not name a species.
// GISMilepo is nil when the service value is blank. A reported zero stays zero.
// Latitude and Longitude stay on a public encoding because the layer has no species.
type NCDOTAnimalCrashRecord struct {
	FID        int64     `json:"fid"`
	CrashID    string    `json:"crash_id,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	Time       string    `json:"time,omitempty"`
	Severity   string    `json:"severity,omitempty"`
	CrashType  string    `json:"crash_type,omitempty"`
	CountyName string    `json:"county_name,omitempty"`
	City       string    `json:"city,omitempty"`
	OnRd       string    `json:"on_rd,omitempty"`
	Distance   string    `json:"distance,omitempty"`
	Direction  string    `json:"direction,omitempty"`
	FromRd     string    `json:"from_rd,omitempty"`
	TowardRd   string    `json:"toward_rd,omitempty"`
	GISRteTxt  string    `json:"gis_rte_txt,omitempty"`
	GISMilepo  *float64  `json:"gis_milepo,omitempty"`
	Latitude   *float64  `json:"latitude,omitempty"`
	Longitude  *float64  `json:"longitude,omitempty"`
	SourceID   string    `json:"source_id,omitempty"`
	MapMethod  string    `json:"map_method,omitempty"`
	CrashRepor string    `json:"crash_repor,omitempty"`
	LocError   string    `json:"loc_error,omitempty"`
}

// PublicRecord returns the crash for a public response. The layer has no species,
// so the reported point is not generalized and is not labeled as a carcass.
func (r NCDOTAnimalCrashRecord) PublicRecord() NCDOTAnimalCrashRecord {
	out := r
	out.GISMilepo = cloneFloat(r.GISMilepo)
	out.Latitude = cloneFloat(r.Latitude)
	out.Longitude = cloneFloat(r.Longitude)
	return out
}

// ParseNCDOTAnimalCrashes reads one ArcGIS feature JSON page. The caller supplies
// the page. Dates are epoch milliseconds or a date string. Geometry is ignored;
// the point comes from the Latitude and Longitude attributes.
func ParseNCDOTAnimalCrashes(r io.Reader) ([]NCDOTAnimalCrashRecord, error) {
	if r == nil {
		return nil, errors.New("nil ncdot animal-crash reader")
	}
	data, err := io.ReadAll(io.LimitReader(r, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("read ncdot animal crashes: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty ncdot animal-crash source")
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("decode ncdot animal-crash JSON: %w", err)
	}
	if errObj, ok := root["error"]; ok && errObj != nil {
		return nil, ncdotServiceError(errObj)
	}
	raw, ok := root["features"].([]any)
	if !ok {
		return nil, errors.New("ncdot animal-crash JSON has no features array")
	}
	out := make([]NCDOTAnimalCrashRecord, 0, len(raw))
	for i, item := range raw {
		feature, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("ncdot animal-crash feature %d: expected object", i)
		}
		rec, err := ncdotAnimalCrashFromFeature(feature)
		if err != nil {
			return nil, fmt.Errorf("ncdot animal-crash feature %d: %w", i, err)
		}
		out = append(out, rec)
	}
	return out, nil
}

func ncdotAnimalCrashFromFeature(feature map[string]any) (NCDOTAnimalCrashRecord, error) {
	attrs, err := ncdotAttributes(feature)
	if err != nil {
		return NCDOTAnimalCrashRecord{}, err
	}
	observed, err := ncdotTime(attrs["Date"])
	if err != nil {
		return NCDOTAnimalCrashRecord{}, fmt.Errorf("Date: %w", err)
	}
	milepost, err := ncdotOptionalFloat(attrs["GIS_Milepo"])
	if err != nil {
		return NCDOTAnimalCrashRecord{}, fmt.Errorf("GIS_Milepo: %w", err)
	}
	latitude, err := ncdotOptionalFloat(attrs["Latitude"])
	if err != nil {
		return NCDOTAnimalCrashRecord{}, fmt.Errorf("Latitude: %w", err)
	}
	longitude, err := ncdotOptionalFloat(attrs["Longitude"])
	if err != nil {
		return NCDOTAnimalCrashRecord{}, fmt.Errorf("Longitude: %w", err)
	}
	return NCDOTAnimalCrashRecord{
		FID:        ncdotInt64(attrs["FID"]),
		CrashID:    ncdotText(attrs["Crash_ID"]),
		Date:       observed,
		Time:       ncdotText(attrs["Time"]),
		Severity:   ncdotText(attrs["Severity"]),
		CrashType:  ncdotText(attrs["CrashType"]),
		CountyName: ncdotText(attrs["CountyName"]),
		City:       ncdotText(attrs["City"]),
		OnRd:       ncdotText(attrs["OnRd"]),
		Distance:   ncdotText(attrs["Distance"]),
		Direction:  ncdotText(attrs["Direction"]),
		FromRd:     ncdotText(attrs["FromRd"]),
		TowardRd:   ncdotText(attrs["TowardRd"]),
		GISRteTxt:  ncdotText(attrs["GIS_RteTxt"]),
		GISMilepo:  milepost,
		Latitude:   latitude,
		Longitude:  longitude,
		SourceID:   ncdotText(attrs["Source_ID"]),
		MapMethod:  ncdotText(attrs["MapMethod"]),
		CrashRepor: ncdotText(attrs["CrashRepor"]),
		LocError:   ncdotText(attrs["LOC_ERROR"]),
	}, nil
}

func ncdotAttributes(feature map[string]any) (map[string]any, error) {
	if attrs, ok := feature["attributes"].(map[string]any); ok {
		return attrs, nil
	}
	if props, ok := feature["properties"].(map[string]any); ok {
		return props, nil
	}
	return nil, errors.New("missing attributes")
}

func ncdotServiceError(errObj any) error {
	fields, ok := errObj.(map[string]any)
	if !ok {
		return fmt.Errorf("ncdot animal-crash service error: %v", errObj)
	}
	code := ncdotInt64(fields["code"])
	message := ncdotText(fields["message"])
	if message == "" {
		return fmt.Errorf("ncdot animal-crash service error %d", code)
	}
	return fmt.Errorf("ncdot animal-crash service error %d: %s", code, message)
}

func ncdotOptionalFloat(v any) (*float64, error) {
	switch t := v.(type) {
	case nil:
		return nil, nil
	case float64:
		f := t
		return &f, nil
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return nil, err
		}
		return &f, nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, nil
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return &f, nil
	default:
		return nil, fmt.Errorf("unsupported number %T", v)
	}
}

func ncdotText(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return ""
		}
		if t == math.Trunc(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func ncdotInt64(v any) int64 {
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

func ncdotTime(v any) (time.Time, error) {
	switch t := v.(type) {
	case nil:
		return time.Time{}, nil
	case float64:
		return time.UnixMilli(int64(t)).UTC(), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			f, ferr := t.Float64()
			if ferr != nil {
				return time.Time{}, err
			}
			n = int64(f)
		}
		return time.UnixMilli(n).UTC(), nil
	case string:
		return ncdotTimeString(t)
	default:
		return time.Time{}, fmt.Errorf("unsupported date %T", v)
	}
}

func ncdotTimeString(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.UnixMilli(n).UTC(), nil
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
		"1/2/2006 15:04:05",
		"1/2/2006",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date %q", s)
}

func cloneFloat(v *float64) *float64 {
	if v == nil {
		return nil
	}
	f := *v
	return &f
}
