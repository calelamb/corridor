package geo_test

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/paulmach/orb"

	"corridor/internal/geo"
)

// FuzzSnapToSegment exercises the highway snapping engine across valid, boundary,
// and corrupt coordinate inputs to verify that it never panics, never produces NaN
// distances or coordinates, and handles all spatial edge cases cleanly.
func FuzzSnapToSegment(f *testing.F) {
	// Seed with valid and boundary coordinates
	f.Add(-110.5, 44.5, 50.0)                       // Yellowstone area
	f.Add(-110.7000, 44.5020, 100.0)                // Yellowstone highway US-191 corridor
	f.Add(0.0, 0.0, 50.0)                           // Equator / Prime Meridian
	f.Add(0.0, 90.0, 50.0)                          // North Pole
	f.Add(0.0, -90.0, 50.0)                         // South Pole
	f.Add(180.0, 0.0, 50.0)                         // Antimeridian East
	f.Add(-180.0, 0.0, 50.0)                        // Antimeridian West
	f.Add(69.5, -44.5, 50.0)                        // Antipode to Yellowstone
	f.Add(-110.5, 44.5, 0.0)                        // Zero tolerance
	f.Add(-110.5, 44.5, -10.0)                      // Negative tolerance
	f.Add(-110.5, 44.5, 1e9)                        // Massive tolerance
	f.Add(0.0, 89.999999, 10.0)                     // Sub-pole boundary
	f.Add(0.0, 90.000001, 10.0)                     // Out of range lat
	f.Add(0.0, -90.000001, 10.0)                    // Out of range lat
	f.Add(180.000001, 0.0, 10.0)                    // Out of range lon
	f.Add(-180.000001, 0.0, 10.0)                   // Out of range lon
	f.Add(math.MaxFloat64, math.MaxFloat64, 50.0)   // Max floats
	f.Add(-math.MaxFloat64, -math.MaxFloat64, 50.0) // Min floats
	f.Add(math.SmallestNonzeroFloat64, 0.0, 50.0)   // Smallest nonzero float
	f.Add(math.NaN(), 44.5, 50.0)                   // NaN lon
	f.Add(-110.5, math.NaN(), 50.0)                 // NaN lat
	f.Add(math.Inf(1), 44.5, 50.0)                  // +Inf lon
	f.Add(-110.5, math.Inf(-1), 50.0)               // -Inf lat
	f.Add(-110.5, 44.5, math.NaN())                 // NaN tolerance
	f.Add(-110.5, 44.5, math.Inf(1))                // +Inf tolerance

	f.Fuzz(func(t *testing.T, lon float64, lat float64, tolerance float64) {
		point := orb.Point{lon, lat}

		// Edge case: assert nil context returns clean error without panic
		if _, _, _, err := geo.SnapToSegment(nil, testPool, point, tolerance); err == nil {
			t.Errorf("expected error for nil context")
		}

		// Edge case: assert nil pool returns clean error without panic
		if _, _, _, err := geo.SnapToSegment(testCtx, nil, point, tolerance); err == nil {
			t.Errorf("expected error for nil pool")
		}

		// Call SnapToSegment with active context and pool
		segID, snapDist, snappedPt, err := geo.SnapToSegment(testCtx, testPool, point, tolerance)
		if err != nil {
			// Cleanly handled error (out of bounds, non-positive or NaN tolerance,
			// or points exceeding tolerance ErrExceedsSnapTolerance)
			return
		}

		// If snapping succeeded:
		// 1. Must never produce NaN distances or infinite distances
		if math.IsNaN(snapDist) || math.IsInf(snapDist, 0) {
			t.Fatalf("produced NaN or Inf snap distance %v for point %v, tolerance %v", snapDist, point, tolerance)
		}

		// 2. Snap distance must be non-negative and within tolerance
		if snapDist < 0 {
			t.Fatalf("produced negative snap distance %v for point %v", snapDist, point)
		}
		if snapDist > tolerance {
			t.Fatalf("produced snap distance %v exceeding tolerance %v", snapDist, tolerance)
		}

		// 3. Snapped point must be valid finite coordinates within WGS 84 bounds
		if math.IsNaN(snappedPt[0]) || math.IsInf(snappedPt[0], 0) ||
			math.IsNaN(snappedPt[1]) || math.IsInf(snappedPt[1], 0) {
			t.Fatalf("produced NaN or Inf snapped coordinates %v", snappedPt)
		}
		if snappedPt[0] < -180 || snappedPt[0] > 180 || snappedPt[1] < -90 || snappedPt[1] > 90 {
			t.Fatalf("snapped point %v out of WGS 84 bounds", snappedPt)
		}

		// 4. Segment ID must be non-nil
		if segID == uuid.Nil {
			t.Fatalf("expected valid non-nil segment ID for successful snap")
		}
	})
}
