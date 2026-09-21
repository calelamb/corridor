//go:build integration

package db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func testDatabase(t *testing.T, image string) (context.Context, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	c, err := postgres.Run(ctx, image, postgres.WithDatabase("corridor_test"), postgres.WithUsername("test_admin"), postgres.WithPassword("synthetic-test-only"), postgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(context.Background()); err != nil {
			t.Error(err)
		}
	})
	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return ctx, dsn
}
func TestSpatialFoundation(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var extensions int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM pg_extension WHERE extname IN ('postgis','h3','h3_postgis')").Scan(&extensions); err != nil || extensions != 3 {
		t.Fatalf("extensions %d: %v", extensions, err)
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "SET ROLE corridor_public"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"events", "source_artifacts", "aggregate_observations", "species", "road_segments", "sources"} {
		if _, err := conn.Exec(ctx, "SELECT * FROM "+table); err == nil {
			t.Fatalf("public can read %s", table)
		}
	}
	var count int64
	if err := conn.QueryRow(ctx, "SELECT ingested_events FROM public_coverage").Scan(&count); err != nil || count != 0 {
		t.Fatalf("public count %d: %v", count, err)
	}
	store := Store{Pool: pool}
	if count, err := store.EventCount(ctx); err != nil || count != 0 {
		t.Fatalf("empty %d: %v", count, err)
	}
	sources, total, err := store.Sources(ctx, 25, 0)
	if err != nil || total != 0 || len(sources) != 0 {
		t.Fatalf("sources %v %d: %v", sources, total, err)
	}
	testConstraints(t, ctx, pool)
}
func TestMissingH3IsAtomic(t *testing.T) {
	ctx, dsn := testDatabase(t, "postgres:16-bookworm@sha256:efedf3595f1d6f415c08568ba171029bf54052e754cc9f030e3f2412b21f3d67")
	if err := Migrate(ctx, dsn); err == nil {
		t.Fatal("missing extensions accepted")
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT to_regclass('public.sources') IS NOT NULL").Scan(&exists); err != nil || exists {
		t.Fatalf("partial schema: %v %v", exists, err)
	}
}
func TestOpenInvalid(t *testing.T) {
	if pool, err := Open(context.Background(), "invalid"); err == nil {
		pool.Close()
		t.Fatal("invalid DSN accepted")
	} else if strings.Contains(err.Error(), "password") {
		t.Fatal("secret in error")
	}
}
