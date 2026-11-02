-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  IF NOT EXISTS(SELECT FROM pg_roles WHERE rolname='corridor_app') THEN
    CREATE ROLE corridor_app NOLOGIN;
  END IF;
  IF NOT EXISTS(SELECT FROM pg_roles WHERE rolname='corridor_public') THEN
    CREATE ROLE corridor_public NOLOGIN;
  END IF;
END $$;
-- +goose StatementEnd

-- Grant application and public query roles necessary access on operational tables.
GRANT SELECT ON road_segments, segment_risk_scores TO corridor_public, corridor_app;
GRANT SELECT, INSERT, UPDATE ON field_reports TO corridor_app;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
