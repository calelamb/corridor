// Package covariates provides spatial and environmental covariate calculators
// for highway wildlife collision risk modeling.
package covariates

import "math"

// MountainRegion identifies major mountain systems in the Western United States.
type MountainRegion string

const (
	RegionSierraNevada    MountainRegion = "Sierra Nevada"
	RegionCentralRockies  MountainRegion = "Central Rockies"
	RegionNorthernRockies MountainRegion = "Northern Rockies"
	RegionGeneralWestern  MountainRegion = "General Western US"
	RegionOutside         MountainRegion = "Outside Western US"
)

const (
	// DefaultAlpineClosureThresholdMeters is the default elevation threshold for seasonal pass closures.
	// At 2000 meters, this accommodates high-latitude Northern Rockies passes such as Logan Pass (2026m in Glacier NP),
	// while keeping valley floor roads (e.g. Jackson Hole at 1900m, Yosemite Valley at 1200m) open year-round.
	DefaultAlpineClosureThresholdMeters = 2000.0

	// HighElevationPassThresholdMeters is the standard threshold for Central Rockies and Sierra Nevada (2500m).
	HighElevationPassThresholdMeters = 2500.0
)

// GetMountainRegion determines the Western US mountain system for a coordinate pair.
func GetMountainRegion(lat, lon float64) MountainRegion {
	// Sierra Nevada: California / Nevada border
	if lat >= 35.0 && lat <= 41.5 && lon >= -122.5 && lon <= -117.5 {
		return RegionSierraNevada
	}
	// Northern Rockies: Northern Wyoming, Montana, Idaho, Glacier NP
	if lat >= 42.5 && lat <= 49.5 && lon >= -117.0 && lon <= -108.0 {
		return RegionNorthernRockies
	}
	// Central Rockies: Colorado, Utah, Southern Wyoming
	if lat >= 36.0 && lat < 42.5 && lon >= -112.5 && lon <= -104.5 {
		return RegionCentralRockies
	}
	// General Western US bounding box
	if lat >= 31.0 && lat <= 49.5 && lon >= -125.0 && lon <= -102.0 {
		return RegionGeneralWestern
	}
	return RegionOutside
}

// MonthlySnowFactor returns the seasonal snowpack accumulation and melt factor (0.0 to 1.0)
// for month (1 = January, ..., 12 = December). Peak snowpack occurs in March (month 3).
func MonthlySnowFactor(month int) float64 {
	switch month {
	case 1: // January
		return 0.80
	case 2: // February
		return 0.92
	case 3: // March (peak snowpack)
		return 1.00
	case 4: // April
		return 0.88
	case 5: // May (active melt)
		return 0.45
	case 6: // June (high alpine remnants)
		return 0.12
	case 7, 8, 9: // July, August, September (melted out)
		return 0.00
	case 10: // October (first high-elevation snowfall)
		return 0.10
	case 11: // November
		return 0.35
	case 12: // December
		return 0.60
	default:
		return 0.00
	}
}

// IsWinterMonth returns true for months November through May (months 11, 12, 1, 2, 3, 4, 5).
func IsWinterMonth(month int) bool {
	return month == 11 || month == 12 || (month >= 1 && month <= 5)
}

// IsShoulderMonth returns true for seasonal shoulder months (June and October).
func IsShoulderMonth(month int) bool {
	return month == 6 || month == 10
}

// GetAlpineElevationThreshold returns the elevation threshold in meters above which passes
// are seasonally closed in winter, adjusted for latitude.
// At lower latitudes (e.g. Sierra Nevada 37-38°N), alpine tree lines and pass closures occur at ~2500m.
// At higher latitudes (e.g. Glacier NP 48-49°N), the alpine zone drops to ~2000m.
func GetAlpineElevationThreshold(lat float64) float64 {
	if lat >= 45.0 {
		threshold := 2500.0 - (lat-38.0)*50.0
		if threshold < 1950.0 {
			threshold = 1950.0
		}
		return threshold
	}
	return HighElevationPassThresholdMeters
}

