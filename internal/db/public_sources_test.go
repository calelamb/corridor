//go:build integration

package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAcquisitionCredentialsStayPrivate(t *testing.T) {
	ctx, dsn := testDatabase(t, "corridor-postgres:foundation")
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, acquisition := range []string{
		"https://test-user:test-password@example.invalid/download",
		"https://example.invalid/download?X-Amz-Credential=test-secret&X-Amz-Signature=test-signature",
	} {
		t.Run(acquisition, func(t *testing.T) { testPublicURL(t, ctx, pool, acquisition) })
	}
}

func testPublicURL(t *testing.T, ctx context.Context, pool *pgxpool.Pool, acquisition string) {
	t.Helper()
	var id string
	if err := pool.QueryRow(ctx, `INSERT INTO sources(name,url,license,status) VALUES ('Synthetic private URL',$1,'test only','approved') RETURNING id`, acquisition).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO source_artifacts(source_id,object_key,retrieved_at,sha256) VALUES ($1,$2,now(),repeat('c',64))`, id, id); err != nil {
		t.Fatal(err)
	}
	store := Store{Pool: pool}
	if rows, total, err := store.Sources(ctx, 25, 0); err != nil || total != 0 || len(rows) != 0 {
		t.Fatalf("source without reviewed public URL exposed: %v, count %d, error %v", rows, total, err)
	}
	for _, unsafe := range []string{acquisition, "https://example.invalid/?token=test", "https://example.invalid/#test", "https://example.invalid\\@evil.invalid", "https://exa mple.invalid", "javascript:alert(1)", "https://%65xample.invalid"} {
		if _, err := pool.Exec(ctx, `UPDATE sources SET public_url=$1 WHERE id=$2`, unsafe, id); err == nil {
			t.Errorf("unsafe public URL accepted: %s", unsafe)
		}
	}
	const publicURL = "https://example.invalid/datasets/collisions"
	if _, err := pool.Exec(ctx, `UPDATE sources SET public_url=$1 WHERE id=$2`, publicURL, id); err != nil {
		t.Fatal(err)
	}
	rows, total, err := store.Sources(ctx, 25, 0)
	if err != nil || total != 1 || len(rows) != 1 || rows[0].URL != publicURL {
		t.Fatalf("public landing page missing: %v, count %d, error %v", rows, total, err)
	}
	var retained string
	if err := pool.QueryRow(ctx, `SELECT url FROM sources WHERE id=$1`, id).Scan(&retained); err != nil || retained != acquisition {
		t.Fatalf("private provenance changed: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE sources SET status='withdrawn' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
}
