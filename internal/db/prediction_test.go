//go:build integration

package db

import (
	"context"
	"corridor/internal/ingest"
	"corridor/internal/predict"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPredictionPrivacyAndWithdrawal(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	a := ingest.MigrationArtifact()
	_, err = pool.Exec(ctx, `INSERT INTO sources(id,name,url,license,status,public_url) VALUES($1,'Synthetic migration',$2,'CC0','approved',$2);`, a.SourceID, a.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `WITH inserted AS (INSERT INTO migration_routes VALUES($1,'synthetic',ST_Multi(ST_GeomFromText('LINESTRING(-115 41,-114.8 41.2)',4326))) RETURNING *) INSERT INTO migration_cells SELECT source_id,h3_lat_lng_to_cell(ST_StartPoint(ST_GeometryN(geom,1)),6)::text FROM inserted`, a.SourceID)
	if err != nil {
		t.Fatal(err)
	}
	roads, err := ingest.ParseRoads(strings.NewReader(`{"elements":[{"type":"way","id":1,"tags":{"highway":"motorway","ref":"I 80"},"geometry":[{"lon":-115.1,"lat":41.1},{"lon":-114.7,"lat":41.1}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = ingest.ImportRoads(ctx, pool, roads, ingest.RoadArtifact()); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, e := c.Exec(ctx, "SET ROLE corridor_public")
		return e
	}
	public, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer public.Close()
	rows, err := (predict.Store{Pool: public}).Samples(ctx)
	if err != nil || len(rows) != 7 {
		t.Fatalf("features %d %v", len(rows), err)
	}
	var safe bool
	if err = public.QueryRow(ctx, `SELECT bool_and(h3_get_resolution(cell::h3index)=6 AND geom=ST_Multi(h3_cell_to_boundary_geometry(cell::h3index))) FROM public_movement_features`).Scan(&safe); err != nil || !safe {
		t.Fatal("privacy", err)
	}
	for _, table := range []string{"analysis_roads", "movement_features", "migration_routes"} {
		if _, err = public.Exec(ctx, "SELECT * FROM "+table); err == nil {
			t.Fatal("raw accessible", table)
		}
	}
	if _, err = public.Exec(ctx, "SELECT rebuild_movement_features()"); err == nil {
		t.Fatal("public mutation allowed")
	}
	for _, id := range []string{a.SourceID, ingest.RoadArtifact().SourceID} {
		if _, err = pool.Exec(ctx, "UPDATE sources SET status='withdrawn' WHERE id=$1", id); err != nil {
			t.Fatal(err)
		}
		rows, err = (predict.Store{Pool: public}).Samples(ctx)
		if err != nil || len(rows) != 0 {
			t.Fatal("withdrawal failed", err)
		}
		if _, err = pool.Exec(ctx, "UPDATE sources SET status='approved' WHERE id=$1", id); err != nil {
			t.Fatal(err)
		}
	}
	public.Close()
	if _, err = (predict.Store{Pool: public}).Samples(ctx); err == nil {
		t.Fatal("outage hidden")
	}
}
