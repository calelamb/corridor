package ingest_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"corridor/internal/db"
	"corridor/internal/ingest"
)

var (
	testPoolOnce sync.Once
	testPool     *pgxpool.Pool
	testCtx      context.Context
	testInitErr  error
)

func setupTestDB(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()
	testPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		_ = cancel

		container, err := postgres.Run(ctx, "corridor-postgres:foundation",
			postgres.WithDatabase("corridor_test"),
			postgres.WithUsername("test_admin"),
			postgres.WithPassword("synthetic-test-only"),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			testInitErr = fmt.Errorf("start postgres container: %w", err)
			return
		}

		dsn, err := container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			testInitErr = fmt.Errorf("connection string: %w", err)
			return
		}

		if err := db.Migrate(ctx, dsn); err != nil {
			testInitErr = fmt.Errorf("migrate: %w", err)
			return
		}

		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			testInitErr = fmt.Errorf("open pool: %w", err)
			return
		}

		testCtx = ctx
		testPool = pool
	})

	if testInitErr != nil {
		t.Fatalf("test database initialization failed: %v", testInitErr)
	}
	return testCtx, testPool
}

func sampleYellowstoneJSON() string {
	// A 1.0-mile highway in Yellowstone (US-191) starting at milestone 15.0
	// 44.5000 to 44.51447 at lon -110.8000 is approximately 1609.34 meters (1.0 mile).
	// A second 1.0-mile highway for US-89 in Grand Teton starting at milepost 20.0
	return `{
  "remark": "",
  "elements": [
    {
      "type": "node",
      "id": 1001,
      "lat": 44.500000,
      "lon": -110.800000,
      "tags": {
        "highway": "milestone",
        "milepost": "15.0"
      }
    },
    {
      "type": "way",
      "id": 2001,
      "nodes": [1001, 1002],
      "tags": {
        "highway": "primary",
        "ref": "US 191",
        "name": "Yellowstone Highway",
        "state": "WY"
      },
      "geometry": [
        {"lat": 44.500000, "lon": -110.800000},
        {"lat": 44.507235, "lon": -110.800000},
        {"lat": 44.514470, "lon": -110.800000}
      ]
    },
    {
      "type": "way",
      "id": 3001,
      "nodes": [2001, 2002],
      "tags": {
        "highway": "trunk",
        "ref": "US 89",
        "name": "Grand Teton Highway",
        "state": "WY",
        "milepost_start": "20.0"
      },
      "geometry": [
        {"lat": 43.800000, "lon": -110.600000},
        {"lat": 43.807235, "lon": -110.600000},
        {"lat": 43.814470, "lon": -110.600000}
      ]
    }
  ]
}`
}

