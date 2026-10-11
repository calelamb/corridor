package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"corridor/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type SegmentFilter struct {
	BBox    *[4]float64
	MinRisk *float64
}

type ShapDriver struct {
	Feature string  `json:"feature"`
	Impact  string  `json:"impact"`
	Weight  float64 `json:"weight"`
}

type SegmentSummary struct {
	ID             uuid.UUID
	Coordinates    [][]float64
	RiskScore      float64
	PercentileRank float64
	P10            float64
	P90            float64
	HighwayRef     *string
	State          *string
}

type SegmentDetail struct {
	ID               uuid.UUID
	Coordinates      [][]float64
	RiskScore        float64
	PercentileRank   float64
	P10              float64
	P90              float64
	HighwayRef       *string
	State            *string
	MonthlyProfile   map[string]float64
	SpeciesBreakdown map[string]float64
	ShapDrivers      []ShapDriver
}

type SourceSummary = generated.PublicSource

type Store struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("database configuration invalid")
	}
	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("database pool unavailable")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("database connection unavailable")
	}
	return pool, nil
}
func (s Store) EventCount(ctx context.Context) (int64, error) {
	count, err := generated.New(s.Pool).EventCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("read public coverage: %w", err)
	}
	return count, nil
}
func (s Store) Sources(ctx context.Context, limit, offset int32) ([]SourceSummary, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return nil, 0, errors.New("pagination out of range")
	}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, fmt.Errorf("begin sources snapshot: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	q := generated.New(tx)
	total, err := q.SourceCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count public sources: %w", err)
	}
	rows, err := q.Sources(ctx, generated.SourcesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("read public sources: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("finish sources snapshot: %w", err)
	}
	return rows, total, nil
}
func (s Store) Check(ctx context.Context) error {
	var healthy bool
	err := s.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*)=3 FROM pg_extension WHERE extname IN ('postgis','h3','h3_postgis'))
 AND to_regclass('public.public_coverage') IS NOT NULL
 AND to_regclass('public.public_sources') IS NOT NULL`).Scan(&healthy)
	if err != nil || !healthy {
		return errors.New("database schema unavailable")
	}
	_, err = s.EventCount(ctx)
	return err
}

func (s Store) Segments(ctx context.Context, filter SegmentFilter) ([]SegmentSummary, error) {
	if s.Pool == nil {
		return nil, errors.New("database pool unavailable")
	}

	query := `
		SELECT 
			r.id,
			ST_AsGeoJSON(r.geom)::text,
			s.risk_score,
			s.percentile_rank,
			s.p10,
			s.p90,
			r.highway_ref,
			r.state
		FROM road_segments r
		JOIN segment_risk_scores s ON r.id = s.segment_id
		WHERE ($1::double precision IS NULL OR s.risk_score >= $1)
		  AND (
			$2::double precision IS NULL OR
			r.geom && ST_MakeEnvelope($2, $3, $4, $5, 4326)
		  )
		ORDER BY s.risk_score DESC
		LIMIT 2000
	`

	var minRisk any
	if filter.MinRisk != nil {
		minRisk = *filter.MinRisk
	}

	var minLon, minLat, maxLon, maxLat any
	if filter.BBox != nil {
		minLon = filter.BBox[0]
		minLat = filter.BBox[1]
		maxLon = filter.BBox[2]
		maxLat = filter.BBox[3]
	}

	rows, err := s.Pool.Query(ctx, query, minRisk, minLon, minLat, maxLon, maxLat)
	if err != nil {
		return nil, fmt.Errorf("query segments: %w", err)
	}
	defer rows.Close()

	var result []SegmentSummary
	for rows.Next() {
		var (
			seg       SegmentSummary
			geomJSON  string
			hwRef     *string
			stateCode *string
		)
		if err := rows.Scan(
			&seg.ID,
			&geomJSON,
			&seg.RiskScore,
			&seg.PercentileRank,
			&seg.P10,
			&seg.P90,
			&hwRef,
			&stateCode,
		); err != nil {
			return nil, fmt.Errorf("scan segment: %w", err)
		}
		seg.HighwayRef = hwRef
		seg.State = stateCode

		var geojson struct {
			Coordinates [][]float64 `json:"coordinates"`
		}
		if err := json.Unmarshal([]byte(geomJSON), &geojson); err != nil {
			return nil, fmt.Errorf("unmarshal geometry: %w", err)
		}
		seg.Coordinates = geojson.Coordinates
		result = append(result, seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate segments: %w", err)
	}
	return result, nil
}

func (s Store) Segment(ctx context.Context, id uuid.UUID) (*SegmentDetail, error) {
	if s.Pool == nil {
		return nil, errors.New("database pool unavailable")
	}

	query := `
		SELECT 
			r.id,
			ST_AsGeoJSON(r.geom)::text,
			s.risk_score,
			s.percentile_rank,
			s.p10,
			s.p90,
			r.highway_ref,
			r.state,
			s.monthly_profile,
			s.species_breakdown,
			s.shap_drivers
		FROM road_segments r
		JOIN segment_risk_scores s ON r.id = s.segment_id
		WHERE r.id = $1
	`

	var (
		seg              SegmentDetail
		geomJSON         string
		hwRef            *string
		stateCode        *string
		monthlyProfile   []byte
		speciesBreakdown []byte
		shapDrivers      []byte
	)

	err := s.Pool.QueryRow(ctx, query, id).Scan(
		&seg.ID,
		&geomJSON,
		&seg.RiskScore,
		&seg.PercentileRank,
		&seg.P10,
		&seg.P90,
		&hwRef,
		&stateCode,
		&monthlyProfile,
		&speciesBreakdown,
		&shapDrivers,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query segment detail: %w", err)
	}

	seg.HighwayRef = hwRef
	seg.State = stateCode

	var geojson struct {
		Coordinates [][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(geomJSON), &geojson); err != nil {
		return nil, fmt.Errorf("unmarshal geometry: %w", err)
	}
	seg.Coordinates = geojson.Coordinates

	if len(monthlyProfile) > 0 {
		if err := json.Unmarshal(monthlyProfile, &seg.MonthlyProfile); err != nil {
			return nil, fmt.Errorf("unmarshal monthly profile: %w", err)
		}
	}
	if len(speciesBreakdown) > 0 {
		if err := json.Unmarshal(speciesBreakdown, &seg.SpeciesBreakdown); err != nil {
			return nil, fmt.Errorf("unmarshal species breakdown: %w", err)
		}
	}
	if len(shapDrivers) > 0 {
		if err := json.Unmarshal(shapDrivers, &seg.ShapDrivers); err != nil {
			return nil, fmt.Errorf("unmarshal shap drivers: %w", err)
		}
	}

	return &seg, nil
}

