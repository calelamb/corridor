package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonas-p/go-shp"
)

func MigrationArtifact() Artifact {
	return Artifact{SourceID: "0e6d797c-a39f-4692-b17a-d7d258882905", Name: "USGS · Pequop mule-deer migration areas", URL: "https://doi.org/10.5066/P9O2YM6I", License: "CC0 · USGS / Nevada Department of Wildlife · modified: generalized H3 areas", SHA256: "07e52bd6218e59f4fccae43af48d2c24e052995a0827327a5fc8c0ed9aaa18bd", ObjectKey: "sha256/07e52bd6218e59f4fccae43af48d2c24e052995a0827327a5fc8c0ed9aaa18bd/pequop.shp", RetrievedAt: time.Date(2026, 9, 21, 21, 38, 0, 0, time.UTC)}
}
func MigrationGeometry(shape shp.Shape) (json.RawMessage, error) {
	line, ok := shape.(*shp.PolyLine)
	if !ok || len(line.Points) < 2 || len(line.Parts) < 1 {
		return nil, errors.New("migration source requires polylines")
	}
	coords := make([][][2]float64, 0, len(line.Parts))
	for i, start := range line.Parts {
		end := len(line.Points)
		if i+1 < len(line.Parts) {
			end = int(line.Parts[i+1])
		}
		if start < 0 || int(start) >= end || end > len(line.Points) || end-int(start) < 2 {
			return nil, errors.New("invalid polyline parts")
		}
		points := make([][2]float64, 0, end-int(start))
		for _, p := range line.Points[start:end] {
			if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) || math.Abs(p.X) > 10000000 || math.Abs(p.Y) > 10000000 {
				return nil, errors.New("invalid projected coordinate")
			}
			points = append(points, [2]float64{p.X, p.Y})
		}
		coords = append(coords, points)
	}
	return json.Marshal(struct {
		Type        string         `json:"type"`
		Coordinates [][][2]float64 `json:"coordinates"`
	}{"MultiLineString", coords})
}
func ImportMigration(ctx context.Context, pool *pgxpool.Pool, path string) (int, error) {
	reader, err := shp.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open migration shapes: %w", err)
	}
	defer reader.Close()
	a := MigrationArtifact()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827366)"); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO sources(id,name,url,license,status,public_url) VALUES($1,$2,$3,$4,'approved',$3) ON CONFLICT(id) DO NOTHING`, a.SourceID, a.Name, a.URL, a.License); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO source_artifacts(source_id,object_key,retrieved_at,sha256) VALUES($1,$2,$3,$4) ON CONFLICT(source_id,sha256) DO NOTHING`, a.SourceID, a.ObjectKey, a.RetrievedAt, a.SHA256); err != nil {
		return 0, err
	}
	count := 0
	for reader.Next() {
		index, shape := reader.Shape()
		geometry, err := MigrationGeometry(shape)
		if err != nil {
			return 0, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO migration_routes(source_id,native_id,geom) VALUES($1,$2,ST_Transform(ST_SetSRID(ST_GeomFromGeoJSON($3),5070),4326)) ON CONFLICT DO NOTHING`, a.SourceID, strconv.Itoa(index), string(geometry)); err != nil {
			return 0, fmt.Errorf("store migration route: %w", err)
		}
		count++
		if count > 10000 {
			return 0, errors.New("migration shape limit exceeded")
		}
	}
	if err = reader.Err(); err != nil {
		return 0, fmt.Errorf("read migration source: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO migration_cells(source_id,cell) SELECT DISTINCT source_id,h3_lat_lng_to_cell((ST_DumpPoints(ST_Segmentize(geom::geography,500)::geometry)).geom,6)::text FROM migration_routes WHERE source_id=$1 ON CONFLICT DO NOTHING`, a.SourceID)
	if err != nil {
		return 0, fmt.Errorf("generalize migration: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return count, nil
}
