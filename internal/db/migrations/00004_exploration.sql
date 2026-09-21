-- +goose Up
CREATE TABLE wildlife_observations (
 source_id uuid NOT NULL REFERENCES sources,
 artifact_id uuid NOT NULL,
 native_id text NOT NULL,
 species text NOT NULL, road text NOT NULL, survey text NOT NULL,
 geom geometry(Point,4326) NOT NULL,
 uncertainty_m double precision NOT NULL CHECK(uncertainty_m>=0 AND uncertainty_m<'Infinity'::float8),
 period_start timestamptz NOT NULL, period_end timestamptz NOT NULL CHECK(period_end>period_start),
 time_precision text NOT NULL CHECK(time_precision IN ('day','month','year')),
 quantity bigint NOT NULL CHECK(quantity>0),
 observation_kind text GENERATED ALWAYS AS (CASE WHEN quantity=1 THEN 'event' ELSE 'aggregate' END) STORED,
 reference_text text NOT NULL,
 PRIMARY KEY(source_id,native_id),
 FOREIGN KEY(artifact_id,source_id) REFERENCES source_artifacts(id,source_id),
 CHECK(ST_X(geom) BETWEEN -180 AND 180 AND ST_Y(geom) BETWEEN -90 AND 90)
);
CREATE INDEX wildlife_observations_geom ON wildlife_observations USING gist(geom);
CREATE INDEX wildlife_observations_time ON wildlife_observations(period_start);
-- Fixed public bins prevent filter queries from exposing precise raw positions/dates.
-- Coarse source uncertainty is withheld until a matching coarser product exists.
CREATE VIEW public_observations WITH(security_barrier=true) AS
 SELECT s.id::text AS source_id,s.name AS source,s.license,s.public_url AS source_url,
 h3_lat_lng_to_cell(o.geom,6)::text AS cell,
 ST_Multi(h3_cell_to_boundary_geometry(h3_lat_lng_to_cell(o.geom,6))) AS geom,
 o.species,o.road,o.survey,
 extract(year FROM o.period_start)::int AS year,
 CASE WHEN o.time_precision='year' THEN 0 ELSE extract(month FROM o.period_start)::int END AS month,
 count(*)::bigint AS records,sum(o.quantity)::bigint AS animals,
 'h3-r6-month-v1'::text AS release_policy
 FROM wildlife_observations o JOIN sources s ON o.source_id=s.id
 WHERE s.status='approved' AND o.period_end<=now()-interval '30 days' AND o.uncertainty_m<=1000
 GROUP BY s.id,s.name,s.license,s.public_url,h3_lat_lng_to_cell(o.geom,6),o.species,o.road,o.survey,
 extract(year FROM o.period_start),CASE WHEN o.time_precision='year' THEN 0 ELSE extract(month FROM o.period_start)::int END;
REVOKE ALL ON wildlife_observations FROM PUBLIC,corridor_public;
GRANT SELECT ON public_observations TO corridor_public;
CREATE OR REPLACE VIEW public_coverage WITH(security_barrier=true) AS
 SELECT ((SELECT count(*) FROM events e JOIN sources s ON s.id=e.source_id
 LEFT JOIN species sp ON sp.id=e.species_id WHERE e.publishable AND s.status='approved'
 AND (sp.sensitive=false OR e.observed_at<=now()-interval '30 days'))+
 (SELECT coalesce(sum(records),0) FROM public_observations))::bigint AS ingested_events;
-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'exploration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
