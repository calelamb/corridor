-- +goose Up
CREATE TABLE IF NOT EXISTS field_reports (
    id UUID PRIMARY KEY,
    species_common_name TEXT NOT NULL,
    geom GEOMETRY(Point, 4326) NOT NULL,
    snapped_geom GEOMETRY(Point, 4326),
    segment_id UUID REFERENCES road_segments(id),
    snap_distance_m DOUBLE PRECISION,
    dedup_group_id UUID,
    is_duplicate BOOLEAN NOT NULL DEFAULT false,
    observed_at TIMESTAMPTZ NOT NULL,
    confidence TEXT NOT NULL DEFAULT 'unconfirmed',
    notes TEXT,
    photo_key TEXT,
    species_suggestions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS field_reports_geom ON field_reports USING gist(geom);
CREATE INDEX IF NOT EXISTS field_reports_snapped_geom ON field_reports USING gist(snapped_geom);
CREATE INDEX IF NOT EXISTS field_reports_observed_at ON field_reports(observed_at);
CREATE INDEX IF NOT EXISTS field_reports_dedup_group_id ON field_reports(dedup_group_id);
CREATE INDEX IF NOT EXISTS field_reports_segment_id ON field_reports(segment_id);

REVOKE ALL ON field_reports FROM PUBLIC, corridor_public;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
