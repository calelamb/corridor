package geo

import (
	"math"
	"testing"
)

func TestStatePlane32100ToWGS84_Benchmarks(t *testing.T) {
	benchmarks := []struct {
		name     string
		lat      float64
		lon      float64
		easting  float64
		northing float64
	}{
		{
			name:     "Helena, MT",
			lat:      46.589100,
			lon:      -112.039100,
			easting:  405557.683,
			northing: 263078.179,
		},
		{
			name:     "Bozeman, MT",
			lat:      45.677000,
			lon:      -111.042900,
			easting:  479836.847,
			northing: 159775.029,
		},
		{
			name:     "Missoula, MT",
			lat:      46.872100,
			lon:      -113.994000,
			easting:  257784.456,
			northing: 301187.545,
		},
		{
			name:     "Billings, MT",
			lat:      45.783300,
			lon:      -108.500700,
			easting:  677679.089,
			northing: 170897.584,
		},
	}

	for _, bm := range benchmarks {
		t.Run(bm.name, func(t *testing.T) {
			// Inverse projection: State Plane -> WGS84
			gotLat, gotLon, err := StatePlane32100ToWGS84(bm.easting, bm.northing)
			if err != nil {
				t.Fatalf("unexpected error converting state plane to wgs84: %v", err)
			}

			// Tolerance within 1e-5 degrees (~1 meter ground resolution)
			if math.Abs(gotLat-bm.lat) > 1e-5 {
				t.Errorf("latitude mismatch: got %.6f, want %.6f (diff: %.8f)", gotLat, bm.lat, math.Abs(gotLat-bm.lat))
			}
			if math.Abs(gotLon-bm.lon) > 1e-5 {
				t.Errorf("longitude mismatch: got %.6f, want %.6f (diff: %.8f)", gotLon, bm.lon, math.Abs(gotLon-bm.lon))
			}

			// Forward projection: WGS84 -> State Plane
			gotEasting, gotNorthing, err := WGS84ToStatePlane32100(bm.lat, bm.lon)
			if err != nil {
				t.Fatalf("unexpected error converting wgs84 to state plane: %v", err)
			}

			// Tolerance within 0.05 meters
			if math.Abs(gotEasting-bm.easting) > 0.05 {
				t.Errorf("easting mismatch: got %.3f, want %.3f (diff: %.4f)", gotEasting, bm.easting, math.Abs(gotEasting-bm.easting))
			}
			if math.Abs(gotNorthing-bm.northing) > 0.05 {
				t.Errorf("northing mismatch: got %.3f, want %.3f (diff: %.4f)", gotNorthing, bm.northing, math.Abs(gotNorthing-bm.northing))
			}
		})
	}
}

func TestStatePlane32100ToWGS84_InvalidCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		easting  float64
		northing float64
	}{
		{"NaN easting", math.NaN(), 100000},
		{"NaN northing", 500000, math.NaN()},
		{"Inf easting", math.Inf(1), 100000},
		{"Inf northing", 500000, math.Inf(-1)},
		{"Extreme negative easting", -1000000, 100000},
		{"Extreme positive easting", 3000000, 100000},
		{"Extreme negative northing", 500000, -1000000},
		{"Extreme positive northing", 500000, 2000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := StatePlane32100ToWGS84(tt.easting, tt.northing)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}