func TestIngestNPSCorridors(t *testing.T) {
	ctx, pool := setupTestDB(t)

	raw := sampleYellowstoneJSON()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "yellowstone-highways.json")
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("IngestNPSCorridors failed: %v", err)
	}

	if stats.WaysParsed != 2 {
		t.Errorf("expected 2 ways parsed, got %d", stats.WaysParsed)
	}
	if stats.SegmentsCreated < 20 {
		t.Errorf("expected at least 20 segments created, got %d", stats.SegmentsCreated)
	}
	if stats.TotalLengthM < 3000.0 {
		t.Errorf("expected total length > 3000m, got %.2f", stats.TotalLengthM)
	}

	// 1. Verify sources registration for provenance
	var sourceCount int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM sources WHERE name LIKE '%Yellowstone%'").Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query sources: %v", err)
	}
	if sourceCount < 1 {
		t.Errorf("expected Yellowstone source registered in sources table, found %d", sourceCount)
	}

	// 2. Verify source_artifacts registration
	digest := sha256.Sum256([]byte(raw))
	hashStr := hex.EncodeToString(digest[:])
	var artifactCount int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM source_artifacts WHERE sha256 = $1", hashStr).Scan(&artifactCount)
	if err != nil {
		t.Fatalf("failed to query source_artifacts: %v", err)
	}
	if artifactCount < 1 {
		t.Errorf("expected artifact registered with sha256 %s, found %d", hashStr, artifactCount)
	}

	// 3. Verify US-191 segment count and 0.1-mile slicing
	type segRow struct {
		geom          string
		lengthM       float64
		state         string
		highwayRef    string
		milepostStart float64
		milepostEnd   float64
		h3R8          string
		osmWayIDs     []int64
	}

	rows, err := pool.Query(ctx, `
		SELECT ST_AsText(geom), length_m, state, highway_ref, milepost_start, milepost_end, h3_r8::text, osm_way_ids
		FROM road_segments
		WHERE highway_ref = 'US-191'
		ORDER BY milepost_start ASC
	`)
	if err != nil {
		t.Fatalf("failed to query road_segments for US-191: %v", err)
	}
	defer rows.Close()

	var us191Segs []segRow
	for rows.Next() {
		var s segRow
		if err := rows.Scan(&s.geom, &s.lengthM, &s.state, &s.highwayRef, &s.milepostStart, &s.milepostEnd, &s.h3R8, &s.osmWayIDs); err != nil {
			t.Fatalf("failed to scan road_segment: %v", err)
		}
		us191Segs = append(us191Segs, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	if len(us191Segs) != 10 {
		t.Fatalf("expected 10 segments for 1.0-mile US-191 corridor, got %d", len(us191Segs))
	}

	// 4. Verify milepost calculations from reference node (start MP 15.0)
	for i, s := range us191Segs {
		expectedStart := math.Round((15.0+float64(i)*0.1)*100) / 100
		expectedEnd := math.Round((15.0+float64(i+1)*0.1)*100) / 100

		if math.Abs(s.milepostStart-expectedStart) > 0.01 {
			t.Errorf("segment %d milepost_start expected %.2f, got %.2f", i, expectedStart, s.milepostStart)
		}
		if math.Abs(s.milepostEnd-expectedEnd) > 0.01 {
			t.Errorf("segment %d milepost_end expected %.2f, got %.2f", i, expectedEnd, s.milepostEnd)
		}

		// Length should be approximately 160.9m (+/- 2.0m)
		if math.Abs(s.lengthM-160.9344) > 3.0 {
			t.Errorf("segment %d length %.2f m out of tolerance (~160.9 m)", i, s.lengthM)
		}

		if s.state != "WY" {
			t.Errorf("segment %d expected state WY, got %s", i, s.state)
		}

		if s.h3R8 == "" {
			t.Errorf("segment %d empty h3_r8", i)
		}

		if len(s.osmWayIDs) == 0 || s.osmWayIDs[0] != 2001 {
			t.Errorf("segment %d expected osm_way_ids [2001], got %v", i, s.osmWayIDs)
		}
	}

	// 5. Verify US-89 segments
	var us89Count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM road_segments WHERE highway_ref = 'US-89'").Scan(&us89Count)
	if err != nil {
		t.Fatalf("failed to query road_segments count for US-89: %v", err)
	}
	if us89Count != 10 {
		t.Fatalf("expected 10 segments for US-89, got %d", us89Count)
	}

	// 6. Test Idempotency: re-ingesting the exact same file should not duplicate records
	stats2, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("second IngestNPSCorridors call failed: %v", err)
	}
	if stats2.WaysParsed != 2 {
		t.Errorf("re-ingest expected 2 ways parsed, got %d", stats2.WaysParsed)
	}

	var final191Count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM road_segments WHERE highway_ref = 'US-191'").Scan(&final191Count)
	if err != nil {
		t.Fatalf("failed to query road_segments count after re-ingest: %v", err)
	}
	if final191Count != 10 {
		t.Fatalf("idempotency violation: expected 10 segments for US-191, got %d", final191Count)
	}
}

func TestNPSRoadsBoundingBoxValidation(t *testing.T) {
	ctx, pool := setupTestDB(t)

	// Way outside Yellowstone / Grand Teton corridor (lat 35.0, lon -80.0, e.g. North Carolina)
	outsideJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 9999,
      "tags": {
        "highway": "primary",
        "ref": "US 191"
      },
      "geometry": [
        {"lat": 35.000000, "lon": -80.000000},
        {"lat": 35.010000, "lon": -80.000000}
      ]
    }
  ]
}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "outside.json")
	if err := os.WriteFile(path, []byte(outsideJSON), 0600); err != nil {
		t.Fatal(err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("expected outside ways to be skipped without crash, got error: %v", err)
	}
	if stats.WaysParsed != 0 || stats.SegmentsCreated != 0 {
		t.Errorf("expected 0 ways/segments from outside corridor, got %d ways, %d segments", stats.WaysParsed, stats.SegmentsCreated)
	}
}

