// Package route implements the public route risk assessment engine for Corridor.
// It calculates road segment risk, aggregates journey-level collision exposure,
// applies peer-reviewed ungulate diurnal movement multipliers, and constructs
// plain-language driver advisories.
package route

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// DefaultSegmentLengthKm represents the standard 0.1-mile (160.9344m) DOT road segment length.
const DefaultSegmentLengthKm = 0.1609344

// Risk threshold constants defining hazard classifications.
const (
	HazardThreshold       = 0.4 // Minimum risk score to be considered a hazard zone (Moderate)
	HighRiskThreshold     = 0.6 // Risk score threshold for High risk classification
	CriticalRiskThreshold = 0.8 // Risk score threshold for Critical risk classification
)

// Waypoint represents a geographic coordinate along a travel corridor.
type Waypoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// HazardZone represents a highway road segment with elevated wildlife collision risk.
type HazardZone struct {
	SegmentID   string  `json:"segment_id"`
	HighwayRef  string  `json:"highway_ref"`
	State       string  `json:"state"`
	Milepost    float64 `json:"milepost"`
	RiskScore   float64 `json:"risk_score"`
	RiskTier    string  `json:"risk_tier"`    // "Moderate", "High", "Critical"
	PrimaryRisk string  `json:"primary_risk"` // e.g. "Ungulate Migration Corridor", "Dense Canopy & Low Visibility"
}

// RouteRiskReport summarizes overall wildlife collision risk and hazard zones along a route.
type RouteRiskReport struct {
	TotalDistanceKm    float64      `json:"total_distance_km"`
	OverallRiskTier    string       `json:"overall_risk_tier"` // "Low", "Moderate", "High", "Severe"
	AverageRiskScore   float64      `json:"average_risk_score"`
	PeakRiskScore      float64      `json:"peak_risk_score"`
	DiurnalMultiplier  float64      `json:"diurnal_multiplier"`
	IsPeakWildlifeHour bool         `json:"is_peak_wildlife_hour"`
	HazardZones        []HazardZone `json:"hazard_zones"`
	AdvisoryNotice     string       `json:"advisory_notice"`
}

// CalculateDiurnalMultiplier returns the wildlife activity risk multiplier and whether
// the specified hour represents peak ungulate movement rhythm.
// Based on peer-reviewed ungulate activity rhythms (Huijser et al. 2009, Santos et al. 2011):
// - Dawn (5:00 - 8:59): 2.1x multiplier, peak = true
// - Dusk (18:00 - 22:59): 2.8x multiplier, peak = true
// - Night (23:00 - 4:59): 1.5x multiplier, peak = false
// - Day (9:00 - 17:59): 1.0x multiplier, peak = false
func CalculateDiurnalMultiplier(hourOfDay int) (multiplier float64, isPeak bool) {
	h := ((hourOfDay % 24) + 24) % 24
	switch {
	case h >= 5 && h <= 8:
		return 2.1, true
	case h >= 18 && h <= 22:
		return 2.8, true
	case h >= 9 && h <= 17:
		return 1.0, false
	default:
		return 1.5, false
	}
}

// DiurnalPeriodName returns the ungulate diurnal period name for a given hour of day.
func DiurnalPeriodName(hourOfDay int) string {
	h := ((hourOfDay % 24) + 24) % 24
	switch {
	case h >= 5 && h <= 8:
		return "Dawn"
	case h >= 18 && h <= 22:
		return "Dusk"
	case h >= 9 && h <= 17:
		return "Day"
	default:
		return "Night"
	}
}

// ClassifyHazardTier categorizes an individual segment risk score into hazard tiers.
func ClassifyHazardTier(score float64) string {
	switch {
	case score >= CriticalRiskThreshold:
		return "Critical"
	case score >= HighRiskThreshold:
		return "High"
	case score >= HazardThreshold:
		return "Moderate"
	default:
		return "Low"
	}
}

