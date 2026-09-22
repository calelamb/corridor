package explore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Read(context.Context, Filters) (json.RawMessage, error)
	Tile(context.Context, Filters, int, int, int) ([]byte, error)
}
type Store struct{ Pool *pgxpool.Pool }

func (s Store) Read(ctx context.Context, f Filters) (json.RawMessage, error) {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin exploration: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	args := append(f.args(), f.Limit, f.Offset)
	query := `WITH filtered AS (SELECT * FROM public_observations` + predicate + `),
 cells AS (SELECT cell,ST_AsGeoJSON(geom)::json AS geometry,sum(records)::bigint AS records,sum(animals)::bigint AS animals,min(year) AS first_year,max(year) AS last_year,string_agg(DISTINCT road,', ' ORDER BY road) AS roads FROM filtered GROUP BY cell,geom),
 page AS (SELECT * FROM cells ORDER BY records DESC,cell LIMIT $13 OFFSET $14),
 timeline AS (SELECT year,sum(records)::bigint AS records,sum(animals)::bigint AS animals FROM filtered GROUP BY year ORDER BY year),
 seasons AS (SELECT month,sum(records)::bigint AS records FROM filtered GROUP BY month ORDER BY month)
 SELECT json_build_object('type','FeatureCollection','features',coalesce((SELECT json_agg(json_build_object('type','Feature','id',cell,'geometry',geometry,'properties',json_build_object('cell',cell,'records',records,'animals',animals,'first_year',first_year,'last_year',last_year,'roads',roads))) FROM page),'[]'),
 'summary',json_build_object('records',(SELECT coalesce(sum(records),0) FROM filtered),'animals',(SELECT coalesce(sum(animals),0) FROM filtered),'cells',(SELECT count(*) FROM cells),'undated_month_records',(SELECT coalesce(sum(records),0) FROM filtered WHERE month=0)),
 'timeline',coalesce((SELECT json_agg(timeline) FROM timeline),'[]'),'months',coalesce((SELECT json_agg(seasons) FROM seasons),'[]'),
 'facets',json_build_object('species',(SELECT coalesce(json_agg(x),'[]') FROM (SELECT species AS value,sum(records)::bigint AS records FROM public_observations GROUP BY species ORDER BY sum(records) DESC LIMIT 100)x),
 'roads',(SELECT coalesce(json_agg(x),'[]') FROM (SELECT road AS value,sum(records)::bigint AS records FROM public_observations WHERE road<>'' GROUP BY road ORDER BY sum(records) DESC LIMIT 100)x),
 'sources',(SELECT coalesce(json_agg(x),'[]') FROM (SELECT DISTINCT source_id AS value,source AS name,license,source_url FROM public_observations)x)),
 'policy','` + Policy + `','limit',$13::int,'offset',$14::int)`
	var result []byte
	if err = tx.QueryRow(ctx, query, args...).Scan(&result); err != nil {
		return nil, fmt.Errorf("read exploration: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("finish exploration: %w", err)
	}
	return result, nil
}
func (s Store) Tile(ctx context.Context, f Filters, z, x, y int) ([]byte, error) {
	query := `WITH filtered AS(SELECT * FROM public_observations` + predicate + `), grouped AS(SELECT cell,geom,sum(records)::bigint AS records,sum(animals)::bigint AS animals FROM filtered GROUP BY cell,geom),
 mvt AS(SELECT cell,records,animals,ST_AsMVTGeom(ST_Transform(geom,3857),ST_TileEnvelope($13,$14,$15),4096,64,true) AS geom FROM grouped WHERE ST_Intersects(geom,ST_Transform(ST_TileEnvelope($13,$14,$15),4326))) SELECT coalesce(ST_AsMVT(mvt,'evidence',4096,'geom'),'') FROM mvt`
	var result []byte
	err := s.Pool.QueryRow(ctx, query, append(f.args(), z, x, y)...).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("read evidence tile: %w", err)
	}
	return result, nil
}

func (s Store) Migration(ctx context.Context) (json.RawMessage, error) {
	var data []byte
	err := s.Pool.QueryRow(ctx, `SELECT json_build_object('type','FeatureCollection','features',coalesce(json_agg(json_build_object('type','Feature','geometry',ST_AsGeoJSON(geom)::json,'properties',json_build_object('name',name,'period',period,'meaning',meaning,'license',license,'source_url',source_url))),'[]')) FROM public_migration`).Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("read migration areas: %w", err)
	}
	return data, nil
}
