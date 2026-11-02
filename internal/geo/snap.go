// Package geo provides PostGIS spatial operations including highway segment snapping
// and carcass observation deduplication.
package geo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paulmach/orb"
)

// ErrExceedsSnapTolerance is returned when a point is farther from all road segments
// than the specified maximum snap distance tolerance.
var ErrExceedsSnapTolerance = errors.New("exceeds snap distance tolerance")

// SnapToSegment projects a raw GPS point onto the nearest road segment in road_segments
// within maxDistanceMeters. If no segment exists within the tolerance, it returns
// ErrExceedsSnapTolerance without modifying any database records.
func SnapToSegment(
	ctx context.Context,
	pool *pgxpool.Pool,
	point orb.Point,
	maxDistanceMeters float64,
) (segmentID uuid.UUID, snapDistance float64, snappedPoint orb.Point, err error) {
	if ctx == nil {
		return uuid.Nil, 0, orb.Point{}, errors.New("context is required")
	}
	if pool == nil {
		return uuid.Nil, 0, orb.Point{}, errors.New("database pool is required")
	}
	if maxDistanceMeters <= 0 || math.IsNaN(maxDistanceMeters) || math.IsInf(maxDistanceMeters, 0) {
		return uuid.Nil, 0, orb.Point{}, errors.New("maxDistanceMeters must be positive finite number")
	}
	if math.IsNaN(point[0]) || math.IsInf(point[0], 0) || math.IsNaN(point[1]) || math.IsInf(point[1], 0) {
		return uuid.Nil, 0, orb.Point{}, errors.New("point coordinates must be finite numbers")
	}
	if point[0] < -180 || point[0] > 180 || point[1] < -90 || point[1] > 90 {
		return uuid.Nil, 0, orb.Point{}, errors.New("point coordinates out of WGS 84 bounds")
	}

	const snapQuery = `
		SELECT id,
		       ST_Distance(geom::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) AS snap_dist,
		       ST_X(ST_ClosestPoint(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326))) AS snap_lon,
		       ST_Y(ST_ClosestPoint(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326))) AS snap_lat
		FROM road_segments
		WHERE ST_DWithin(geom::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
		ORDER BY ST_Distance(geom::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) ASC
		LIMIT 1
	`

	err = pool.QueryRow(ctx, snapQuery, point[0], point[1], maxDistanceMeters).Scan(
		&segmentID,
		&snapDistance,
		&snappedPoint[0],
		&snappedPoint[1],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, 0, orb.Point{}, ErrExceedsSnapTolerance
		}
		return uuid.Nil, 0, orb.Point{}, fmt.Errorf("snap to road segment: %w", err)
	}

	return segmentID, snapDistance, snappedPoint, nil
}

// DeduplicateEvents determines if an observation corresponds to a known carcass report
// of the same species within distanceThresholdM and timeThresholdHours.
// If an existing duplicate is found, it links to its dedup_group_id and returns isDuplicate=true.
// If no duplicate is found, it generates a new group UUID and returns isDuplicate=false.
// Uses a parameterized transaction advisory lock to serialize deduplication checks per species.
func DeduplicateEvents(
	ctx context.Context,
	pool *pgxpool.Pool,
	speciesID string,
	geom orb.Point,
	observedAt time.Time,
	distanceThresholdM float64,
	timeThresholdHours float64,
) (dedupGroupID *uuid.UUID, isDuplicate bool, err error) {
	if ctx == nil {
		return nil, false, errors.New("context is required")
	}
	if pool == nil {
		return nil, false, errors.New("database pool is required")
	}
	if speciesID == "" {
		return nil, false, errors.New("speciesID is required")
	}
	if distanceThresholdM <= 0 {
		return nil, false, errors.New("distanceThresholdM must be positive")
	}
	if timeThresholdHours <= 0 {
		return nil, false, errors.New("timeThresholdHours must be positive")
	}
	if observedAt.IsZero() {
		return nil, false, errors.New("observedAt must be non-zero")
	}
	if math.IsNaN(geom[0]) || math.IsInf(geom[0], 0) || math.IsNaN(geom[1]) || math.IsInf(geom[1], 0) {
		return nil, false, errors.New("geom coordinates must be finite numbers")
	}
	if geom[0] < -180 || geom[0] > 180 || geom[1] < -90 || geom[1] > 90 {
		return nil, false, errors.New("geom coordinates out of WGS 84 bounds")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin deduplication transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			// ignore rollback error on completed transaction
		}
	}()

	// Acquire parameterized advisory lock per species to serialize concurrent deduplication checks
	lockKey := "event_dedup:" + speciesID
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey); err != nil {
		return nil, false, fmt.Errorf("acquire dedup advisory lock: %w", err)
	}

	const dedupQuery = `
		SELECT COALESCE(e.dedup_group_id, e.id)
		FROM events e
		LEFT JOIN species s ON s.id = e.species_id
		WHERE (e.species_id::text = $1 OR lower(s.name) = lower($1) OR s.id::text = $1)
		  AND ST_DWithin(e.geom::geography, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4)
		  AND abs(EXTRACT(EPOCH FROM (e.observed_at - $5))) <= ($6 * 3600)
		ORDER BY ST_Distance(e.geom::geography, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography) ASC
		LIMIT 1
	`

	var existingGroupID uuid.UUID
	err = tx.QueryRow(ctx, dedupQuery, speciesID, geom[0], geom[1], distanceThresholdM, observedAt, timeThresholdHours).Scan(&existingGroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.Commit(ctx); err != nil {
				return nil, false, fmt.Errorf("commit deduplication transaction: %w", err)
			}
			newID := uuid.New()
			return &newID, false, nil
		}
		return nil, false, fmt.Errorf("query existing events: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit deduplication transaction: %w", err)
	}

	return &existingGroupID, true, nil
}