// EstimateSWE models monthly Snow Water Equivalent (SWE) in millimeters based on
// latitude, longitude, elevation (meters), and month (1-12) across Western US mountain systems.
func EstimateSWE(lat, lon, elevationMeters float64, month int) float64 {
	if month < 1 || month > 12 {
		return 0.0
	}
	region := GetMountainRegion(lat, lon)
	if region == RegionOutside {
		return 0.0
	}

	mFactor := MonthlySnowFactor(month)
	if mFactor <= 0.0 {
		return 0.0
	}

	var baseSnowline float64
	var sweRate float64

	switch region {
	case RegionSierraNevada:
		baseSnowline = 1400.0
		sweRate = 0.55
	case RegionCentralRockies:
		baseSnowline = 2000.0
		sweRate = 0.45
	case RegionNorthernRockies:
		baseSnowline = 1800.0 - (lat-40.0)*80.0
		if baseSnowline < 1000.0 {
			baseSnowline = 1000.0
		}
		sweRate = 0.60
	default:
		baseSnowline = 1800.0 - (lat-40.0)*70.0
		sweRate = 0.50
	}

	// Seasonal snowline shift during spring melt
	if month == 5 {
		baseSnowline += 300.0
	} else if month == 6 {
		baseSnowline += 600.0
	}

	if elevationMeters <= baseSnowline {
		return 0.0
	}

	effectiveElev := elevationMeters - baseSnowline
	if effectiveElev > 3000.0 {
		effectiveElev = 3000.0
	}

	swe := effectiveElev * sweRate * mFactor
	return math.Round(swe*100) / 100
}

// EstimateSnowDepth returns estimated snow depth in centimeters based on
// latitude, longitude, elevation (meters), and month (1-12).
// Models snow density differences between maritime Sierra snow and continental powder.
func EstimateSnowDepth(lat, lon, elevationMeters float64, month int) float64 {
	swe := EstimateSWE(lat, lon, elevationMeters, month)
	if swe <= 0.0 {
		return 0.0
	}

	region := GetMountainRegion(lat, lon)
	var densityFactor float64
	switch region {
	case RegionSierraNevada:
		densityFactor = 3.3 // ~0.33 g/cm^3 maritime snowpack ("Sierra cement")
	case RegionCentralRockies:
		densityFactor = 2.5 // ~0.25 g/cm^3 cold continental powder
	case RegionNorthernRockies:
		densityFactor = 2.8 // ~0.28 g/cm^3
	default:
		densityFactor = 3.0
	}

	depth := swe / densityFactor
	return math.Round(depth*10) / 10
}

// IsPassSeasonallyClosed returns true if an alpine pass is closed to vehicular traffic due to snow.
// Closes roads with elevation >= 2000m during winter months (November through May).
func IsPassSeasonallyClosed(elevationMeters float64, month int) bool {
	return IsWinterMonth(month) && elevationMeters >= DefaultAlpineClosureThresholdMeters
}

// IsPassSeasonallyClosedAt returns true if an alpine pass is closed to vehicular traffic due to snow,
// taking into account latitude-adjusted elevation thresholds.
func IsPassSeasonallyClosedAt(lat, lon, elevationMeters float64, month int) bool {
	if !IsWinterMonth(month) {
		return false
	}
	threshold := GetAlpineElevationThreshold(lat)
	return elevationMeters >= threshold
}

// AdjustTrafficForSnow scales traffic volume (AADT) for alpine snow conditions.
// Traffic is scaled to zero when the pass is seasonally closed (November through May at alpine elevations).
// During shoulder months (June and October), traffic is reduced due to snow clearing and shoulder closures.
// Valley floor roads and open summer months maintain full traffic volumes.
func AdjustTrafficForSnow(baseAADT float64, elevationMeters float64, month int) float64 {
	if baseAADT <= 0.0 {
		return 0.0
	}
	if month < 1 || month > 12 {
		return baseAADT
	}

	if IsPassSeasonallyClosed(elevationMeters, month) {
		return 0.0
	}

	// Shoulder month traffic reductions on alpine passes
	if elevationMeters >= DefaultAlpineClosureThresholdMeters {
		if month == 6 {
			// June shoulder: plowing operations, partial opening
			return baseAADT * 0.50
		}
		if month == 10 {
			// October shoulder: early snow events, pre-closure restrictions
			return baseAADT * 0.60
		}
	}

	return baseAADT
}

// AdjustTrafficForSnowAt scales traffic volume taking into account latitude-adjusted pass closure.
func AdjustTrafficForSnowAt(baseAADT float64, lat, lon, elevationMeters float64, month int) float64 {
	if baseAADT <= 0.0 {
		return 0.0
	}
	if month < 1 || month > 12 {
		return baseAADT
	}

	if IsPassSeasonallyClosedAt(lat, lon, elevationMeters, month) {
		return 0.0
	}

	threshold := GetAlpineElevationThreshold(lat)
	if elevationMeters >= threshold {
		if month == 6 {
			return baseAADT * 0.50
		}
		if month == 10 {
			return baseAADT * 0.60
		}
	}

	return baseAADT
}
