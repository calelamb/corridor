-- +goose Up
CREATE TABLE analysis_roads(source_id uuid REFERENCES sources NOT NULL,native_id text NOT NULL,ref text NOT NULL,geom geometry(LineString,4326) NOT NULL,PRIMARY KEY(source_id,native_id));
CREATE INDEX analysis_roads_geom ON analysis_roads USING gist(geom);
CREATE INDEX migration_routes_geom ON migration_routes USING gist(geom);
CREATE TABLE movement_features(cell text PRIMARY KEY,block text NOT NULL,geom geometry(MultiPolygon,4326) NOT NULL,x double precision NOT NULL,y double precision NOT NULL,routes integer NOT NULL,road boolean NOT NULL,crossings integer NOT NULL);
-- +goose StatementBegin
CREATE FUNCTION rebuild_movement_features() RETURNS void LANGUAGE plpgsql AS $$
BEGIN
 DELETE FROM movement_features;
 INSERT INTO movement_features
 WITH domain AS(SELECT DISTINCT h3_grid_disk(cell::h3index,1) AS h FROM migration_cells WHERE source_id='0e6d797c-a39f-4692-b17a-d7d258882905'),
 cells AS(SELECT h,ST_Multi(h3_cell_to_boundary_geometry(h)) AS g FROM domain)
 SELECT h::text,h3_cell_to_parent(h,5)::text,g,ST_X(ST_Transform(ST_Centroid(g),5070)),ST_Y(ST_Transform(ST_Centroid(g),5070)),
 (SELECT count(*) FROM migration_routes r WHERE r.source_id='0e6d797c-a39f-4692-b17a-d7d258882905' AND ST_Intersects(r.geom,g))::integer,
 EXISTS(SELECT 1 FROM analysis_roads a WHERE a.source_id='48d59a58-9b6b-4f1a-a1fa-ce0cdedc1f81' AND ST_Intersects(a.geom,g)),
 (SELECT count(*) FROM migration_routes r WHERE r.source_id='0e6d797c-a39f-4692-b17a-d7d258882905' AND ST_Intersects(r.geom,g) AND EXISTS(SELECT 1 FROM analysis_roads a WHERE a.source_id='48d59a58-9b6b-4f1a-a1fa-ce0cdedc1f81' AND ST_Intersects(a.geom,g) AND ST_Intersects(ST_Intersection(a.geom,g),r.geom)))::integer
 FROM cells;
END $$;
-- +goose StatementEnd
CREATE VIEW public_movement_features WITH(security_barrier=true) AS
 SELECT f.* FROM movement_features f
 WHERE EXISTS(SELECT 1 FROM sources WHERE id='0e6d797c-a39f-4692-b17a-d7d258882905' AND status='approved')
 AND EXISTS(SELECT 1 FROM sources WHERE id='48d59a58-9b6b-4f1a-a1fa-ce0cdedc1f81' AND status='approved');
REVOKE ALL ON analysis_roads,movement_features FROM PUBLIC,corridor_public;
REVOKE EXECUTE ON FUNCTION rebuild_movement_features() FROM PUBLIC,corridor_public;
GRANT SELECT ON public_movement_features TO corridor_public;
-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
