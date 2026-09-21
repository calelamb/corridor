-- +goose Up
-- Acquisition URLs may contain credentials. Publication requires a separately
-- reviewed landing page; never infer or backfill it from the acquisition URL.
ALTER TABLE sources ADD COLUMN public_url text
 CHECK(public_url ~ '^https?://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(:[0-9]{1,5})?(/[A-Za-z0-9._~!$&()*+,;=:@%/-]*)?$');
COMMENT ON COLUMN sources.url IS 'Private acquisition URL; may contain credentials. Never publish.';
COMMENT ON COLUMN sources.public_url IS 'Reviewed public landing page. No userinfo, query, fragment, whitespace, or backslash. Check path for secrets before approval.';
CREATE OR REPLACE VIEW public_sources WITH (security_barrier=true) AS
 SELECT s.id::text AS id,s.name,coalesce(s.public_url,'') AS url,s.license,s.status FROM sources s
 WHERE s.status='approved' AND s.public_url IS NOT NULL
 AND EXISTS(SELECT 1 FROM source_artifacts a WHERE a.source_id=s.id);

-- +goose Down
-- Restoring the old projection would expose private acquisition URLs.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'public URL isolation requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