func TestNPSRoadsParsingValidation(t *testing.T) {
	ctx, pool := setupTestDB(t)
	tmpDir := t.TempDir()

	// 1. Missing file
	_, err := ingest.IngestNPSCorridors(ctx, pool, nil, filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	// 2. Malformed JSON
	badPath := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badPath, []byte("invalid-json{"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = ingest.IngestNPSCorridors(ctx, pool, nil, badPath)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// 3. Coordinate out of WGS 84 bounds
	invalidCoordJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 5555,
      "tags": {
        "highway": "primary",
        "ref": "US 191"
      },
      "geometry": [
        {"lat": 144.500000, "lon": -110.800000},
        {"lat": 44.510000, "lon": -110.800000}
      ]
    }
  ]
}`
	invalidPath := filepath.Join(tmpDir, "invalid-coord.json")
	if err := os.WriteFile(invalidPath, []byte(invalidCoordJSON), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = ingest.IngestNPSCorridors(ctx, pool, nil, invalidPath)
	if err == nil {
		t.Fatal("expected error for coordinate out of WGS84 bounds, got nil")
	}
}

func TestNPSRoadsMontanaCorridor(t *testing.T) {
	ctx, pool := setupTestDB(t)

	// US-89 north of Gardiner, MT (lat >= 45.0)
	montanaJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 8801,
      "tags": {
        "highway": "primary",
        "ref": "US 89",
        "name": "Gardiner North Corridor"
      },
      "geometry": [
        {"lat": 45.030000, "lon": -110.700000},
        {"lat": 45.037235, "lon": -110.700000},
        {"lat": 45.044470, "lon": -110.700000}
      ]
    }
  ]
}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "montana.json")
	if err := os.WriteFile(path, []byte(montanaJSON), 0600); err != nil {
		t.Fatal(err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("IngestNPSCorridors failed: %v", err)
	}
	if stats.WaysParsed != 1 || stats.SegmentsCreated != 10 {
		t.Fatalf("expected 1 way, 10 segments; got %d ways, %d segments", stats.WaysParsed, stats.SegmentsCreated)
	}

	var stateCount int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM road_segments WHERE osm_way_ids @> ARRAY[8801::bigint] AND state = 'MT'").Scan(&stateCount)
	if err != nil {
		t.Fatalf("query state failed: %v", err)
	}
	if stateCount != 10 {
		t.Errorf("expected 10 segments in state MT, got %d", stateCount)
	}
}

func TestNPSRoadsConsecutiveCorridorChaining(t *testing.T) {
	ctx, pool := setupTestDB(t)

	// Two connected ways for US-191 without explicit milepost tags:
	// way 1 should get 0.0 to 1.0 (10 segments)
	// way 2 should get 1.0 to 2.0 (10 segments)
	chainJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 7701,
      "tags": {
        "highway": "primary",
        "ref": "US 191"
      },
      "geometry": [
        {"lat": 44.600000, "lon": -111.000000},
        {"lat": 44.614470, "lon": -111.000000}
      ]
    },
    {
      "type": "way",
      "id": 7702,
      "tags": {
        "highway": "primary",
        "ref": "US 191"
      },
      "geometry": [
        {"lat": 44.614470, "lon": -111.000000},
        {"lat": 44.628940, "lon": -111.000000}
      ]
    }
  ]
}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "chain.json")
	if err := os.WriteFile(path, []byte(chainJSON), 0600); err != nil {
		t.Fatal(err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("IngestNPSCorridors failed: %v", err)
	}
	if stats.WaysParsed != 2 || stats.SegmentsCreated != 20 {
		t.Fatalf("expected 2 ways, 20 segments; got %d ways, %d segments", stats.WaysParsed, stats.SegmentsCreated)
	}

	var maxEnd float64
	err = pool.QueryRow(ctx, "SELECT max(milepost_end) FROM road_segments WHERE osm_way_ids @> ARRAY[7702::bigint]").Scan(&maxEnd)
	if err != nil {
		t.Fatalf("query max end failed: %v", err)
	}
	if math.Abs(maxEnd-2.0) > 0.1 {
		t.Errorf("expected chained corridor to reach milepost ~2.0, got %.2f", maxEnd)
	}
}

// TestNPSRoadsRefTagDoesNotOverrideMilepost verifies that route reference tag (ref="191" or ref="89")
// is never misinterpreted as starting milepost 191.0 or 89.0.
func TestNPSRoadsRefTagDoesNotOverrideMilepost(t *testing.T) {
	ctx, pool := setupTestDB(t)

	refJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 9901,
      "tags": {
        "highway": "primary",
        "ref": "191",
        "name": "Yellowstone Highway"
      },
      "geometry": [
        {"lat": 44.700000, "lon": -111.100000},
        {"lat": 44.714470, "lon": -111.100000}
      ]
    }
  ]
}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ref-test.json")
	if err := os.WriteFile(path, []byte(refJSON), 0600); err != nil {
		t.Fatal(err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("IngestNPSCorridors failed: %v", err)
	}
	if stats.WaysParsed != 1 || stats.SegmentsCreated != 10 {
		t.Fatalf("expected 1 way, 10 segments; got %d ways, %d segments", stats.WaysParsed, stats.SegmentsCreated)
	}

	var minStart, maxEnd float64
	err = pool.QueryRow(ctx, "SELECT min(milepost_start), max(milepost_end) FROM road_segments WHERE osm_way_ids @> ARRAY[9901::bigint]").Scan(&minStart, &maxEnd)
	if err != nil {
		t.Fatalf("query mileposts failed: %v", err)
	}

	// Must NOT be 191.0; should start at 0.0 (or corridor chain start) and end at ~1.0
	if minStart >= 190.0 {
		t.Fatalf("CRITICAL REGRESSION: way ref='191' was misinterpreted as starting milepost %.2f", minStart)
	}
	if math.Abs(minStart-0.0) > 0.05 {
		t.Errorf("expected starting milepost near 0.0, got %.2f", minStart)
	}
	if math.Abs(maxEnd-1.0) > 0.05 {
		t.Errorf("expected ending milepost near 1.0, got %.2f", maxEnd)
	}
}

// TestNPSRoadsPermissiveSubstringPrevention verifies that local streets like "1910 Pioneer Trail"
// or "89th Street" are not falsely matched as US-191 or US-89.
func TestNPSRoadsPermissiveSubstringPrevention(t *testing.T) {
	ctx, pool := setupTestDB(t)

	falsePositiveJSON := `{
  "elements": [
    {
      "type": "way",
      "id": 9902,
      "tags": {
        "highway": "residential",
        "name": "1910 Pioneer Trail"
      },
      "geometry": [
        {"lat": 44.500000, "lon": -110.800000},
        {"lat": 44.505000, "lon": -110.800000}
      ]
    },
    {
      "type": "way",
      "id": 9903,
      "tags": {
        "highway": "residential",
        "name": "89th Street"
      },
      "geometry": [
        {"lat": 44.500000, "lon": -110.800000},
        {"lat": 44.505000, "lon": -110.800000}
      ]
    }
  ]
}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "false-positives.json")
	if err := os.WriteFile(path, []byte(falsePositiveJSON), 0600); err != nil {
		t.Fatal(err)
	}

	stats, err := ingest.IngestNPSCorridors(ctx, pool, nil, path)
	if err != nil {
		t.Fatalf("IngestNPSCorridors failed: %v", err)
	}
	if stats.WaysParsed != 0 || stats.SegmentsCreated != 0 {
		t.Errorf("permissive substring bug: matched %d non-highway ways (%d segments)", stats.WaysParsed, stats.SegmentsCreated)
	}
}

