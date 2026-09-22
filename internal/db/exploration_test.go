//go:build integration

package db

import (
	"context"
	"corridor/internal/explore"
	"corridor/internal/ingest"
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonas-p/go-shp"
)

func TestPilotImportAndPublicRelease(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	data, err := ingest.ParseRoadkill(strings.NewReader("occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\nsynth,US,Synthetic test species,-115,43,30,3,2014,2,2,Test road,Systematic,\n"))
	if err != nil {
		t.Fatal(err)
	}
	artifact := ingest.PilotArtifact()
	artifact.SHA256 = strings.Repeat("a", 64)
	for range 2 {
		if err := ingest.Import(ctx, pool, data, artifact); err != nil {
			t.Fatal(err)
		}
	}
	var rows, animals int
	if err := pool.QueryRow(ctx, "SELECT count(*),sum(quantity) FROM wildlife_observations").Scan(&rows, &animals); err != nil || rows != 1 || animals != 3 {
		t.Fatalf("idempotency %d %d %v", rows, animals, err)
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}()
	if _, err = conn.Exec(ctx, "SET ROLE corridor_public"); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(ctx, "SELECT * FROM wildlife_observations"); err == nil {
		t.Fatal("raw observations exposed")
	}
	var cell string
	var generalized bool
	if err = conn.QueryRow(ctx, "SELECT cell,h3_get_resolution(cell::h3index)=6 AND ST_GeometryType(geom) IN ('ST_Polygon','ST_MultiPolygon') FROM public_observations").Scan(&cell, &generalized); err != nil || !generalized {
		t.Fatalf("public geometry %v", err)
	}
	if _, err = pool.Exec(ctx, "UPDATE sources SET status='withdrawn' WHERE id=$1", artifact.SourceID); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, "SELECT count(*) FROM public_observations").Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("withdrawal %d %v", rows, err)
	}
}

func TestExplorationQueries(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	artifact := ingest.PilotArtifact()
	artifact.SHA256 = strings.Repeat("b", 64)
	data, err := ingest.ParseRoadkill(strings.NewReader("occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\nsynth,US,Synthetic species,-115,43,30,1,2014,2,2,Test road,Systematic,\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err = ingest.Import(ctx, pool, data, artifact); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET ROLE corridor_public")
		return err
	}
	publicPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer publicPool.Close()
	store := explore.Store{Pool: publicPool}
	f, err := explore.Parse(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	body, err := store.Read(ctx, f)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Summary struct {
			Records int `json:"records"`
		} `json:"summary"`
		Features []any `json:"features"`
	}
	if err = json.Unmarshal(body, &result); err != nil || result.Summary.Records != 1 || len(result.Features) != 1 {
		t.Fatalf("query response %s %v", body, err)
	}
	if tile, err := store.Tile(ctx, f, 8, 46, 94); err != nil || len(tile) == 0 {
		t.Fatalf("tile %d %v", len(tile), err)
	}
	f.Season = "summer"
	body, err = store.Read(ctx, f)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &result); err != nil || result.Summary.Records != 0 {
		t.Fatalf("season filter %s %v", body, err)
	}
}

func TestMigrationPublicGeneralization(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	path := filepath.Join(t.TempDir(), "synthetic.shp")
	writer, err := shp.Create(path, shp.POLYLINE)
	if err != nil {
		t.Fatal(err)
	}
	writer.Write(&shp.PolyLine{NumParts: 1, NumPoints: 2, Parts: []int32{0}, Points: []shp.Point{{X: -1500000, Y: 2100000}, {X: -1499000, Y: 2101000}}})
	writer.Close()
	for range 2 {
		n, err := ingest.ImportMigration(ctx, pool, path)
		if err != nil || n != 1 {
			t.Fatalf("migration import %d %v", n, err)
		}
	}
	var count int
	var coarse bool
	if err = pool.QueryRow(ctx, "SELECT count(*),bool_and(h3_get_resolution(cell::h3index)=6) FROM public_migration").Scan(&count, &coarse); err != nil || count < 1 || !coarse {
		t.Fatalf("migration release %d %v %v", count, coarse, err)
	}
	if _, err = (explore.Store{Pool: pool}).Migration(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE sources SET status='withdrawn' WHERE id=$1", ingest.MigrationArtifact().SourceID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM public_migration").Scan(&count); err != nil || count != 0 {
		t.Fatal("withdrawn migration visible", err)
	}
}
