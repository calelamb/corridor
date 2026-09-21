-- +goose Up
ALTER TABLE source_artifacts ADD CONSTRAINT source_artifact_identity UNIQUE(id,source_id);
ALTER TABLE events ADD CONSTRAINT event_artifact_source
 FOREIGN KEY(artifact_id,source_id) REFERENCES source_artifacts(id,source_id);
ALTER TABLE aggregate_observations ADD CONSTRAINT aggregate_artifact_source
 FOREIGN KEY(artifact_id,source_id) REFERENCES source_artifacts(id,source_id);
-- +goose Down
ALTER TABLE aggregate_observations DROP CONSTRAINT aggregate_artifact_source;
ALTER TABLE events DROP CONSTRAINT event_artifact_source;
ALTER TABLE source_artifacts DROP CONSTRAINT source_artifact_identity;
