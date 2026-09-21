-- +goose Up
-- Public coordinate reference definitions are required by PostGIS ST_Transform.
GRANT SELECT ON spatial_ref_sys TO corridor_public;
-- +goose Down
REVOKE SELECT ON spatial_ref_sys FROM corridor_public;
