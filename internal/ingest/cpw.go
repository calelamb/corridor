// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CPWStructureType classifies wildlife crossing infrastructure and fencing surveyed
// by Colorado Parks and Wildlife (CPW) and the Colorado Department of Transportation (CDOT).
type CPWStructureType string

const (
	// CPWStructureOverpass represents engineered wildlife land bridges / overpasses.
	CPWStructureOverpass CPWStructureType = "overpass"
	// CPWStructureUnderpass represents dedicated large-span wildlife bridges or underpasses.
	CPWStructureUnderpass CPWStructureType = "underpass"
	// CPWStructureLargeMammalCulvert represents oversized box culverts accommodating elk, deer, and bighorn sheep.
	CPWStructureLargeMammalCulvert CPWStructureType = "large mammal culvert"
	// CPWStructureSmallMammalCulvert represents small pipes or culverts for mesocarnivores and herptiles.
	CPWStructureSmallMammalCulvert CPWStructureType = "small mammal culvert"
	// CPWStructureWildlifeJumpOut represents earthen escape ramps enabling wildlife to exit the highway right-of-way.
	CPWStructureWildlifeJumpOut CPWStructureType = "wildlife jump-out"
	// CPWStructure8ftExclusionFence represents standard 8-foot high-tensile wildlife exclusion fencing.
	CPWStructure8ftExclusionFence CPWStructureType = "8ft ungulate exclusion fence"
	// CPWStructureOther represents other or uncategorized mitigation structures.
	CPWStructureOther CPWStructureType = "other"
)

// CPWEffectivenessTier classifies the empirical collision reduction capability
// of a mitigation structure based on peer-reviewed wildlife ecology literature.
type CPWEffectivenessTier string

const (
	// CPWEffectivenessTierHigh achieves 80-95% collision reduction (e.g. Overpasses, Underpasses with fencing).
	CPWEffectivenessTierHigh CPWEffectivenessTier = "High"
	// CPWEffectivenessTierMedium achieves 50-80% collision reduction (e.g. Large mammal culverts, standalone fencing).
	CPWEffectivenessTierMedium CPWEffectivenessTier = "Medium"
	// CPWEffectivenessTierLow achieves <50% collision reduction (e.g. Small mammal culverts, warning systems).
	CPWEffectivenessTierLow CPWEffectivenessTier = "Low"
	// CPWEffectivenessTierEscape indicates egress/safety infrastructure (e.g. Wildlife jump-outs).
	CPWEffectivenessTierEscape CPWEffectivenessTier = "Escape"
)

// CPWCrossingRecord represents a parsed crossing structure or fencing segment in Colorado.
type CPWCrossingRecord struct {
	ID                string               `json:"id"`
	StructureType     CPWStructureType     `json:"structure_type"`     // overpass, underpass, large mammal culvert, etc.
	RawStructureType  string               `json:"raw_structure_type"` // Original raw type
	HighwayRoute      string               `json:"highway_route"`      // Canonical corridor, e.g. "US-34 Rocky Mountain", "I-70", "US-550"
	RawHighwayRoute   string               `json:"raw_highway_route"`  // Original route text
	Milepost          float64              `json:"milepost,omitempty"` // Highway milepost if available
	Coordinates       [2]float64           `json:"coordinates"`        // [longitude, latitude]
	Longitude         float64              `json:"longitude"`
	Latitude          float64              `json:"latitude"`
	ConstructionYear  int                  `json:"construction_year"`  // Construction or installation year (e.g. 2016)
	EffectivenessTier CPWEffectivenessTier `json:"effectiveness_tier"` // High, Medium, Low, Escape
	TargetSpecies     []string             `json:"target_species,omitempty"`
	FenceHeightFeet   float64              `json:"fence_height_feet,omitempty"`
	FenceLengthMeters float64              `json:"fence_length_meters,omitempty"`
	Notes             string               `json:"notes,omitempty"`
	GeometryJSON      json.RawMessage      `json:"geometry,omitempty"`
}

