//go:build integration

package db

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestProvision(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Provision(ctx, dsn, strings.Repeat("x", 32)); err != nil {
			t.Fatal(err)
		}
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
	testProvisionRejectsPublicRole(t, ctx, conn, dsn)
	if _, err := conn.Exec(ctx, "SET ROLE corridor_app"); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{"SELECT * FROM events", "CREATE TABLE public.forbidden(id int)", "SELECT * FROM source_artifacts"} {
		if _, err := conn.Exec(ctx, sql); err == nil {
			t.Fatalf("unexpected app access: %s", sql)
		}
	}
	var n int64
	if err := conn.QueryRow(ctx, "SELECT ingested_events FROM public_coverage").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if err := Provision(ctx, "", ""); err == nil {
		t.Fatal("missing credentials accepted")
	}
}

func testProvisionRejectsPublicRole(t *testing.T, ctx context.Context, conn *pgx.Conn, dsn string) {
	t.Helper()
	// A non-administrator must not rotate another role's credentials.
	if _, err := conn.Exec(ctx, "CREATE ROLE synthetic_viewer LOGIN PASSWORD 'synthetic-test-only'"); err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	limited := url.URL{Scheme: endpoint.Scheme, User: url.UserPassword("synthetic_viewer", "synthetic-test-only"), Host: endpoint.Host, Path: endpoint.Path, RawQuery: endpoint.RawQuery}
	if err := Provision(ctx, limited.String(), strings.Repeat("y", 32)); err == nil {
		t.Fatal("unprivileged provisioning accepted")
	}
	app := url.URL{Scheme: endpoint.Scheme, User: url.UserPassword("corridor_app", strings.Repeat("x", 32)), Host: endpoint.Host, Path: endpoint.Path, RawQuery: endpoint.RawQuery}
	unchanged, err := pgx.Connect(ctx, app.String())
	if err != nil {
		t.Fatal("failed provisioning changed application credentials")
	}
	if err := unchanged.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
