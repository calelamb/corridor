package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SpeciesSuggestion represents a candidate species prediction with confidence.
type SpeciesSuggestion struct {
	Species    string  `json:"species"`
	Confidence float64 `json:"confidence"`
}

// FieldReportRecord represents an ingested carcass report in the field_reports table.
type FieldReportRecord struct {
	ID                 uuid.UUID
	SpeciesCommonName  string
	Latitude           float64
	Longitude          float64
	SnappedLatitude    *float64
	SnappedLongitude   *float64
	SegmentID          *uuid.UUID
	SnapDistanceM      *float64
	DedupGroupID       *uuid.UUID
	IsDuplicate        bool
	ObservedAt         time.Time
	Confidence         string
	Notes              *string
	PhotoKey           *string
	SpeciesSuggestions []SpeciesSuggestion
	CreatedAt          time.Time
}

// CreateReport records a new field carcass report idempotently.
// If a report with the same ID already exists, it returns the existing report with created=false.
func (s Store) CreateReport(ctx context.Context, report FieldReportRecord) (*FieldReportRecord, bool, error) {
	if s.Pool == nil {
		return nil, false, errors.New("database pool unavailable")
	}

	existing, err := s.GetReport(ctx, report.ID)
	if err == nil && existing != nil {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, false, fmt.Errorf("check existing report: %w", err)
	}

	suggsJSON, err := json.Marshal(report.SpeciesSuggestions)
	if err != nil {
		suggsJSON = []byte("[]")
	}

	const insertQuery = `
		INSERT INTO field_reports (
			id,
			species_common_name,
			geom,
			snapped_geom,
			segment_id,
			snap_distance_m,
			dedup_group_id,
			is_duplicate,
			observed_at,
			confidence,
			notes,
			photo_key,
			species_suggestions,
			created_at
		) VALUES (
			$1,
			$2,
			ST_SetSRID(ST_MakePoint($3, $4), 4326),
			CASE WHEN $5::double precision IS NOT NULL AND $6::double precision IS NOT NULL
			     THEN ST_SetSRID(ST_MakePoint($6, $5), 4326)
			     ELSE NULL END,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			now()
		)
		RETURNING created_at
	`

	var createdAt time.Time
	err = s.Pool.QueryRow(ctx, insertQuery,
		report.ID,
		report.SpeciesCommonName,
		report.Longitude,
		report.Latitude,
		report.SnappedLatitude,
		report.SnappedLongitude,
		report.SegmentID,
		report.SnapDistanceM,
		report.DedupGroupID,
		report.IsDuplicate,
		report.ObservedAt,
		report.Confidence,
		report.Notes,
		report.PhotoKey,
		suggsJSON,
	).Scan(&createdAt)
	if err != nil {
		return nil, false, fmt.Errorf("insert field report: %w", err)
	}

	res := report
	res.CreatedAt = createdAt
	return &res, true, nil
}

// GetReport looks up a field report by its primary UUID.
func (s Store) GetReport(ctx context.Context, id uuid.UUID) (*FieldReportRecord, error) {
	if s.Pool == nil {
		return nil, errors.New("database pool unavailable")
	}

	const selectQuery = `
		SELECT
			id,
			species_common_name,
			ST_Y(geom),
			ST_X(geom),
			ST_Y(snapped_geom),
			ST_X(snapped_geom),
			segment_id,
			snap_distance_m,
			dedup_group_id,
			is_duplicate,
			observed_at,
			confidence,
			notes,
			photo_key,
			species_suggestions,
			created_at
		FROM field_reports
		WHERE id = $1
	`

	var r FieldReportRecord
	var suggsJSON []byte
	err := s.Pool.QueryRow(ctx, selectQuery, id).Scan(
		&r.ID,
		&r.SpeciesCommonName,
		&r.Latitude,
		&r.Longitude,
		&r.SnappedLatitude,
		&r.SnappedLongitude,
		&r.SegmentID,
		&r.SnapDistanceM,
		&r.DedupGroupID,
		&r.IsDuplicate,
		&r.ObservedAt,
		&r.Confidence,
		&r.Notes,
		&r.PhotoKey,
		&suggsJSON,
		&r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query field report: %w", err)
	}

	if len(suggsJSON) > 0 {
		_ = json.Unmarshal(suggsJSON, &r.SpeciesSuggestions)
	}
	return &r, nil
}

// UpdateReportPhoto updates the photo key and species suggestion chips for a field report.
func (s Store) UpdateReportPhoto(ctx context.Context, id uuid.UUID, photoKey string, suggestions []SpeciesSuggestion) (*FieldReportRecord, error) {
	if s.Pool == nil {
		return nil, errors.New("database pool unavailable")
	}

	suggsJSON, err := json.Marshal(suggestions)
	if err != nil {
		suggsJSON = []byte("[]")
	}

	const updateQuery = `
		UPDATE field_reports
		SET photo_key = $2,
		    species_suggestions = $3
		WHERE id = $1
		RETURNING
			id,
			species_common_name,
			ST_Y(geom),
			ST_X(geom),
			ST_Y(snapped_geom),
			ST_X(snapped_geom),
			segment_id,
			snap_distance_m,
			dedup_group_id,
			is_duplicate,
			observed_at,
			confidence,
			notes,
			photo_key,
			species_suggestions,
			created_at
	`

	var r FieldReportRecord
	var outSuggsJSON []byte
	err = s.Pool.QueryRow(ctx, updateQuery, id, photoKey, suggsJSON).Scan(
		&r.ID,
		&r.SpeciesCommonName,
		&r.Latitude,
		&r.Longitude,
		&r.SnappedLatitude,
		&r.SnappedLongitude,
		&r.SegmentID,
		&r.SnapDistanceM,
		&r.DedupGroupID,
		&r.IsDuplicate,
		&r.ObservedAt,
		&r.Confidence,
		&r.Notes,
		&r.PhotoKey,
		&outSuggsJSON,
		&r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update field report photo: %w", err)
	}

	if len(outSuggsJSON) > 0 {
		_ = json.Unmarshal(outSuggsJSON, &r.SpeciesSuggestions)
	}
	return &r, nil
}