// CPWCrossingsArtifact returns the provenance artifact metadata for CPW crossing infrastructure.
func CPWCrossingsArtifact() Artifact {
	const hash = "06e467aec7ba27d5107a5e00d79c4b11221d46836625a2b30a62c8dea0e63623"
	return Artifact{
		SourceID:    "c9b1399f-e68a-4db3-982c-b5f76e737521",
		Name:        "CPW · Colorado Wildlife Crossing Structures & Fencing",
		URL:         "https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Wildlife_Crossing_public/FeatureServer/0",
		License:     "State Government Public Record · Colorado Parks and Wildlife (CPW / CDOT)",
		SHA256:      hash,
		ObjectKey:   "sha256/" + hash + "/cpw-crossings.json",
		RetrievedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

var (
	reYearDigits   = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	reCOInterstate = map[int]bool{25: true, 70: true, 76: true, 225: true, 270: true}
	reCOUSHighway  = map[int]bool{6: true, 24: true, 34: true, 36: true, 40: true, 50: true, 84: true, 85: true, 138: true, 160: true, 285: true, 287: true, 350: true, 385: true, 400: true, 491: true, 550: true}
)

// NormalizeCPWStructureType parses and standardizes structure descriptions into canonical CPWStructureType.
func NormalizeCPWStructureType(raw string) CPWStructureType {
	clean := strings.ToLower(strings.TrimSpace(raw))
	if clean == "" {
		return CPWStructureOther
	}

	// 1. Overpass
	if strings.Contains(clean, "overpass") || strings.Contains(clean, "landbridge") ||
		strings.Contains(clean, "land bridge") || strings.Contains(clean, "over-crossing") ||
		strings.Contains(clean, "overcrossing") {
		return CPWStructureOverpass
	}

	// 2. Underpass (bridge underpasses, mammal spans)
	if strings.Contains(clean, "underpass") || strings.Contains(clean, "wildlife span") ||
		strings.Contains(clean, "bridge underpass") || strings.Contains(clean, "wildlife bridge") {
		return CPWStructureUnderpass
	}

	// 3. Jump-out (escape ramp)
	if strings.Contains(clean, "jump-out") || strings.Contains(clean, "jump out") ||
		strings.Contains(clean, "jumpout") || strings.Contains(clean, "escape ramp") ||
		strings.Contains(clean, "earthen ramp") {
		return CPWStructureWildlifeJumpOut
	}

	// 4. Small mammal culvert (evaluated before large mammal culvert)
	if strings.Contains(clean, "small mammal culvert") ||
		(strings.Contains(clean, "culvert") && (strings.Contains(clean, "small") || strings.Contains(clean, "pipe") || strings.Contains(clean, "herp"))) ||
		strings.Contains(clean, "pipe") {
		return CPWStructureSmallMammalCulvert
	}

	// 5. Large mammal culvert
	if strings.Contains(clean, "large mammal culvert") ||
		(strings.Contains(clean, "culvert") && (strings.Contains(clean, "large") || strings.Contains(clean, "box") || strings.Contains(clean, "mammal"))) {
		return CPWStructureLargeMammalCulvert
	}

	// 6. Exclusion fence
	if strings.Contains(clean, "fence") || strings.Contains(clean, "fencing") ||
		strings.Contains(clean, "8ft") || strings.Contains(clean, "8-ft") ||
		strings.Contains(clean, "8 foot") || strings.Contains(clean, "exclusion") {
		return CPWStructure8ftExclusionFence
	}

	return CPWStructureOther
}

// NormalizeCPWRoute extracts and normalizes highway route identifiers across Colorado corridors,
// including iconic segments like "US-34 Rocky Mountain", "I-70", and "US-550".
func NormalizeCPWRoute(rawRoute, corridorDesc string, lon, lat float64) string {
	clean := strings.TrimSpace(rawRoute)
	desc := strings.TrimSpace(corridorDesc)
	combo := strings.ToLower(clean + " " + desc)

	// Explicit override checks for national park and mountain corridors
	if strings.Contains(combo, "rocky mountain") || strings.Contains(combo, "rmnp") ||
		strings.Contains(combo, "estes park") || strings.Contains(combo, "big thompson") {
		return "US-34 Rocky Mountain"
	}

	// Coordinate-based detection for Rocky Mountain US-34 corridor (Grand Lake to Estes Park)
	if (lon >= -105.9 && lon <= -105.2 && lat >= 40.2 && lat <= 40.55) &&
		(strings.Contains(clean, "34") || strings.EqualFold(clean, "US-34") || strings.EqualFold(clean, "US 34")) {
		return "US-34 Rocky Mountain"
	}

	// Extract numerical route digits
	routeNum := -1
	digitsOnly := ""
	for _, ch := range clean {
		if ch >= '0' && ch <= '9' {
			digitsOnly += string(ch)
		}
	}
	if digitsOnly != "" {
		if val, err := strconv.Atoi(digitsOnly); err == nil {
			routeNum = val
		}
	}

	if routeNum == 34 || strings.Contains(clean, "34") {
		if strings.Contains(combo, "rocky") || (lon >= -106.0 && lon <= -105.1 && lat >= 40.1 && lat <= 40.6) {
			return "US-34 Rocky Mountain"
		}
		return "US-34"
	}

	if routeNum == 70 || strings.Contains(clean, "70") {
		return "I-70"
	}

	if routeNum == 550 || strings.Contains(clean, "550") {
		return "US-550"
	}

	if routeNum == 9 || strings.Contains(clean, "SH-9") || strings.Contains(clean, "CO-9") || strings.EqualFold(clean, "9") {
		return "SH-9"
	}

	if routeNum == 160 || strings.Contains(clean, "160") {
		return "US-160"
	}

	if routeNum == 285 || strings.Contains(clean, "285") {
		return "US-285"
	}

	if routeNum == 82 || strings.Contains(clean, "82") {
		return "SH-82"
	}

	if routeNum > 0 {
		if reCOInterstate[routeNum] {
			return fmt.Sprintf("I-%d", routeNum)
		}
		if reCOUSHighway[routeNum] {
			return fmt.Sprintf("US-%d", routeNum)
		}
		return fmt.Sprintf("SH-%d", routeNum)
	}

	if clean != "" {
		return clean
	}
	return "Unknown Route"
}

// DetermineCPWEffectivenessTier determines the mitigation tier based on an explicit attribute
// or peer-reviewed literature defaults for the given structure type.
func DetermineCPWEffectivenessTier(explicit string, st CPWStructureType) CPWEffectivenessTier {
	norm := strings.ToLower(strings.TrimSpace(explicit))
	switch {
	case strings.Contains(norm, "high") || strings.Contains(norm, "tier 1") || strings.Contains(norm, "rank 1") || norm == "tier1":
		return CPWEffectivenessTierHigh
	case strings.Contains(norm, "med") || strings.Contains(norm, "mod") || strings.Contains(norm, "tier 2") || strings.Contains(norm, "rank 2") || norm == "tier2":
		return CPWEffectivenessTierMedium
	case strings.Contains(norm, "low") || strings.Contains(norm, "tier 3") || strings.Contains(norm, "rank 3") || norm == "tier3":
		return CPWEffectivenessTierLow
	case strings.Contains(norm, "escape") || strings.Contains(norm, "egress"):
		return CPWEffectivenessTierEscape
	}

	// Literature default derivation: Rytwinski et al. (2016) / Clevenger & Huijser (2011)
	switch st {
	case CPWStructureOverpass:
		return CPWEffectivenessTierHigh // 86-95% reduction
	case CPWStructureUnderpass:
		return CPWEffectivenessTierHigh // 80-90% reduction
	case CPWStructure8ftExclusionFence:
		return CPWEffectivenessTierHigh // Continuous exclusion standard
	case CPWStructureLargeMammalCulvert:
		return CPWEffectivenessTierMedium // 60-75% reduction
	case CPWStructureSmallMammalCulvert:
		return CPWEffectivenessTierLow // Specialized, non-ungulate
	case CPWStructureWildlifeJumpOut:
		return CPWEffectivenessTierEscape // Safety escape mechanism
	default:
		return CPWEffectivenessTierMedium
	}
}

// ParseConstructionYear extracts an integer 4-digit construction year from diverse field types.
func ParseConstructionYear(val any) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		if v >= 1900 && v <= 2100 {
			return v
		}
	case int64:
		if v >= 1900 && v <= 2100 {
			return int(v)
		}
	case float64:
		year := int(math.Round(v))
		if year >= 1900 && year <= 2100 {
			return year
		}
	case string:
		clean := strings.TrimSpace(v)
		matches := reYearDigits.FindStringSubmatch(clean)
		if len(matches) > 1 {
			if y, err := strconv.Atoi(matches[1]); err == nil {
				return y
			}
		}
	}
	return 0
}