// DetermineOverallRiskTier calculates the overall journey risk tier given average and peak risk.
func DetermineOverallRiskTier(avgRisk, peakRisk float64) string {
	if avgRisk <= 0 && peakRisk <= 0 {
		return "Low"
	}
	switch {
	case peakRisk >= CriticalRiskThreshold || avgRisk >= 0.75:
		return "Severe"
	case peakRisk >= HighRiskThreshold || avgRisk >= 0.45:
		return "High"
	case peakRisk >= HazardThreshold || avgRisk >= 0.25:
		return "Moderate"
	default:
		return "Low"
	}
}

// EvaluateRouteRisk assesses wildlife collision hazard across a slice of road segments
// at a specific travel hour. It computes average and peak risk, determines the overall
// risk tier, extracts and sorts top hazard zones, applies the diurnal factor, and constructs
// a plain-language driver advisory.
func EvaluateRouteRisk(segments []HazardZone, hourOfDay int) RouteRiskReport {
	multiplier, isPeak := CalculateDiurnalMultiplier(hourOfDay)

	if len(segments) == 0 {
		return RouteRiskReport{
			TotalDistanceKm:    0.0,
			OverallRiskTier:    "Low",
			AverageRiskScore:   0.0,
			PeakRiskScore:      0.0,
			DiurnalMultiplier:  multiplier,
			IsPeakWildlifeHour: isPeak,
			HazardZones:        []HazardZone{},
			AdvisoryNotice:     "No road segments analyzed. Maintain standard driving vigilance.",
		}
	}

	totalDistKm := math.Round(float64(len(segments))*DefaultSegmentLengthKm*100) / 100

	var (
		sumScore    float64
		peakScore   float64
		hazardZones = make([]HazardZone, 0)
	)

	for _, seg := range segments {
		rawScore := seg.RiskScore
		if rawScore < 0 {
			rawScore = 0
		}

		adjustedScore := rawScore * multiplier
		sumScore += adjustedScore
		if adjustedScore > peakScore {
			peakScore = adjustedScore
		}

		if adjustedScore >= HazardThreshold {
			tier := ClassifyHazardTier(adjustedScore)
			hz := seg
			hz.RiskScore = math.Round(adjustedScore*1000) / 1000
			hz.RiskTier = tier
			hazardZones = append(hazardZones, hz)
		}
	}

	avgScore := sumScore / float64(len(segments))
	roundedAvg := math.Round(avgScore*1000) / 1000
	roundedPeak := math.Round(peakScore*1000) / 1000

	sort.Slice(hazardZones, func(i, j int) bool {
		if hazardZones[i].RiskScore != hazardZones[j].RiskScore {
			return hazardZones[i].RiskScore > hazardZones[j].RiskScore
		}
		if hazardZones[i].Milepost != hazardZones[j].Milepost {
			return hazardZones[i].Milepost < hazardZones[j].Milepost
		}
		return hazardZones[i].SegmentID < hazardZones[j].SegmentID
	})

	overallTier := DetermineOverallRiskTier(roundedAvg, roundedPeak)
	advisory := buildAdvisoryNotice(overallTier, hazardZones, hourOfDay, multiplier)

	return RouteRiskReport{
		TotalDistanceKm:    totalDistKm,
		OverallRiskTier:    overallTier,
		AverageRiskScore:   roundedAvg,
		PeakRiskScore:      roundedPeak,
		DiurnalMultiplier:  multiplier,
		IsPeakWildlifeHour: isPeak,
		HazardZones:        hazardZones,
		AdvisoryNotice:     advisory,
	}
}

