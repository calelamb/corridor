//go:build integration

package db

import (
	"math"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestRoadSegmentation(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// 1. Create a 1.0-mile highway LineString (1609.344 meters) in Wyoming near Yellowstone (US-191).
	var highwayGeom string
	err = pool.QueryRow(ctx, `
		SELECT ST_AsText(ST_MakeLine(
			ST_SetSRID(ST_Point(-110.8, 44.5), 4326),
			ST_Project(ST_SetSRID(ST_Point(-110.8, 44.5), 4326)::geography, 1609.344, 0)::geometry
		))
	`).Scan(&highwayGeom)
	if err != nil {
		t.Fatalf("failed to create 1.0-mile highway LineString: %v", err)
	}

	// 2. Call discretize_highway to split the 1.0-mile LineString into ~160.9 m (0.1-mile) segments.
	rows, err := pool.Query(ctx, `
		SELECT ST_AsText(geom), length_m, h3_r8::text
		FROM discretize_highway(ST_GeomFromText($1, 4326), 160.9344)
	`, highwayGeom)
	if err != nil {
		t.Fatalf("failed to execute discretize_highway: %v", err)
	}
	defer rows.Close()

	type segmentResult struct {
		geom    string
		lengthM float64
		h3R8    string
	}
	var segments []segmentResult
	for rows.Next() {
		var s segmentResult
		if err := rows.Scan(&s.geom, &s.lengthM, &s.h3R8); err != nil {
			t.Fatalf("failed to scan segment: %v", err)
		}
		segments = append(segments, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	// Verify that exactly 10 segments are produced for a 1.0-mile highway.
	if len(segments) != 10 {
		t.Fatalf("expected 10 segments, got %d", len(segments))
	}

	// Verify each segment length is ~160.9 m (+/- 2.0 m tolerance) and has valid H3-r8 index.
	for i, s := range segments {
		if math.Abs(s.lengthM-160.9344) > 2.0 {
			t.Errorf("segment %d length %.2f m out of tolerance (~160.9 m)", i, s.lengthM)
		}
		if s.h3R8 == "" {
			t.Errorf("segment %d has empty h3_r8", i)
		}

		// Verify H3 resolution is 8.
		var res int
		if err := pool.QueryRow(ctx, "SELECT h3_get_resolution($1::h3index)", s.h3R8).Scan(&res); err != nil {
			t.Errorf("segment %d invalid h3 index %s: %v", i, s.h3R8, err)
		} else if res != 8 {
			t.Errorf("segment %d expected H3 resolution 8, got %d", i, res)
		}

		// Verify H3 index matches centroid of segment geometry.
		var centroidH3 string
		if err := pool.QueryRow(ctx, "SELECT h3_lat_lng_to_cell(ST_Centroid(ST_GeomFromText($1, 4326)), 8)::text", s.geom).Scan(&centroidH3); err != nil {
			t.Errorf("segment %d centroid calculation failed: %v", i, err)
		} else if centroidH3 != s.h3R8 {
			t.Errorf("segment %d h3_r8 mismatch: got %s, centroid has %s", i, s.h3R8, centroidH3)
		}
	}

	// 3. Populate road_segments table and verify schema and constraints.
	for i, s := range segments {
		milepostStart := float64(i) * 0.1
		milepostEnd := float64(i+1) * 0.1
		osmWayIDs := []int64{1000000 + int64(i)}

		_, err := pool.Exec(ctx, `
			INSERT INTO road_segments(
				geom, state, highway_ref, milepost_start, milepost_end, length_m, h3_r8, osm_way_ids
			) VALUES (
				ST_GeomFromText($1, 4326), $2, $3, $4, $5, $6, $7::h3index, $8
			)
		`, s.geom, "WY", "US-191", milepostStart, milepostEnd, s.lengthM, s.h3R8, osmWayIDs)
		if err != nil {
			t.Fatalf("failed to insert segment %d into road_segments: %v", i, err)
		}
	}

	// Verify count and stored fields in road_segments.
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM road_segments WHERE highway_ref = 'US-191'").Scan(&count); err != nil {
		t.Fatalf("query road_segments count failed: %v", err)
	}
	if count != 10 {
		t.Fatalf("expected 10 road_segments, found %d", count)
	}

	// Verify query using H3 resolution-8 index.
	var h3Count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM road_segments WHERE h3_r8 = $1::h3index
	`, segments[0].h3R8).Scan(&h3Count); err != nil {
		t.Fatalf("query by h3_r8 failed: %v", err)
	}
	if h3Count < 1 {
		t.Errorf("expected at least 1 match for h3_r8 query, got %d", h3Count)
	}

	// Verify spatial index query (ST_DWithin).
	var spatialCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM road_segments
		WHERE ST_DWithin(geom::geography, ST_SetSRID(ST_Point(-110.8, 44.5), 4326)::geography, 250)
	`).Scan(&spatialCount); err != nil {
		t.Fatalf("spatial query failed: %v", err)
	}
	if spatialCount < 1 {
		t.Errorf("expected spatial match near starting point, got %d", spatialCount)
	}

	// 4. Test automatic H3 computation via trigger when inserting without explicit h3_r8.
	var autoH3 string
	err = pool.QueryRow(ctx, `
		INSERT INTO road_segments(
			geom, state, highway_ref, milepost_start, milepost_end, length_m
		) VALUES (
			ST_GeomFromText($1, 4326), 'WY', 'US-89', 0.0, 0.1, 160.9
		) RETURNING h3_r8::text
	`, segments[0].geom).Scan(&autoH3)
	if err != nil {
		t.Fatalf("failed to insert road_segment with auto-computed h3_r8: %v", err)
	}
	if autoH3 != segments[0].h3R8 {
		t.Errorf("auto-computed h3_r8 mismatch: expected %s, got %s", segments[0].h3R8, autoH3)
	}

	// 5. Test default parameter of discretize_highway (1 argument).
	var defaultSegCount int
	err = pool.QueryRow(ctx, `
		SELECT count(*) FROM discretize_highway(ST_GeomFromText($1, 4326))
	`, highwayGeom).Scan(&defaultSegCount)
	if err != nil {
		t.Fatalf("failed to call discretize_highway with default length: %v", err)
	}
	if defaultSegCount != 10 {
		t.Errorf("expected 10 segments with default 160.9344m parameter, got %d", defaultSegCount)
	}

	// 6. Test short highway (< 160.9 m) returning exactly 1 segment.
	var shortHighwayGeom string
	err = pool.QueryRow(ctx, `
		SELECT ST_AsText(ST_MakeLine(
			ST_SetSRID(ST_Point(-110.8, 44.5), 4326),
			ST_Project(ST_SetSRID(ST_Point(-110.8, 44.5), 4326)::geography, 50.0, 0)::geometry
		))
	`).Scan(&shortHighwayGeom)
	if err != nil {
		t.Fatalf("failed to create 50m highway: %v", err)
	}
	var shortSegCount int
	var shortLength float64
	err = pool.QueryRow(ctx, `
		SELECT count(*), coalesce(sum(length_m), 0)
		FROM discretize_highway(ST_GeomFromText($1, 4326), 160.9344)
	`, shortHighwayGeom).Scan(&shortSegCount, &shortLength)
	if err != nil {
		t.Fatalf("short highway discretize failed: %v", err)
	}
	if shortSegCount != 1 {
		t.Errorf("expected 1 segment for 50m highway, got %d", shortSegCount)
	}
	if math.Abs(shortLength-50.0) > 1.0 {
		t.Errorf("expected ~50m length, got %.2f", shortLength)
	}

	// 7. Test invalid parameter (non-positive segment length) raises error.
	_, err = pool.Exec(ctx, "SELECT * FROM discretize_highway(ST_GeomFromText($1, 4326), -10.0)", highwayGeom)
	if err == nil {
		t.Fatal("expected error for negative segment_length_m, got nil")
	}

	// 8. Test role permissions: corridor_public cannot read road_segments or call discretize_highway.
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Errorf("conn cleanup failed: %v", err)
		}
	}()
	if _, err := conn.Exec(ctx, "SET ROLE corridor_public"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, "SELECT * FROM road_segments"); err == nil {
		t.Fatal("corridor_public should NOT be able to read road_segments")
	}
	if _, err := conn.Exec(ctx, "SELECT * FROM discretize_highway(ST_GeomFromText('LINESTRING(-110.8 44.5, -110.8 44.6)', 4326), 160.9)"); err == nil {
		t.Fatal("corridor_public should NOT be able to execute discretize_highway")
	}
}