func TestYosemiteBoundingBoxValidation(t *testing.T) {
	tests := []struct {
		name string
		lon  float64
		lat  float64
		want bool
	}{
		// Inside Yosemite corridor bounds [37.4, 38.3], [-119.95, -119.15]
		{"Crane Flat / Tioga Pass", -119.80, 37.85, true},
		{"Merced River Canyon", -119.75, 37.70, true},
		{"SW Boundary Corner", -119.95, 37.40, true},
		{"NE Boundary Corner", -119.15, 38.30, true},
		{"Yosemite Valley Central", -119.58, 37.74, true},
		{"Tuolumne Meadows", -119.35, 37.87, true},

		// Outside Yosemite bounds
		{"Yellowstone US-191", -110.80, 44.50, false},
		{"Grand Teton US-89", -110.60, 43.80, false},
		{"Too far West", -120.00, 37.85, false},
		{"Too far East", -119.10, 37.85, false},
		{"Too far South", -119.50, 37.35, false},
		{"Too far North", -119.50, 38.35, false},
		{"Out of state (North Carolina)", -80.00, 35.00, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ingest.InYosemiteCorridorBbox(tc.lon, tc.lat)
			if got != tc.want {
				t.Errorf("InYosemiteCorridorBbox(%.2f, %.2f) = %v; want %v", tc.lon, tc.lat, got, tc.want)
			}
		})
	}

	// Also verify combined corridor bbox allows both Yellowstone and Yosemite
	if !ingest.InCorridorBbox(-110.80, 44.50) {
		t.Errorf("InCorridorBbox should accept Yellowstone coordinate (-110.80, 44.50)")
	}
	if !ingest.InCorridorBbox(-119.80, 37.85) {
		t.Errorf("InCorridorBbox should accept Yosemite coordinate (-119.80, 37.85)")
	}
	if ingest.InCorridorBbox(-80.00, 35.00) {
		t.Errorf("InCorridorBbox should reject coordinate outside both corridors (-80.00, 35.00)")
	}
}

