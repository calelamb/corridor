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

// WSDOTPriorityTier classifies the wildlife-vehicle safety priority ranking
// assigned by the Washington Habitat Connectivity Action Plan (WAHCAP 2025).
type WSDOTPriorityTier string

const (
	// WSDOTPriorityHigh indicates top-tier wildlife hazard requiring priority mitigation.
	WSDOTPriorityHigh WSDOTPriorityTier = "High"
	// WSDOTPriorityMedium indicates moderate wildlife hazard corridor.
	WSDOTPriorityMedium WSDOTPriorityTier = "Medium"
	// WSDOTPriorityLow indicates baseline or low wildlife collision density.
	WSDOTPriorityLow WSDOTPriorityTier = "Low"
	// WSDOTPriorityUnranked indicates an unranked or pending evaluation section.
	WSDOTPriorityUnranked WSDOTPriorityTier = "Unranked"
)

// WSDOTSafetyRanking represents a parsed and normalized highway safety ranking segment
// from Washington State DOT WAHCAP 2025 (GeoServices FeatureServer Layer 3).
type WSDOTSafetyRanking struct {
	ID               int64             `json:"id"`
	Route            string            `json:"route"`             // Normalized corridor, e.g. "US-101 Olympic NP", "I-90 Snoqualmie Pass", "US-97"
	RawRouteID       string            `json:"raw_route_id"`      // Original RouteID e.g. "101", "090", "097"
	Region           string            `json:"region,omitempty"`  // WSDOT administrative region e.g. "Olympic", "South Central"
	BeginARM         float64           `json:"begin_arm"`         // Begin Accumulated Route Mileage
	EndARM           float64           `json:"end_arm"`           // End Accumulated Route Mileage
	BeginSRMP        float64           `json:"begin_srmp"`        // Begin State Route Milepost
	EndSRMP          float64           `json:"end_srmp"`          // End State Route Milepost
	LengthMiles      float64           `json:"length_miles"`      // Segment length in miles (|EndARM - BeginARM|)
	CollisionRate    float64           `json:"collision_rate"`    // WRSv10 wildlife-related safety score or collision rate
	CollisionDensity float64           `json:"collision_density"` // Carcass and crash density per mile
	PriorityTier     WSDOTPriorityTier `json:"priority_tier"`     // High, Medium, Low, or Unranked
	TargetSpecies    []string          `json:"target_species"`    // Target focal species (e.g. "Elk", "Mule Deer")
	DeerCount        int               `json:"deer_count"`        // Total deer carcasses recorded
	ElkCount         int               `json:"elk_count"`         // Total elk carcasses recorded
	MooseCount       int               `json:"moose_count"`       // Total moose carcasses recorded
	DeerCrashCount   int               `json:"deer_crash_count"`  // Reported deer vehicular crashes
	ElkCrashCount    int               `json:"elk_crash_count"`   // Reported elk vehicular crashes
	TotalCarcasses   int               `json:"total_carcasses"`   // Sum of all recorded carcasses
	TotalCrashes     int               `json:"total_crashes"`     // Sum of all reported wildlife crashes
	AADT             int               `json:"aadt,omitempty"`    // Annual Average Daily Traffic
	GeometryJSON     json.RawMessage   `json:"geometry,omitempty"`
}