// buildAdvisoryNotice formats a human-readable driver advisory based on risk metrics.
func buildAdvisoryNotice(overallTier string, hazardZones []HazardZone, hourOfDay int, multiplier float64) string {
	period := DiurnalPeriodName(hourOfDay)

	// Determine highway references
	highways := extractUniqueHighways(hazardZones)
	highwayStr := "along the route"
	if len(highways) == 1 {
		highwayStr = fmt.Sprintf("along %s", highways[0])
	} else if len(highways) == 2 {
		highwayStr = fmt.Sprintf("along %s and %s", highways[0], highways[1])
	} else if len(highways) > 2 {
		highwayStr = fmt.Sprintf("along %s, and %s", strings.Join(highways[:len(highways)-1], ", "), highways[len(highways)-1])
	}

	// 1. Alert prefix and sentence 1
	var prefix string
	var zoneSummary string

	switch overallTier {
	case "Severe", "High":
		prefix = "Caution:"
	case "Moderate":
		prefix = "Notice:"
	default:
		prefix = "Normal conditions:"
	}

	numZones := len(hazardZones)
	if numZones == 0 {
		zoneSummary = "No elevated wildlife collision zones identified along the route."
	} else if numZones == 1 {
		zoneSummary = fmt.Sprintf("1 high-risk wildlife collision zone identified %s.", highwayStr)
	} else {
		zoneSummary = fmt.Sprintf("%d high-risk wildlife collision zones identified %s.", numZones, highwayStr)
	}

	// 2. Diurnal factor statement
	var diurnalStatement string
	switch period {
	case "Dawn":
		diurnalStatement = fmt.Sprintf("Dawn travel increases animal crossing activity by %.1fx.", multiplier)
	case "Dusk":
		diurnalStatement = fmt.Sprintf("Dusk travel increases animal crossing activity by %.1fx.", multiplier)
	case "Night":
		diurnalStatement = fmt.Sprintf("Night travel increases animal crossing activity by %.1fx.", multiplier)
	default:
		diurnalStatement = "Daytime travel represents baseline animal crossing activity."
	}

	// 3. Recommended driver action
	var actionRecommendation string
	if numZones == 0 {
		actionRecommendation = "Maintain posted speeds and standard road vigilance."
	} else {
		hasCanyon := false
		hasCanopy := false
		for _, z := range hazardZones {
			lower := strings.ToLower(z.PrimaryRisk)
			if strings.Contains(lower, "canyon") {
				hasCanyon = true
			}
			if strings.Contains(lower, "canopy") || strings.Contains(lower, "visibility") {
				hasCanopy = true
			}
		}

		if hasCanyon || overallTier == "Severe" || overallTier == "High" {
			actionRecommendation = "Reduce speed to 45 mph through canyon sections."
		} else if hasCanopy {
			actionRecommendation = "Reduce speed through dense canopy sections and scan road shoulders vigilantly."
		} else {
			actionRecommendation = "Stay alert for wildlife movements near roadway edges."
		}
	}

	return fmt.Sprintf("%s %s %s %s", prefix, zoneSummary, diurnalStatement, actionRecommendation)
}

// extractUniqueHighways returns sorted unique highway identifiers from hazard zones.
func extractUniqueHighways(zones []HazardZone) []string {
	seen := make(map[string]bool)
	var list []string
	for _, z := range zones {
		ref := strings.TrimSpace(z.HighwayRef)
		if ref != "" && !seen[ref] {
			seen[ref] = true
			list = append(list, ref)
		}
	}
	sort.Strings(list)
	return list
}

// HaversineDistanceKm calculates the great-circle distance between two coordinates in kilometers.
func HaversineDistanceKm(p1, p2 Waypoint) float64 {
	const earthRadiusKm = 6371.0088
	lat1 := p1.Latitude * math.Pi / 180.0
	lat2 := p2.Latitude * math.Pi / 180.0
	dLat := (p2.Latitude - p1.Latitude) * math.Pi / 180.0
	dLon := (p2.Longitude - p1.Longitude) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// CalculatePolylineDistanceKm calculates total distance along a sequence of waypoints in kilometers.
func CalculatePolylineDistanceKm(waypoints []Waypoint) float64 {
	if len(waypoints) < 2 {
		return 0.0
	}
	var total float64
	for i := 0; i < len(waypoints)-1; i++ {
		total += HaversineDistanceKm(waypoints[i], waypoints[i+1])
	}
	return total
}