// ParseCPWCrossings parses Colorado Parks and Wildlife (CPW) crossing structures and fencing
// from an io.Reader stream. Supports GeoJSON FeatureCollection and ArcGIS FeatureSet JSON.
func ParseCPWCrossings(r io.Reader) ([]CPWCrossingRecord, error) {
	if r == nil {
		return nil, errors.New("input reader is nil")
	}

	data, err := io.ReadAll(io.LimitReader(r, 64<<20)) // 64 MB maximum
	if err != nil {
		return nil, fmt.Errorf("read CPW source: %w", err)
	}

	return ParseCPWCrossingsJSON(data)
}

// ParseCPWCrossingsJSON parses CPW crossing structures from raw JSON bytes.
func ParseCPWCrossingsJSON(data []byte) ([]CPWCrossingRecord, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty CPW crossing source")
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		var arr []map[string]any
		if arrErr := json.Unmarshal(data, &arr); arrErr == nil && len(arr) > 0 {
			return parseCPWFeatureList(arr)
		}
		return nil, fmt.Errorf("decode CPW crossing JSON: %w", err)
	}

	// 1. GeoJSON FeatureCollection or ArcGIS FeatureSet with "features"
	if rawFeatures, ok := root["features"]; ok {
		if featList, ok := rawFeatures.([]any); ok {
			var maps []map[string]any
			for _, f := range featList {
				if m, ok := f.(map[string]any); ok {
					maps = append(maps, m)
				}
			}
			return parseCPWFeatureList(maps)
		}
	}

	// 2. Single Feature
	if isCPWFeature(root) {
		return parseCPWFeatureList([]map[string]any{root})
	}

	return nil, errors.New("unrecognized CPW crossing JSON structure")
}