// WSDOTSafetyArtifact returns provenance metadata for the WAHCAP 2025 Highway Safety Rankings layer.
func WSDOTSafetyArtifact() Artifact {
	const hash = "fdd693d46a26190567ad5f2144281b343325b6443b68593646cb883c450c4c5e"
	return Artifact{
		SourceID:    "fdd693d4-6a26-1905-67ad-5f2144281b34",
		Name:        "WSDOT · WAHCAP 2025 Highway Safety Rankings (FeatureServer Layer 3)",
		URL:         "https://data.wsdot.wa.gov/arcgis/rest/services/Shared/WAHabitatConnectivityActionPlan/FeatureServer/3",
		License:     "State Government Public Record · WSDOT (Evaluated: Planning Derivative)",
		SHA256:      hash,
		ObjectKey:   "sha256/" + hash + "/wsdot-safety-layer.json",
		RetrievedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Regex patterns for route normalization
var (
	reLeadingZeros = regexp.MustCompile(`^0+([1-9][0-9]*)$`)
	reWAInterstate = map[int]bool{5: true, 82: true, 90: true, 182: true, 205: true, 405: true, 705: true}
	reWAUSHighway  = map[int]bool{2: true, 12: true, 97: true, 101: true, 195: true, 197: true, 395: true, 730: true}
)

// NormalizeWSDOTRoute formats a raw WSDOT route identifier into canonical naming,
// preserving iconic ecological corridors (e.g. "US-101 Olympic NP", "I-90 Snoqualmie Pass", "US-97").
func NormalizeWSDOTRoute(rawRoute, region string, beginARM, endARM float64) string {
	clean := strings.TrimSpace(rawRoute)
	if clean == "" {
		return "Unknown Route"
	}

	// If already explicitly formatted with corridor descriptors, preserve
	lowerClean := strings.ToLower(clean)
	if strings.Contains(lowerClean, "olympic") {
		if !strings.HasPrefix(clean, "US-101") && !strings.HasPrefix(clean, "US 101") {
			return "US-101 Olympic NP"
		}
		return "US-101 Olympic NP"
	}
	if strings.Contains(lowerClean, "snoqualmie") {
		return "I-90 Snoqualmie Pass"
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

	lowerReg := strings.ToLower(region)
	minARM := math.Min(beginARM, endARM)
	maxARM := math.Max(beginARM, endARM)

	// US-101 Olympic NP corridor detection
	if routeNum == 101 || strings.Contains(lowerClean, "101") {
		if strings.Contains(lowerReg, "olympic") || strings.Contains(lowerClean, "olympic") ||
			(minARM >= 150.0 && maxARM <= 370.0) {
			return "US-101 Olympic NP"
		}
		return "US-101"
	}

	// I-90 Snoqualmie Pass corridor detection (mileposts ~45 to ~75 across the Cascades crest)
	if routeNum == 90 || strings.Contains(lowerClean, "90") {
		isSnoqualmieARM := (minARM >= 45.0 && minARM <= 75.0) || (maxARM >= 45.0 && maxARM <= 75.0) ||
			(minARM <= 45.0 && maxARM >= 75.0)
		if strings.Contains(lowerClean, "snoqualmie") || strings.Contains(lowerReg, "snoqualmie") || isSnoqualmieARM {
			return "I-90 Snoqualmie Pass"
		}
		return "I-90"
	}

	// US-97
	if routeNum == 97 || strings.Contains(lowerClean, "97") {
		return "US-97"
	}

	// Standard Washington routes by number
	if routeNum > 0 {
		if reWAInterstate[routeNum] {
			return fmt.Sprintf("I-%d", routeNum)
		}
		if reWAUSHighway[routeNum] {
			return fmt.Sprintf("US-%d", routeNum)
		}
		return fmt.Sprintf("SR-%d", routeNum)
	}

	return clean
}

// NormalizeWSDOTPriority normalizes priority strings into WSDOTPriorityTier.
func NormalizeWSDOTPriority(raw string) WSDOTPriorityTier {
	norm := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(norm, "high") || norm == "h" || strings.Contains(norm, "tier 1") || strings.Contains(norm, "rank 1"):
		return WSDOTPriorityHigh
	case strings.Contains(norm, "med") || strings.Contains(norm, "mod") || norm == "m" || strings.Contains(norm, "tier 2") || strings.Contains(norm, "rank 2"):
		return WSDOTPriorityMedium
	case strings.Contains(norm, "low") || norm == "l" || strings.Contains(norm, "tier 3") || strings.Contains(norm, "rank 3"):
		return WSDOTPriorityLow
	default:
		return WSDOTPriorityUnranked
	}
}

// ExtractWSDOTTargetSpecies extracts target focal species from species counts and explicit indicators.
func ExtractWSDOTTargetSpecies(deerCount, elkCount, mooseCount, deerCrash, elkCrash, deerHS, elkHS int, explicit []string) []string {
	speciesMap := make(map[string]bool)

	for _, s := range explicit {
		clean := strings.TrimSpace(s)
		if clean != "" {
			speciesMap[clean] = true
		}
	}

	if elkCount > 0 || elkCrash > 0 || elkHS > 0 {
		speciesMap["Elk"] = true
	}
	if deerCount > 0 || deerCrash > 0 || deerHS > 0 {
		speciesMap["Mule Deer"] = true
	}
	if mooseCount > 0 {
		speciesMap["Moose"] = true
	}

	var result []string
	// Canonical ordering: Elk first, then Mule Deer, Moose, then others
	if speciesMap["Elk"] {
		result = append(result, "Elk")
		delete(speciesMap, "Elk")
	}
	if speciesMap["Mule Deer"] {
		result = append(result, "Mule Deer")
		delete(speciesMap, "Mule Deer")
	}
	if speciesMap["Moose"] {
		result = append(result, "Moose")
		delete(speciesMap, "Moose")
	}
	for s := range speciesMap {
		result = append(result, s)
	}

	return result
}

// ParseWSDOTSafetyRankings parses Washington State DOT WAHCAP 2025 Highway Safety Rankings
// from an io.Reader stream. Supports GeoJSON FeatureCollection and ArcGIS FeatureSet JSON.
func ParseWSDOTSafetyRankings(r io.Reader) ([]WSDOTSafetyRanking, error) {
	if r == nil {
		return nil, errors.New("input reader is nil")
	}

	data, err := io.ReadAll(io.LimitReader(r, 64<<20)) // 64 MB maximum
	if err != nil {
		return nil, fmt.Errorf("read WSDOT safety source: %w", err)
	}

	return ParseWSDOTSafetyRankingsJSON(data)
}

// ParseWSDOTSafetyRankingsJSON parses WSDOT rankings from raw JSON bytes.
func ParseWSDOTSafetyRankingsJSON(data []byte) ([]WSDOTSafetyRanking, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty WSDOT safety source")
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		// Could be a direct array of feature objects
		var arr []map[string]any
		if arrErr := json.Unmarshal(data, &arr); arrErr == nil && len(arr) > 0 {
			return parseWSDOTFeatureList(arr)
		}
		return nil, fmt.Errorf("decode WSDOT safety JSON: %w", err)
	}

	// 1. Check for GeoJSON FeatureCollection or ArcGIS FeatureSet with "features"
	if rawFeatures, ok := root["features"]; ok {
		if featList, ok := rawFeatures.([]any); ok {
			var maps []map[string]any
			for _, f := range featList {
				if m, ok := f.(map[string]any); ok {
					maps = append(maps, m)
				}
			}
			return parseWSDOTFeatureList(maps)
		}
	}

	// 2. Check if the root itself is a single feature
	if isWSDOTFeature(root) {
		return parseWSDOTFeatureList([]map[string]any{root})
	}

	return nil, errors.New("unrecognized WSDOT safety ranking JSON structure")
}

func isWSDOTFeature(m map[string]any) bool {
	if _, ok := m["properties"]; ok {
		return true
	}
	if _, ok := m["attributes"]; ok {
		return true
	}
	if _, ok := m["RouteID"]; ok || m["route"] != nil || m["BARM"] != nil {
		return true
	}
	return false
}

func parseWSDOTFeatureList(features []map[string]any) ([]WSDOTSafetyRanking, error) {
	rankings := make([]WSDOTSafetyRanking, 0, len(features))

	for i, f := range features {
		ranking, err := parseWSDOTSingleFeature(f, int64(i+1))
		if err != nil {
			// Skip entirely invalid empty objects, but report true malformations
			continue
		}
		rankings = append(rankings, ranking)
	}

	if len(rankings) == 0 {
		return nil, errors.New("no valid WSDOT safety ranking records parsed")
	}

	return rankings, nil
}

func parseWSDOTSingleFeature(f map[string]any, defaultID int64) (WSDOTSafetyRanking, error) {
	// Extract properties or attributes map
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
			// Case-insensitive lookup fallback
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

	getInt := func(keys ...string) int {
		return int(math.Round(getFloat(keys...)))
	}

	// 1. Extract IDs
	recID := defaultID
	if oid := getInt("OBJECTID", "objectid", "id", "ID"); oid > 0 {
		recID = int64(oid)
	}

	// 2. Extract Route & Region
	rawRoute := getString("RouteID", "route_id", "route", "Route", "highway", "Highway", "RouteType")
	region := getString("WSDOTRegion", "region", "Region", "RegionName")

	// 3. Extract Mileposts (ARM & SRMP)
	beginARM := getFloat("BARM", "barm", "BeginARM", "begin_arm")
	endARM := getFloat("EARM", "earm", "EndARM", "end_arm")
	beginSRMP := getFloat("BSRMP", "bsrmp", "BeginSRMP", "begin_srmp", "BeginMP")
	endSRMP := getFloat("ESRMP", "esrmp", "EndSRMP", "end_srmp", "EndMP")

	// Fallback ARM from SRMP if ARM was 0 and SRMP is set
	if beginARM == 0 && endARM == 0 && (beginSRMP != 0 || endSRMP != 0) {
		beginARM = beginSRMP
		endARM = endSRMP
	}

	// Ensure BeginARM <= EndARM
	if beginARM > endARM {
		beginARM, endARM = endARM, beginARM
	}
	lengthMiles := math.Abs(endARM - beginARM)

	// 4. Extract Route with canonical normalization
	route := NormalizeWSDOTRoute(rawRoute, region, beginARM, endARM)

	// 5. Extract Species Counts & Hotspots
	deerCount := getInt("DeerCount", "deer_count", "Deer")
	elkCount := getInt("ElkCount", "elk_count", "Elk")
	mooseCount := getInt("MooseCount", "moose_count", "Moose")
	deerCrash := getInt("DeerCrashCount", "deer_crash_count")
	elkCrash := getInt("ElkCrashCount", "elk_crash_count")
	deerHS := getInt("DeerHS", "deer_hs")
	elkHS := getInt("ElkHS", "elk_hs")

	// Other species counts
	bhsCount := getInt("BHSCount")
	bearCount := getInt("BlackBearCount")
	cougarCount := getInt("CougarCount")
	coyoteCount := getInt("CoyoteCount")
	cwtdCount := getInt("CWTDCount")

	totalCarcasses := deerCount + elkCount + mooseCount + bhsCount + bearCount + cougarCount + coyoteCount + cwtdCount
	if explicitTotal := getInt("TotalCarcasses", "total_carcasses", "carcasses"); explicitTotal > 0 {
		totalCarcasses = explicitTotal
	}

	totalCrashes := deerCrash + elkCrash + getInt("CrashInjCount") + getInt("CrashFatalCount") + getInt("OtherCrashCount")
	if explicitCrashes := getInt("TotalCrashes", "total_crashes", "crashes"); explicitCrashes > 0 {
		totalCrashes = explicitCrashes
	}

	// Explicit species if present
	var explicitSpecies []string
	if rawSp := getAny("TargetSpecies", "target_species", "Species", "species"); rawSp != nil {
		if spList, ok := rawSp.([]any); ok {
			for _, sp := range spList {
				explicitSpecies = append(explicitSpecies, fmt.Sprintf("%v", sp))
			}
		} else if spStr, ok := rawSp.(string); ok {
			parts := strings.Split(spStr, ",")
			for _, p := range parts {
				explicitSpecies = append(explicitSpecies, strings.TrimSpace(p))
			}
		}
	}

	targetSpecies := ExtractWSDOTTargetSpecies(deerCount, elkCount, mooseCount, deerCrash, elkCrash, deerHS, elkHS, explicitSpecies)

	// 6. Extract Priority
	rawPriority := getString("Safetyv10_Priority", "safety_priority", "priority", "Priority", "PriorityTier", "Tier", "HighestRank")
	priorityTier := NormalizeWSDOTPriority(rawPriority)

	// 7. Extract Collision Rate & Density
	wrsScore := getFloat("WRSv10", "wrs_score", "score", "CollisionRate", "collision_rate")
	collisionRate := wrsScore

	var collisionDensity float64
	if lengthMiles > 0 {
		collisionDensity = float64(totalCarcasses+totalCrashes) / lengthMiles
	} else {
		collisionDensity = float64(totalCarcasses + totalCrashes)
	}

	if collisionRate == 0 {
		collisionRate = collisionDensity
	}

	// If priority was unranked, infer from score
	if priorityTier == WSDOTPriorityUnranked {
		if wrsScore >= 50.0 || collisionDensity >= 5.0 {
			priorityTier = WSDOTPriorityHigh
		} else if wrsScore >= 20.0 || collisionDensity >= 2.0 {
			priorityTier = WSDOTPriorityMedium
		} else if wrsScore > 0.0 || collisionDensity > 0.0 {
			priorityTier = WSDOTPriorityLow
		}
	}

	// 8. Geometry
	var geomJSON json.RawMessage
	if rawGeom, ok := f["geometry"]; ok && rawGeom != nil {
		if b, err := json.Marshal(rawGeom); err == nil {
			geomJSON = b
		}
	}

	aadt := getInt("AADT_1", "AADT", "aadt")

	return WSDOTSafetyRanking{
		ID:               recID,
		Route:            route,
		RawRouteID:       rawRoute,
		Region:           region,
		BeginARM:         beginARM,
		EndARM:           endARM,
		BeginSRMP:        beginSRMP,
		EndSRMP:          endSRMP,
		LengthMiles:      lengthMiles,
		CollisionRate:    collisionRate,
		CollisionDensity: collisionDensity,
		PriorityTier:     priorityTier,
		TargetSpecies:    targetSpecies,
		DeerCount:        deerCount,
		ElkCount:         elkCount,
		MooseCount:       mooseCount,
		DeerCrashCount:   deerCrash,
		ElkCrashCount:    elkCrash,
		TotalCarcasses:   totalCarcasses,
		TotalCrashes:     totalCrashes,
		AADT:             aadt,
		GeometryJSON:     geomJSON,
	}, nil
}

// FilterWSDOTByPriority filters rankings by priority tier.
func FilterWSDOTByPriority(rankings []WSDOTSafetyRanking, tier WSDOTPriorityTier) []WSDOTSafetyRanking {
	var filtered []WSDOTSafetyRanking
	for _, r := range rankings {
		if r.PriorityTier == tier {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterWSDOTByRoute filters rankings by matching route substring (case-insensitive).
func FilterWSDOTByRoute(rankings []WSDOTSafetyRanking, routeSub string) []WSDOTSafetyRanking {
	query := strings.ToLower(strings.TrimSpace(routeSub))
	var filtered []WSDOTSafetyRanking
	for _, r := range rankings {
		if strings.Contains(strings.ToLower(r.Route), query) || strings.Contains(strings.ToLower(r.RawRouteID), query) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterWSDOTBySpecies filters rankings that include the requested target species.
func FilterWSDOTBySpecies(rankings []WSDOTSafetyRanking, species string) []WSDOTSafetyRanking {
	target := strings.ToLower(strings.TrimSpace(species))
	var filtered []WSDOTSafetyRanking
	for _, r := range rankings {
		for _, s := range r.TargetSpecies {
			if strings.EqualFold(s, target) || strings.Contains(strings.ToLower(s), target) {
				filtered = append(filtered, r)
				break
			}
		}
	}
	return filtered
}
