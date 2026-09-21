package ingest

import (
	"math"
	"testing"

	"github.com/jonas-p/go-shp"
)

func TestMigrationGeometry(t *testing.T) {
	line := &shp.PolyLine{NumParts: 1, NumPoints: 2, Parts: []int32{0}, Points: []shp.Point{{X: 1, Y: 2}, {X: 3, Y: 4}}}
	b, err := MigrationGeometry(line)
	if err != nil || len(b) == 0 {
		t.Fatal(err)
	}
	invalid := &shp.PolyLine{NumParts: 1, NumPoints: 2, Parts: []int32{0}, Points: []shp.Point{{X: math.NaN(), Y: 2}, {X: 3, Y: 4}}}
	if _, err := MigrationGeometry(invalid); err == nil {
		t.Fatal("nonfinite shape accepted")
	}
	if _, err := MigrationGeometry(&shp.Point{X: 1, Y: 2}); err == nil {
		t.Fatal("point presented as migration route")
	}
}