func isCPWFeature(m map[string]any) bool {
	if _, ok := m["properties"]; ok {
		return true
	}
	if _, ok := m["attributes"]; ok {
		return true
	}
	if _, ok := m["Structure_Type"]; ok || m["StructureType"] != nil || m["structure_type"] != nil {
		return true
	}
	return false
}

func parseCPWFeatureList(features []map[string]any) ([]CPWCrossingRecord, error) {
	records := make([]CPWCrossingRecord, 0, len(features))

	for i, f := range features {
		rec, err := parseCPWSingleFeature(f, i+1)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}

	if len(records) == 0 {
		return nil, errors.New("no valid CPW crossing records parsed")
	}

	return records, nil
}

func parseCPWSingleFeature(f map[string]any, index int) (CPWCrossingRecord, error) {
	var props map[string]any
	if p, ok := f["properties"].(map[string]any); ok {
		props = p
	} else if a, ok := f["attributes"].(map[string]any); ok {
		props = a
	} else {
		props = f
	}

	getAny := func(keys ...string) any {
		for _, k := range keys {
			if v, ok := props[k]; ok && v != nil {
				return v
			}
			for pk, pv := range props {
				if strings.EqualFold(pk, k) && pv != nil {
					return pv
				}
			}
		}
		return nil
	}

	getString := func(keys ...string) string {
		val := getAny(keys...)
		if val == nil {
			return ""
		}
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(fmt.Sprintf("%v", val))
	}

	getFloat := func(keys ...string) float64 {
		val := getAny(keys...)
		if val == nil {
			return 0
		}
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				return f
			}
		}
		return 0
	}

	// 1. Structure Type
	rawType := getString(
		"Structure_Type", "StructureType", "structure_type",
		"Type", "type", "CROSSING_TYPE", "Crossing_Type",
		"FEATURE_TYPE", "FeatureType", "SubType", "Description",
	)
	structureType := NormalizeCPWStructureType(rawType)

	// 2. Highway Route & Corridor
	rawRoute := getString("Highway", "highway", "Route", "route", "HWY", "Road", "Highway_Ref", "RouteID")
	corridorDesc := getString("Corridor", "corridor", "Location", "location", "Project_Name", "ProjectName", "Notes", "Description")

	// 3. Extract Coordinates & Geometry
	var (
		lon, lat float64
		geomJSON json.RawMessage
	)

	if rawGeom, ok := f["geometry"]; ok && rawGeom != nil {
		if b, err := json.Marshal(rawGeom); err == nil {
			geomJSON = b
		}
		if geomMap, ok := rawGeom.(map[string]any); ok {
			geomType, _ := geomMap["type"].(string)
			// Point: [lon, lat]
			if strings.EqualFold(geomType, "Point") {
				if coords, ok := geomMap["coordinates"].([]any); ok && len(coords) >= 2 {
					if c0, ok0 := coords[0].(float64); ok0 {
						lon = c0
					}
					if c1, ok1 := coords[1].(float64); ok1 {
						lat = c1
					}
				}
			} else if strings.EqualFold(geomType, "LineString") {
				// Midpoint of line coordinates
				if coords, ok := geomMap["coordinates"].([]any); ok && len(coords) > 0 {
					midIdx := len(coords) / 2
					if pt, ok := coords[midIdx].([]any); ok && len(pt) >= 2 {
						if c0, ok0 := pt[0].(float64); ok0 {
							lon = c0
						}
						if c1, ok1 := pt[1].(float64); ok1 {
							lat = c1
						}
					}
				}
			} else if strings.EqualFold(geomType, "MultiLineString") || strings.EqualFold(geomType, "Polygon") {
				// Center or first valid point
				if rawCoords, ok := geomMap["coordinates"].([]any); ok && len(rawCoords) > 0 {
					if ring, ok := rawCoords[0].([]any); ok && len(ring) > 0 {
						if pt, ok := ring[0].([]any); ok && len(pt) >= 2 {
							if c0, ok0 := pt[0].(float64); ok0 {
								lon = c0
							}
							if c1, ok1 := pt[1].(float64); ok1 {
								lat = c1
							}
						}
					}
				}
			}
			// ESRI geometry fallback: x, y
			if lon == 0 && lat == 0 {
				if x, ok := geomMap["x"].(float64); ok {
					lon = x
				}
				if y, ok := geomMap["y"].(float64); ok {
					lat = y
				}
			}
		}
	}

	// Properties coordinate fallback
	if lon == 0 && lat == 0 {
		lon = getFloat("Longitude", "longitude", "Lon", "lon", "Long", "long", "X", "x")
		lat = getFloat("Latitude", "latitude", "Lat", "lat", "Y", "y")
	}

	// Normalize Route with location context
	highwayRoute := NormalizeCPWRoute(rawRoute, corridorDesc, lon, lat)

	// 4. Construction Year
	rawYear := getAny(
		"Construction_Year", "ConstructionYear", "Const_Year", "Year_Built",
		"YearBuilt", "Year", "year", "Date_Built", "InstallDate", "YearCompleted",
	)
	constYear := ParseConstructionYear(rawYear)

	// 5. Effectiveness Tier
	rawEffectiveness := getString(
		"Effectiveness_Tier", "EffectivenessTier", "Effectiveness",
		"Efficacy_Tier", "EfficacyTier", "Tier", "Rating", "Rank",
	)
	effectivenessTier := DetermineCPWEffectivenessTier(rawEffectiveness, structureType)

	// 6. Identifier
	recID := getString("GlobalID", "globalid", "OBJECTID", "objectid", "ID", "id")
	if recID == "" {
		recID = fmt.Sprintf("cpw-%d", index)
	}

	// 7. Milepost & Fencing details
	milepost := getFloat("Milepost", "milepost", "MP", "mp", "Begin_MP")
	fenceHeight := getFloat("Fence_Height_Ft", "FenceHeight", "Height_Ft", "Height")
	if fenceHeight == 0 && structureType == CPWStructure8ftExclusionFence {
		fenceHeight = 8.0
	}
	fenceLength := getFloat("Length_M", "FenceLength", "Length_Meters", "Shape__Length")

	// 8. Target Species
	var targetSpecies []string
	if rawSp := getAny("Target_Species", "TargetSpecies", "Species", "Focal_Species"); rawSp != nil {
		if spList, ok := rawSp.([]any); ok {
			for _, sp := range spList {
				targetSpecies = append(targetSpecies, fmt.Sprintf("%v", sp))
			}
		} else if spStr, ok := rawSp.(string); ok {
			parts := strings.Split(spStr, ",")
			for _, p := range parts {
				cleanP := strings.TrimSpace(p)
				if cleanP != "" {
					targetSpecies = append(targetSpecies, cleanP)
				}
			}
		}
	}

	return CPWCrossingRecord{
		ID:                recID,
		StructureType:     structureType,
		RawStructureType:  rawType,
		HighwayRoute:      highwayRoute,
		RawHighwayRoute:   rawRoute,
		Milepost:          milepost,
		Coordinates:       [2]float64{lon, lat},
		Longitude:         lon,
		Latitude:          lat,
		ConstructionYear:  constYear,
		EffectivenessTier: effectivenessTier,
		TargetSpecies:     targetSpecies,
		FenceHeightFeet:   fenceHeight,
		FenceLengthMeters: fenceLength,
		Notes:             corridorDesc,
		GeometryJSON:      geomJSON,
	}, nil
}

