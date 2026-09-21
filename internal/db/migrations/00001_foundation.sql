-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_raster;
CREATE EXTENSION IF NOT EXISTS h3;
CREATE EXTENSION IF NOT EXISTS h3_postgis;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE TABLE sources (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL CHECK(length(name)>0),
 url text NOT NULL CHECK(url ~ '^https?://'), license text NOT NULL CHECK(length(license)>0),
 status text NOT NULL CHECK(status IN ('quarantined','approved','withdrawn')),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE source_artifacts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), source_id uuid NOT NULL REFERENCES sources,
 object_key text NOT NULL UNIQUE CHECK(length(object_key)>0), retrieved_at timestamptz NOT NULL,
 sha256 text NOT NULL CHECK(sha256 ~ '^[a-f0-9]{64}$'), UNIQUE(source_id,sha256)
);
-- +goose StatementBegin
CREATE FUNCTION reject_artifact_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'source artifacts are immutable'; END; $$;
-- +goose StatementEnd
CREATE TRIGGER immutable_artifacts BEFORE UPDATE OR DELETE ON source_artifacts
 FOR EACH ROW EXECUTE FUNCTION reject_artifact_change();
CREATE TABLE species (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL UNIQUE,
 sensitive boolean NOT NULL DEFAULT true
);
CREATE TABLE road_segments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), source_id uuid NOT NULL REFERENCES sources,
 native_id text NOT NULL, geom geometry(LineString,4326) NOT NULL,
 length_m double precision NOT NULL CHECK(length_m>0 AND length_m<'Infinity'::float8),
 UNIQUE(source_id,native_id)
);
CREATE INDEX road_segments_geom ON road_segments USING gist(geom);
CREATE TABLE events (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), source_id uuid NOT NULL REFERENCES sources,
 artifact_id uuid REFERENCES source_artifacts, native_id text NOT NULL,
 species_id uuid REFERENCES species, geom geometry(Point,4326) NOT NULL,
 observed_at timestamptz NOT NULL, time_precision text NOT NULL CHECK(time_precision IN ('instant','day','month','year')),
 uncertainty_m double precision NOT NULL CHECK(uncertainty_m>=0 AND uncertainty_m<'Infinity'::float8),
 publishable boolean NOT NULL DEFAULT false, raw jsonb NOT NULL DEFAULT '{}',
 UNIQUE(source_id,native_id), CHECK(ST_X(geom) BETWEEN -180 AND 180 AND ST_Y(geom) BETWEEN -90 AND 90)
);
CREATE INDEX events_geom ON events USING gist(geom);
CREATE INDEX events_observed ON events(observed_at);
CREATE TABLE aggregate_observations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), source_id uuid NOT NULL REFERENCES sources,
 artifact_id uuid REFERENCES source_artifacts, native_id text NOT NULL,
 geom geometry(Geometry,4326) NOT NULL, period_start timestamptz NOT NULL,
 period_end timestamptz NOT NULL CHECK(period_end>period_start),
 event_count bigint NOT NULL CHECK(event_count>=0), raw jsonb NOT NULL DEFAULT '{}',
 UNIQUE(source_id,native_id)
);
CREATE INDEX aggregates_geom ON aggregate_observations USING gist(geom);
CREATE VIEW public_coverage WITH (security_barrier=true) AS
 SELECT count(*)::bigint AS ingested_events FROM events e JOIN sources s ON s.id=e.source_id
 LEFT JOIN species sp ON sp.id=e.species_id
 WHERE e.publishable AND s.status='approved'
 AND (sp.sensitive=false OR e.observed_at<=now()-interval '30 days');
CREATE VIEW public_sources WITH (security_barrier=true) AS
 SELECT s.id::text AS id,s.name,s.url,s.license,s.status FROM sources s
 WHERE s.status='approved' AND EXISTS(SELECT 1 FROM source_artifacts a WHERE a.source_id=s.id);
-- +goose StatementBegin
DO $$ BEGIN
 IF NOT EXISTS(SELECT FROM pg_roles WHERE rolname='corridor_public') THEN
  CREATE ROLE corridor_public NOLOGIN;
 END IF;
END $$;
-- +goose StatementEnd
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO corridor_public;
GRANT SELECT ON public_coverage,public_sources TO corridor_public;

-- +goose Down
-- Destructive rollback is intentionally unsupported; restore a verified backup.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'foundation rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
