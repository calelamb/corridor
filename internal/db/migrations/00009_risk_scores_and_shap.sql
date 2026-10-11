-- +goose Up
CREATE TABLE IF NOT EXISTS segment_risk_scores (
    segment_id UUID PRIMARY KEY REFERENCES road_segments(id) ON DELETE CASCADE,
    risk_score DOUBLE PRECISION NOT NULL CHECK (risk_score >= 0),
    percentile_rank DOUBLE PRECISION NOT NULL CHECK (percentile_rank >= 0 AND percentile_rank <= 100),
    p10 DOUBLE PRECISION NOT NULL CHECK (p10 >= 0),
    p90 DOUBLE PRECISION NOT NULL CHECK (p90 >= 0),
    monthly_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    species_breakdown JSONB NOT NULL DEFAULT '{}'::jsonb,
    shap_drivers JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS segment_risk_scores_risk_score ON segment_risk_scores(risk_score);
CREATE INDEX IF NOT EXISTS segment_risk_scores_percentile_rank ON segment_risk_scores(percentile_rank);

REVOKE ALL ON segment_risk_scores FROM PUBLIC, corridor_public;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
