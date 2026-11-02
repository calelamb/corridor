// Package geo provides PostGIS spatial operations, coordinate projections,
// and geospatial helpers for wildlife hazard corridors.
package geo

import (
	"errors"
	"math"
)

// GRS 1980 / WGS 84 ellipsoid parameters
const (
	grs80A    = 6378137.0         // Semi-major axis (meters)
	grs80InvF = 298.257222101     // Inverse flattening (1/f)
	grs80F    = 1.0 / grs80InvF   // Flattening
	grs80E2   = 2.0*grs80F - grs80F*grs80F // First eccentricity squared: e^2
)

// Montana State Plane NAD83 (EPSG:32100) Lambert Conformal Conic (2SP) parameters
const (
	mtLat1 = 45.0 * math.Pi / 180.0   // 45° 00' N (Standard Parallel 1)
	mtLat2 = 49.0 * math.Pi / 180.0   // 49° 00' N (Standard Parallel 2)
	mtLat0 = 44.25 * math.Pi / 180.0  // 44° 15' N (Latitude of False Origin)
	mtLon0 = -109.5 * math.Pi / 180.0 // 109° 30' W (Central Meridian / Longitude of False Origin)
	mtX0   = 600000.0                // False Easting (meters)
	mtY0   = 0.0                     // False Northing (meters)
)

var (
	grs80E   float64 // First eccentricity e = sqrt(e^2)
	mtN      float64 // Cone constant n
	mtF      float64 // Scale factor F
	mtRho0   float64 // Mapping radius at latitude of origin rho_0
)

func init() {
	grs80E = math.Sqrt(grs80E2)

	m := func(phi float64) float64 {
		s := math.Sin(phi)
		return math.Cos(phi) / math.Sqrt(1.0-grs80E2*s*s)
	}

	t := func(phi float64) float64 {
		s := math.Sin(phi)
		con := grs80E * s
		return math.Tan(math.Pi/4.0-phi/2.0) / math.Pow((1.0-con)/(1.0+con), grs80E/2.0)
	}

	m1 := m(mtLat1)
	m2 := m(mtLat2)
	t1 := t(mtLat1)
	t2 := t(mtLat2)
	t0 := t(mtLat0)

	mtN = (math.Log(m1) - math.Log(m2)) / (math.Log(t1) - math.Log(t2))
	mtF = m1 / (mtN * math.Pow(t1, mtN))
	mtRho0 = grs80A * mtF * math.Pow(t0, mtN)
}

// StatePlane32100ToWGS84 converts Montana State Plane NAD83 (EPSG:32100, Lambert Conformal Conic 2SP)
// Easting and Northing coordinates (in meters) to WGS84 decimal latitude and longitude (EPSG:4326).
// This is a pure Go implementation without CGo dependencies using standard ellipsoidal projection formulas.
func StatePlane32100ToWGS84(easting, northing float64) (lat, lon float64, err error) {
	if math.IsNaN(easting) || math.IsInf(easting, 0) || math.IsNaN(northing) || math.IsInf(northing, 0) {
		return 0, 0, errors.New("coordinates must be finite numbers")
	}

	// Montana easting/northing sanity range check (State of Montana bounds + 100km buffer)
	// Easting typically 100,000 to 1,100,000 m; Northing typically -50,000 to 600,000 m
	if easting < -500000 || easting > 2000000 || northing < -500000 || northing > 1500000 {
		return 0, 0, errors.New("easting or northing out of valid state plane range")
	}

	dx := easting - mtX0
	dy := mtRho0 - (northing - mtY0)
	rho := math.Copysign(math.Hypot(dx, dy), mtN)
	if rho == 0 {
		return math.Copysign(90.0, mtN), mtLon0 * 180.0 / math.Pi, nil
	}

	theta := math.Atan2(dx, dy)
	lonRad := mtLon0 + theta/mtN
	tVal := math.Pow(rho/(grs80A*mtF), 1.0/mtN)

	// Iterative solution for latitude phi (Snyder USGS PP 1395)
	phi := math.Pi/2.0 - 2.0*math.Atan(tVal)
	const maxIter = 15
	const tol = 1e-12
	converged := false
	for i := 0; i < maxIter; i++ {
		con := grs80E * math.Sin(phi)
		phiNext := math.Pi/2.0 - 2.0*math.Atan(tVal*math.Pow((1.0-con)/(1.0+con), grs80E/2.0))
		if math.Abs(phiNext-phi) < tol {
			phi = phiNext
			converged = true
			break
		}
		phi = phiNext
	}
	if !converged {
		return 0, 0, errors.New("latitude calculation failed to converge")
	}

	lat = phi * 180.0 / math.Pi
	lon = lonRad * 180.0 / math.Pi

	// Normalize longitude to [-180, 180]
	for lon > 180.0 {
		lon -= 360.0
	}
	for lon < -180.0 {
		lon += 360.0
	}

	return lat, lon, nil
}

// WGS84ToStatePlane32100 converts WGS84 decimal latitude and longitude (EPSG:4326)
// to Montana State Plane NAD83 (EPSG:32100) Easting and Northing (in meters).
func WGS84ToStatePlane32100(lat, lon float64) (easting, northing float64, err error) {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || math.IsNaN(lon) || math.IsInf(lon, 0) {
		return 0, 0, errors.New("coordinates must be finite numbers")
	}
	if lat < -90.0 || lat > 90.0 || lon < -180.0 || lon > 180.0 {
		return 0, 0, errors.New("coordinates out of WGS 84 bounds")
	}

	phi := lat * math.Pi / 180.0
	lam := lon * math.Pi / 180.0
	theta := mtN * (lam - mtLon0)

	s := math.Sin(phi)
	con := grs80E * s
	tVal := math.Tan(math.Pi/4.0-phi/2.0) / math.Pow((1.0-con)/(1.0+con), grs80E/2.0)
	rho := grs80A * mtF * math.Pow(tVal, mtN)

	easting = mtX0 + rho*math.Sin(theta)
	northing = mtY0 + mtRho0 - rho*math.Cos(theta)

	return easting, northing, nil
}
