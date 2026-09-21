//go:build integration

package db

import (
	"github.com/jackc/pgx/v5"
	"strings"
	"testing"
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
	defer conn.Close(ctx)
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
