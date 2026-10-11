package geo_test

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paulmach/orb"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"corridor/internal/db"
	"corridor/internal/geo"
)

var (
	testCtx  context.Context
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx, "corridor-postgres:foundation",
		postgres.WithDatabase("corridor_test"),
		postgres.WithUsername("test_admin"),
		postgres.WithPassword("synthetic-test-only"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer func() {
		_ = container.Terminate(context.Background())
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	if err := db.Migrate(ctx, dsn); err != nil {
		panic("failed to migrate db: " + err.Error())
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer pool.Close()

	testCtx = ctx
	testPool = pool

	os.Exit(m.Run())
}

// helper to insert a test road segment
func insertTestRoadSegment(t *testing.T, lineWKT string) uuid.UUID {
	t.Helper()
	var segID uuid.UUID
	err := testPool.QueryRow(testCtx, `
		INSERT INTO road_segments(geom, length_m, highway_ref)
		VALUES (
			ST_GeomFromText($1, 4326),
			ST_Length(ST_GeomFromText($1, 4326)::geography),
			'US-191'
		)
		RETURNING id
	`, lineWKT).Scan(&segID)
	if err != nil {
		t.Fatalf("failed to insert test road segment: %v", err)
	}
	return segID
}

// helper to insert a test species
func insertTestSpecies(t *testing.T, name string) uuid.UUID {
	t.Helper()
	var speciesID uuid.UUID
	err := testPool.QueryRow(testCtx, `
		INSERT INTO species(name, sensitive)
		VALUES ($1, false)
		ON CONFLICT(name) DO UPDATE SET sensitive=EXCLUDED.sensitive
		RETURNING id
	`, name).Scan(&speciesID)
	if err != nil {
		t.Fatalf("failed to insert test species: %v", err)
	}
	return speciesID
}

// helper to insert a test event
func insertTestEvent(t *testing.T, speciesID uuid.UUID, point orb.Point, observedAt time.Time, dedupGroupID *uuid.UUID) uuid.UUID {
	t.Helper()
	var eventID uuid.UUID
	err := testPool.QueryRow(testCtx, `
		INSERT INTO events(species_id, geom, observed_at, time_precision, uncertainty_m, dedup_group_id)
		VALUES (
			$1,
			ST_SetSRID(ST_MakePoint($2, $3), 4326),
			$4,
			'instant',
			10.0,
			$5
		)
		RETURNING id
	`, speciesID, point[0], point[1], observedAt, dedupGroupID).Scan(&eventID)
	if err != nil {
		t.Fatalf("failed to insert test event: %v", err)
	}
	return eventID
}

func TestSnapToSegment_WithinTolerance(t *testing.T) {
	// 1. Create a highway segment running North-South along longitude -110.8 from lat 44.500 to 44.505
	segID := insertTestRoadSegment(t, "LINESTRING(-110.8000 44.5000, -110.8000 44.5050)")

	// 2. Compute a point ~15 meters to the East of the highway at lat 44.5020 using PostGIS ST_Project
	var pointX, pointY float64
	err := testPool.QueryRow(testCtx, `
		SELECT ST_X(pt), ST_Y(pt)
		FROM (
			SELECT ST_Project(ST_SetSRID(ST_Point(-110.8000, 44.5020), 4326)::geography, 15.0, radians(90))::geometry AS pt
		) q
	`).Scan(&pointX, &pointY)
	if err != nil {
		t.Fatalf("failed to project test point: %v", err)
	}

	testPoint := orb.Point{pointX, pointY}

	// 3. Snap with 50.0m maxDistanceMeters
	snappedSegID, snapDist, snappedPt, err := geo.SnapToSegment(testCtx, testPool, testPoint, 50.0)
	if err != nil {
		t.Fatalf("expected successful snap within tolerance, got error: %v", err)
	}

	if snappedSegID != segID {
		t.Errorf("expected segment ID %s, got %s", segID, snappedSegID)
	}

	if math.Abs(snapDist-15.0) > 1.0 {
		t.Errorf("expected snap distance ~15.0m (+/- 1.0m), got %.2fm", snapDist)
	}

	// Snapped point must lie on the highway line: longitude should be exactly -110.8000
	if math.Abs(snappedPt[0]-(-110.8000)) > 1e-5 {
		t.Errorf("expected snapped point lon -110.8000, got %.6f", snappedPt[0])
	}
	if math.Abs(snappedPt[1]-44.5020) > 1e-4 {
		t.Errorf("expected snapped point lat ~44.5020, got %.6f", snappedPt[1])
	}
}

func TestSnapToSegment_ExceedsTolerance(t *testing.T) {
	// 1. Create a highway segment
	segID := insertTestRoadSegment(t, "LINESTRING(-110.9000 44.6000, -110.9000 44.6050)")

	// Get initial segment data to verify it is not modified
	var initialLength float64
	err := testPool.QueryRow(testCtx, "SELECT length_m FROM road_segments WHERE id = $1", segID).Scan(&initialLength)
	if err != nil {
		t.Fatalf("failed to get initial segment data: %v", err)
	}

	// 2. Compute a point ~150 meters to the East of the highway at lat 44.6020 (> 50m tolerance)
	var pointX, pointY float64
	err = testPool.QueryRow(testCtx, `
		SELECT ST_X(pt), ST_Y(pt)
		FROM (
			SELECT ST_Project(ST_SetSRID(ST_Point(-110.9000, 44.6020), 4326)::geography, 150.0, radians(90))::geometry AS pt
		) q
	`).Scan(&pointX, &pointY)
	if err != nil {
		t.Fatalf("failed to project 150m test point: %v", err)
	}

	testPoint := orb.Point{pointX, pointY}

	// 3. Snap with maxDistanceMeters = 50.0
	_, _, _, err = geo.SnapToSegment(testCtx, testPool, testPoint, 50.0)
	if err == nil {
		t.Fatalf("expected ErrExceedsSnapTolerance, got nil error")
	}

	if !errors.Is(err, geo.ErrExceedsSnapTolerance) {
		t.Fatalf("expected ErrExceedsSnapTolerance, got: %v", err)
	}

	// Verify segment data in DB was not corrupted or modified
	var afterLength float64
	err = testPool.QueryRow(testCtx, "SELECT length_m FROM road_segments WHERE id = $1", segID).Scan(&afterLength)
	if err != nil {
		t.Fatalf("failed to verify segment data after failed snap: %v", err)
	}
	if afterLength != initialLength {
		t.Errorf("road segment length changed from %.2f to %.2f", initialLength, afterLength)
	}
}

func TestDeduplicateEvents_SameSpeciesWithinThreshold(t *testing.T) {
	speciesID := insertTestSpecies(t, "Cervus canadensis - Elk")
	basePoint := orb.Point{-110.8500, 44.5500}
	t0 := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	initialGroupID := uuid.New()

	// Existing carcass record E1
	_ = insertTestEvent(t, speciesID, basePoint, t0, &initialGroupID)

	// Compute point 80m away from basePoint
	var point80X, point80Y float64
	err := testPool.QueryRow(testCtx, `
		SELECT ST_X(pt), ST_Y(pt)
		FROM (
			SELECT ST_Project(ST_SetSRID(ST_Point($1, $2), 4326)::geography, 80.0, radians(45))::geometry AS pt
		) q
	`, basePoint[0], basePoint[1]).Scan(&point80X, &point80Y)
	if err != nil {
		t.Fatalf("failed to project 80m point: %v", err)
	}
	point80 := orb.Point{point80X, point80Y}

	// New report 12 hours later (threshold: 200m, 72h)
	t1 := t0.Add(12 * time.Hour)
	dedupGroupID, isDuplicate, err := geo.DeduplicateEvents(testCtx, testPool, speciesID.String(), point80, t1, 200.0, 72.0)
	if err != nil {
		t.Fatalf("unexpected error during deduplication: %v", err)
	}

	if !isDuplicate {
		t.Errorf("expected isDuplicate=true for report within 80m and 12h")
	}

	if dedupGroupID == nil || *dedupGroupID != initialGroupID {
		t.Errorf("expected dedupGroupID to link to existing group %s, got %v", initialGroupID, dedupGroupID)
	}
}

func TestDeduplicateEvents_SixMonthsLater_IndependentEvent(t *testing.T) {
	speciesID := insertTestSpecies(t, "Odocoileus hemionus - Mule Deer")
	basePoint := orb.Point{-110.8600, 44.5600}
	t0 := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	initialGroupID := uuid.New()

	// Existing carcass record E1
	_ = insertTestEvent(t, speciesID, basePoint, t0, &initialGroupID)

	// Point 80m away
	var point80X, point80Y float64
	err := testPool.QueryRow(testCtx, `
		SELECT ST_X(pt), ST_Y(pt)
		FROM (
			SELECT ST_Project(ST_SetSRID(ST_Point($1, $2), 4326)::geography, 80.0, radians(45))::geometry AS pt
		) q
	`, basePoint[0], basePoint[1]).Scan(&point80X, &point80Y)
	if err != nil {
		t.Fatalf("failed to project 80m point: %v", err)
	}
	point80 := orb.Point{point80X, point80Y}

	// New report 6 months later (180 days)
	tLater := t0.Add(180 * 24 * time.Hour)
	dedupGroupID, isDuplicate, err := geo.DeduplicateEvents(testCtx, testPool, speciesID.String(), point80, tLater, 200.0, 72.0)
	if err != nil {
		t.Fatalf("unexpected error during deduplication: %v", err)
	}

	if isDuplicate {
		t.Errorf("expected isDuplicate=false for event 6 months later")
	}

	if dedupGroupID == nil || *dedupGroupID == initialGroupID {
		t.Errorf("expected new independent dedupGroupID, but got initial group ID %s", initialGroupID)
	}
}

func TestDeduplicateEvents_DifferentSpecies_IndependentEvent(t *testing.T) {
	speciesA := insertTestSpecies(t, "Ursus arctos horribilis")
	speciesB := insertTestSpecies(t, "Canis lupus")
	basePoint := orb.Point{-110.8700, 44.5700}
	t0 := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	groupA := uuid.New()

	// Existing carcass of Species A
	_ = insertTestEvent(t, speciesA, basePoint, t0, &groupA)

	// New report at same location, 1 hour later, but Species B
	dedupGroupID, isDuplicate, err := geo.DeduplicateEvents(testCtx, testPool, speciesB.String(), basePoint, t0.Add(1*time.Hour), 200.0, 72.0)
	if err != nil {
		t.Fatalf("unexpected error during deduplication: %v", err)
	}

	if isDuplicate {
		t.Errorf("expected isDuplicate=false for different species")
	}

	if dedupGroupID == nil || *dedupGroupID == groupA {
		t.Errorf("expected new independent group ID, got %v", dedupGroupID)
	}
}

func TestValidationBoundaries(t *testing.T) {
	// SnapToSegment input boundaries
	if _, _, _, err := geo.SnapToSegment(nil, testPool, orb.Point{0, 0}, 50.0); err == nil {
		t.Errorf("expected error for nil context")
	}
	if _, _, _, err := geo.SnapToSegment(testCtx, nil, orb.Point{0, 0}, 50.0); err == nil {
		t.Errorf("expected error for nil pool")
	}
	if _, _, _, err := geo.SnapToSegment(testCtx, testPool, orb.Point{200.0, 0.0}, 50.0); err == nil {
		t.Errorf("expected error for invalid longitude")
	}
	if _, _, _, err := geo.SnapToSegment(testCtx, testPool, orb.Point{0.0, 100.0}, 50.0); err == nil {
		t.Errorf("expected error for invalid latitude")
	}
	if _, _, _, err := geo.SnapToSegment(testCtx, testPool, orb.Point{math.NaN(), 0.0}, 50.0); err == nil {
		t.Errorf("expected error for NaN coordinate")
	}
	if _, _, _, err := geo.SnapToSegment(testCtx, testPool, orb.Point{0.0, 0.0}, -5.0); err == nil {
		t.Errorf("expected error for negative max distance")
	}

	// DeduplicateEvents input boundaries
	if _, _, err := geo.DeduplicateEvents(nil, testPool, "test-species", orb.Point{0, 0}, time.Now(), 200.0, 72.0); err == nil {
		t.Errorf("expected error for nil context")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, nil, "test-species", orb.Point{0, 0}, time.Now(), 200.0, 72.0); err == nil {
		t.Errorf("expected error for nil pool")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "", orb.Point{0.0, 0.0}, time.Now(), 200.0, 72.0); err == nil {
		t.Errorf("expected error for empty speciesID")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "test-species", orb.Point{0.0, 0.0}, time.Now(), -10.0, 72.0); err == nil {
		t.Errorf("expected error for negative distance threshold")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "test-species", orb.Point{0.0, 0.0}, time.Now(), 200.0, -1.0); err == nil {
		t.Errorf("expected error for negative time threshold")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "test-species", orb.Point{0.0, 0.0}, time.Time{}, 200.0, 72.0); err == nil {
		t.Errorf("expected error for zero observedAt")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "test-species", orb.Point{math.NaN(), 0.0}, time.Now(), 200.0, 72.0); err == nil {
		t.Errorf("expected error for NaN coordinate")
	}
	if _, _, err := geo.DeduplicateEvents(testCtx, testPool, "test-species", orb.Point{250.0, 0.0}, time.Now(), 200.0, 72.0); err == nil {
		t.Errorf("expected error for out of bounds coordinate")
	}
}

func TestSnapToSegment_MultipleSegments_PicksClosest(t *testing.T) {
	// Segment A: along longitude -110.7000
	segA := insertTestRoadSegment(t, "LINESTRING(-110.7000 44.5000, -110.7000 44.5050)")
	// Segment B: along longitude -110.7010 (~80m further west)
	_ = insertTestRoadSegment(t, "LINESTRING(-110.7010 44.5000, -110.7010 44.5050)")

	// Point close to Segment A (~10m east of Segment A)
	var pointX, pointY float64
	err := testPool.QueryRow(testCtx, `
		SELECT ST_X(pt), ST_Y(pt)
		FROM (
			SELECT ST_Project(ST_SetSRID(ST_Point(-110.7000, 44.5020), 4326)::geography, 10.0, radians(90))::geometry AS pt
		) q
	`).Scan(&pointX, &pointY)
	if err != nil {
		t.Fatalf("failed to project point: %v", err)
	}

	snappedSegID, snapDist, _, err := geo.SnapToSegment(testCtx, testPool, orb.Point{pointX, pointY}, 100.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snappedSegID != segA {
		t.Errorf("expected snap to closest segment %s, got %s", segA, snappedSegID)
	}
	if math.Abs(snapDist-10.0) > 1.0 {
		t.Errorf("expected snap dist ~10m, got %.2fm", snapDist)
	}
}

func TestDeduplicateEvents_BySpeciesName(t *testing.T) {
	speciesName := "Antilocapra americana"
	speciesID := insertTestSpecies(t, speciesName)
	basePoint := orb.Point{-110.7500, 44.5500}
	t0 := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	initialGroupID := uuid.New()

	_ = insertTestEvent(t, speciesID, basePoint, t0, &initialGroupID)

	// Call DeduplicateEvents using speciesName string (case-insensitive) instead of UUID
	dedupGroupID, isDuplicate, err := geo.DeduplicateEvents(testCtx, testPool, "antilocapra americana", basePoint, t0.Add(2*time.Hour), 100.0, 24.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !isDuplicate {
		t.Errorf("expected isDuplicate=true when matching by species name")
	}
	if dedupGroupID == nil || *dedupGroupID != initialGroupID {
		t.Errorf("expected group ID %s, got %v", initialGroupID, dedupGroupID)
	}
}

func TestDeduplicateEvents_NullDedupGroupLinksToEventID(t *testing.T) {
	speciesID := insertTestSpecies(t, "Alces alces")
	basePoint := orb.Point{-110.7600, 44.5600}
	t0 := time.Date(2026, 9, 15, 14, 0, 0, 0, time.UTC)

	// Existing event has dedup_group_id = nil
	eID := insertTestEvent(t, speciesID, basePoint, t0, nil)

	// Second report nearby within time threshold
	dedupGroupID, isDuplicate, err := geo.DeduplicateEvents(testCtx, testPool, speciesID.String(), basePoint, t0.Add(4*time.Hour), 100.0, 24.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !isDuplicate {
		t.Errorf("expected isDuplicate=true")
	}
	if dedupGroupID == nil || *dedupGroupID != eID {
		t.Errorf("expected fallback to existing event ID %s, got %v", eID, dedupGroupID)
	}
}

func TestDeduplicateEvents_ConcurrentExecution(t *testing.T) {
	speciesID := insertTestSpecies(t, "Bison bison")
	basePoint := orb.Point{-110.7700, 44.5700}
	t0 := time.Date(2026, 9, 15, 16, 0, 0, 0, time.UTC)
	initialGroupID := uuid.New()

	_ = insertTestEvent(t, speciesID, basePoint, t0, &initialGroupID)

	const concurrency = 5
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(offset int) {
			tObs := t0.Add(time.Duration(offset) * time.Hour)
			gID, isDup, err := geo.DeduplicateEvents(testCtx, testPool, speciesID.String(), basePoint, tObs, 200.0, 72.0)
			if err != nil {
				errChan <- err
				return
			}
			if !isDup || gID == nil || *gID != initialGroupID {
				errChan <- errors.New("concurrent dedup failed match")
				return
			}
			errChan <- nil
		}(i + 1)
	}

	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Fatalf("concurrent dedup error: %v", err)
		}
	}
}
