-- +goose Up
-- Grant application and public query roles necessary access on operational tables.
GRANT SELECT ON road_segments, segment_risk_scores TO corridor_public, corridor_app;
GRANT SELECT, INSERT, UPDATE ON field_reports TO corridor_app;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
