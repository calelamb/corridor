-- +goose Up
ALTER TABLE road_segments
  ALTER COLUMN source_id DROP NOT NULL,
  ALTER COLUMN native_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS state VARCHAR(2),
  ADD COLUMN IF NOT EXISTS highway_ref VARCHAR(32),
  ADD COLUMN IF NOT EXISTS milepost_start NUMERIC,
  ADD COLUMN IF NOT EXISTS milepost_end NUMERIC,
  ADD COLUMN IF NOT EXISTS h3_r8 h3index,
  ADD COLUMN IF NOT EXISTS osm_way_ids BIGINT[];

CREATE INDEX IF NOT EXISTS road_segments_geom ON road_segments USING gist(geom);
CREATE INDEX IF NOT EXISTS road_segments_h3_r8 ON road_segments USING btree(h3_r8);
CREATE INDEX IF NOT EXISTS road_segments_highway_ref ON road_segments(highway_ref);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION road_segments_set_h3()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.h3_r8 IS NULL AND NEW.geom IS NOT NULL THEN
    NEW.h3_r8 := h3_lat_lng_to_cell(ST_Centroid(NEW.geom), 8);
  END IF;
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS road_segments_h3_trigger ON road_segments;
CREATE TRIGGER road_segments_h3_trigger
  BEFORE INSERT OR UPDATE OF geom ON road_segments
  FOR EACH ROW
  EXECUTE FUNCTION road_segments_set_h3();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION discretize_highway(
    highway_geom GEOMETRY,
    segment_length_m DOUBLE PRECISION DEFAULT 160.9344
)
RETURNS TABLE (
    geom GEOMETRY,
    length_m DOUBLE PRECISION,
    h3_r8 H3INDEX
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    clean_geom GEOMETRY;
    total_len DOUBLE PRECISION;
    num_segs INTEGER;
    i INTEGER;
    start_frac DOUBLE PRECISION;
    end_frac DOUBLE PRECISION;
    sub_geom GEOMETRY;
    sub_len DOUBLE PRECISION;
    sub_h3 H3INDEX;
BEGIN
    IF highway_geom IS NULL THEN
        RETURN;
    END IF;

    IF segment_length_m IS NULL OR segment_length_m <= 0 THEN
        RAISE EXCEPTION 'segment_length_m must be positive';
    END IF;

    clean_geom := ST_LineMerge(ST_Force2D(highway_geom));
    clean_geom := ST_Segmentize(clean_geom::geography, segment_length_m)::geometry;
    total_len := ST_Length(clean_geom::geography);

    IF total_len <= 0 THEN
        RETURN;
    END IF;

    num_segs := GREATEST(1, CEIL(total_len / segment_length_m)::integer);
    IF num_segs > 1 AND (total_len - (num_segs - 1) * segment_length_m) < 0.001 THEN
        num_segs := num_segs - 1;
    END IF;

    FOR i IN 1..num_segs LOOP
        start_frac := ((i - 1)::double precision * segment_length_m) / total_len;
        end_frac := LEAST(1.0::double precision, (i::double precision * segment_length_m) / total_len);

        IF start_frac >= 1.0 THEN
            EXIT;
        END IF;
        IF end_frac > 1.0 OR i = num_segs THEN
            end_frac := 1.0;
        END IF;

        sub_geom := ST_LineSubstring(clean_geom, start_frac, end_frac);
        sub_len := ST_Length(sub_geom::geography);
        sub_h3 := h3_lat_lng_to_cell(ST_Centroid(sub_geom), 8);

        geom := sub_geom;
        length_m := sub_len;
        h3_r8 := sub_h3;
        RETURN NEXT;
    END LOOP;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON road_segments FROM PUBLIC, corridor_public;
REVOKE EXECUTE ON FUNCTION discretize_highway(geometry, double precision) FROM PUBLIC, corridor_public;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'migration rollback requires an explicit recovery procedure'; END $$;
-- +goose StatementEnd
