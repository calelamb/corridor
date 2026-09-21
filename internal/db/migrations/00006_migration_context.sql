-- +goose Up
CREATE TABLE migration_routes(source_id uuid REFERENCES sources NOT NULL,native_id text NOT NULL,geom geometry(MultiLineString,4326) NOT NULL,PRIMARY KEY(source_id,native_id));
CREATE TABLE migration_cells(source_id uuid REFERENCES sources NOT NULL,cell text NOT NULL,PRIMARY KEY(source_id,cell));
CREATE VIEW public_migration WITH(security_barrier=true) AS
 SELECT c.cell,ST_Multi(h3_cell_to_boundary_geometry(c.cell::h3index)) AS geom,s.name,s.license,s.public_url AS source_url,
 'Mapped migration areas · not live tracks'::text AS meaning,'2011–2017'::text AS period
 FROM migration_cells c JOIN sources s ON s.id=c.source_id WHERE s.status='approved';
REVOKE ALL ON migration_routes,migration_cells FROM PUBLIC,corridor_public;
GRANT SELECT ON public_migration TO corridor_public;
-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