func TestCaliforniaStateDetermination(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		lat  float64
		lon  float64
		want string
	}{
		{
			name: "explicit state tag CA",
			tags: map[string]string{"state": "CA"},
			lat:  37.85,
			lon:  -119.80,
			want: "CA",
		},
		{
			name: "explicit lowercase state tag ca",
			tags: map[string]string{"state": "ca"},
			lat:  37.85,
			lon:  -119.80,
			want: "CA",
		},
		{
			name: "explicit addr:state tag CA",
			tags: map[string]string{"addr:state": "CA"},
			lat:  37.85,
			lon:  -119.80,
			want: "CA",
		},
		{
			name: "full state name California",
			tags: map[string]string{"state": "California"},
			lat:  37.85,
			lon:  -119.80,
			want: "CA",
		},
		{
			name: "coordinates within California bbox without tags",
			tags: map[string]string{},
			lat:  37.85,
			lon:  -119.80,
			want: "CA",
		},
		{
			name: "coordinates at CA bbox SW boundary",
			tags: map[string]string{},
			lat:  37.0,
			lon:  -121.0,
			want: "CA",
		},
		{
			name: "coordinates at CA bbox NE boundary",
			tags: map[string]string{},
			lat:  39.0,
			lon:  -118.0,
			want: "CA",
		},
		{
			name: "Montana latitude",
			tags: map[string]string{},
			lat:  45.5,
			lon:  -110.8,
			want: "MT",
		},
		{
			name: "Idaho coordinates",
			tags: map[string]string{},
			lat:  44.0,
			lon:  -111.2,
			want: "ID",
		},
		{
			name: "Yellowstone Wyoming coordinates",
			tags: map[string]string{},
			lat:  44.5,
			lon:  -110.8,
			want: "WY",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ingest.DetermineState(tc.tags, tc.lat, tc.lon)
			if got != tc.want {
				t.Errorf("DetermineState() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestYosemiteHighwayRefMatching(t *testing.T) {
	tests := []struct {
		name       string
		tags       map[string]string
		wantRef    string
		wantTarget bool
	}{
		// CA-120
		{
			name:       "CA-120 ref formatted",
			tags:       map[string]string{"highway": "primary", "ref": "CA-120"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "CA 120 ref spaced",
			tags:       map[string]string{"highway": "primary", "ref": "CA 120"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "120 plain ref",
			tags:       map[string]string{"highway": "primary", "ref": "120"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "Tioga Road name",
			tags:       map[string]string{"highway": "primary", "name": "Tioga Road"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "Big Oak Flat Road name",
			tags:       map[string]string{"highway": "primary", "name": "Big Oak Flat Road"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "Tioga Pass Road name",
			tags:       map[string]string{"highway": "trunk", "name": "Tioga Pass Road"},
			wantRef:    "CA-120",
			wantTarget: true,
		},
		{
			name:       "SR 120 ref",
			tags:       map[string]string{"highway": "primary", "ref": "SR 120"},
			wantRef:    "CA-120",
			wantTarget: true,
		},

		// CA-140
		{
			name:       "CA-140 ref formatted",
			tags:       map[string]string{"highway": "primary", "ref": "CA-140"},
			wantRef:    "CA-140",
			wantTarget: true,
		},
		{
			name:       "CA 140 ref spaced",
			tags:       map[string]string{"highway": "primary", "ref": "CA 140"},
			wantRef:    "CA-140",
			wantTarget: true,
		},
		{
			name:       "140 plain ref",
			tags:       map[string]string{"highway": "primary", "ref": "140"},
			wantRef:    "CA-140",
			wantTarget: true,
		},
		{
			name:       "El Portal Road name",
			tags:       map[string]string{"highway": "primary", "name": "El Portal Road"},
			wantRef:    "CA-140",
			wantTarget: true,
		},
		{
			name:       "Merced River Canyon Road name",
			tags:       map[string]string{"highway": "primary", "name": "Merced River Canyon Road"},
			wantRef:    "CA-140",
			wantTarget: true,
		},

		// CA-41
		{
			name:       "CA-41 ref formatted",
			tags:       map[string]string{"highway": "primary", "ref": "CA-41"},
			wantRef:    "CA-41",
			wantTarget: true,
		},
		{
			name:       "CA 41 ref spaced",
			tags:       map[string]string{"highway": "primary", "ref": "CA 41"},
			wantRef:    "CA-41",
			wantTarget: true,
		},
		{
			name:       "41 plain ref",
			tags:       map[string]string{"highway": "primary", "ref": "41"},
			wantRef:    "CA-41",
			wantTarget: true,
		},
		{
			name:       "Wawona Road name",
			tags:       map[string]string{"highway": "primary", "name": "Wawona Road"},
			wantRef:    "CA-41",
			wantTarget: true,
		},

		// False positives and non-target roads
		{
			name:       "120th Avenue residential street",
			tags:       map[string]string{"highway": "residential", "name": "120th Avenue"},
			wantRef:    "",
			wantTarget: false,
		},
		{
			name:       "140th Street residential road",
			tags:       map[string]string{"highway": "residential", "name": "140th Street"},
			wantRef:    "",
			wantTarget: false,
		},
		{
			name:       "4100 Pine Way address",
			tags:       map[string]string{"highway": "residential", "name": "4100 Pine Way"},
			wantRef:    "",
			wantTarget: false,
		},
		{
			name:       "Footway excluded even with highway ref",
			tags:       map[string]string{"highway": "footway", "ref": "CA-120"},
			wantRef:    "",
			wantTarget: false,
		},
		{
			name:       "Cycleway excluded",
			tags:       map[string]string{"highway": "cycleway", "name": "Tioga Road Trail"},
			wantRef:    "",
			wantTarget: false,
		},
		{
			name:       "Generic unrelated street",
			tags:       map[string]string{"highway": "secondary", "name": "Main Street"},
			wantRef:    "",
			wantTarget: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ref, ok := ingest.MatchHighwayRef(tc.tags)
			if ok != tc.wantTarget {
				t.Errorf("MatchHighwayRef(%v) target = %v; want %v", tc.tags, ok, tc.wantTarget)
			}
			if ref != tc.wantRef {
				t.Errorf("MatchHighwayRef(%v) ref = %q; want %q", tc.tags, ref, tc.wantRef)
			}
		})
	}
}
