package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedSegments populates benchmark 0.1-mile road segments and risk scores
// for pilot verification and frontend visualization.
func SeedSegments(ctx context.Context, pool *pgxpool.Pool, stdout io.Writer) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if pool == nil {
		return errors.New("database pool is required")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const seedSQL = `
	-- Benchmark road segments
	INSERT INTO road_segments (
		id, state, highway_ref, milepost_start, milepost_end, length_m, native_id, geom
	) VALUES 
	(
		'00000000-0000-0000-0000-000000000001',
		'WY',
		'US-191',
		14.2,
		14.3,
		160.9344,
		'benchmark-us191-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-114.66 43.0, -114.64 43.0)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000002',
		'NV',
		'I-80',
		352.1,
		352.2,
		160.9344,
		'benchmark-i80-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-114.70 41.1, -114.68 41.1)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000003',
		'UT',
		'US-89',
		45.0,
		45.1,
		160.9344,
		'benchmark-us89-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.50 41.8, -111.48 41.8)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000004',
		'ID',
		'I-84',
		115.0,
		115.1,
		160.9344,
		'benchmark-i84-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-115.10 43.1, -115.08 43.1)'), 4326)
	)
	ON CONFLICT (id) DO UPDATE SET
		geom = EXCLUDED.geom,
		state = EXCLUDED.state,
		highway_ref = EXCLUDED.highway_ref,
		milepost_start = EXCLUDED.milepost_start,
		milepost_end = EXCLUDED.milepost_end,
		length_m = EXCLUDED.length_m;

	-- Benchmark segment risk scores
	INSERT INTO segment_risk_scores (
		segment_id, risk_score, percentile_rank, p10, p90, monthly_profile, species_breakdown, shap_drivers
	) VALUES
	(
		'00000000-0000-0000-0000-000000000001',
		14.2,
		94.0,
		10.1,
		18.5,
		'{"1": 0.8, "2": 0.9, "3": 1.1, "4": 1.4, "5": 1.8, "6": 2.1, "7": 2.0, "8": 1.7, "9": 1.3, "10": 1.0, "11": 0.9, "12": 0.8}'::jsonb,
		'{"Mule deer": 75.0, "Elk": 25.0}'::jsonb,
		'[{"feature": "Canopy cover", "impact": "High attraction", "weight": 0.45}, {"feature": "Distance to water", "impact": "Medium attraction", "weight": 0.25}, {"feature": "Traffic volume", "impact": "Avoidance", "weight": -0.30}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000002',
		18.5,
		96.5,
		12.0,
		24.0,
		'{"1": 1.2, "2": 1.5, "11": 2.0, "12": 1.8}'::jsonb,
		'{"Mule deer": 90.0, "Pronghorn": 10.0}'::jsonb,
		'[{"feature": "Migration corridor overlap", "impact": "Critical", "weight": 0.65}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000003',
		12.1,
		91.0,
		8.5,
		15.2,
		'{"5": 1.8, "6": 2.2, "7": 2.0}'::jsonb,
		'{"Pronghorn": 70.0, "Mule deer": 30.0}'::jsonb,
		'[{"feature": "Open terrain bottleneck", "impact": "High", "weight": 0.50}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000004',
		9.8,
		87.0,
		6.0,
		13.0,
		'{"10": 1.5, "11": 1.9, "12": 1.4}'::jsonb,
		'{"Mule deer": 60.0, "Elk": 40.0}'::jsonb,
		'[{"feature": "Canyon slope", "impact": "Medium", "weight": 0.35}]'::jsonb
	)
	ON CONFLICT (segment_id) DO UPDATE SET
		risk_score = EXCLUDED.risk_score,
		percentile_rank = EXCLUDED.percentile_rank,
		p10 = EXCLUDED.p10,
		p90 = EXCLUDED.p90,
		monthly_profile = EXCLUDED.monthly_profile,
		species_breakdown = EXCLUDED.species_breakdown,
		shap_drivers = EXCLUDED.shap_drivers,
		updated_at = now();
	`

	if _, err := tx.Exec(ctx, seedSQL); err != nil {
		return fmt.Errorf("execute seed sql: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, "Seeded 4 benchmark road segments with risk scores and SHAP drivers.")
	}
	return nil
}

func seedSegmentsCmd(ctx context.Context, stdout io.Writer) error {
	dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("CORRIDOR_DATABASE_URL")
	}
	if dsn == "" {
		return errors.New("CORRIDOR_MIGRATION_DATABASE_URL or CORRIDOR_DATABASE_URL required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	return SeedSegments(ctx, pool, stdout)
}