// FilterCPWByType filters crossing records by structure type.
func FilterCPWByType(records []CPWCrossingRecord, st CPWStructureType) []CPWCrossingRecord {
	var filtered []CPWCrossingRecord
	for _, r := range records {
		if r.StructureType == st {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterCPWByRoute filters records that match the given highway route substring (case-insensitive).
func FilterCPWByRoute(records []CPWCrossingRecord, routeSub string) []CPWCrossingRecord {
	q := strings.ToLower(strings.TrimSpace(routeSub))
	var filtered []CPWCrossingRecord
	for _, r := range records {
		if strings.Contains(strings.ToLower(r.HighwayRoute), q) || strings.Contains(strings.ToLower(r.RawHighwayRoute), q) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterCPWByEffectiveness filters records by effectiveness tier.
func FilterCPWByEffectiveness(records []CPWCrossingRecord, tier CPWEffectivenessTier) []CPWCrossingRecord {
	var filtered []CPWCrossingRecord
	for _, r := range records {
		if r.EffectivenessTier == tier {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterCPWByYearRange filters records constructed within [minYear, maxYear] inclusive.
func FilterCPWByYearRange(records []CPWCrossingRecord, minYear, maxYear int) []CPWCrossingRecord {
	var filtered []CPWCrossingRecord
	for _, r := range records {
		if r.ConstructionYear >= minYear && (maxYear == 0 || r.ConstructionYear <= maxYear) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
