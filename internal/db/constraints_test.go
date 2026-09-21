//go:build integration

package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
)

func testConstraints(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	const source = "11111111-1111-4111-8111-111111111111"
	_, err := pool.Exec(ctx, `INSERT INTO sources(id,name,url,license,status) VALUES ($1,'Synthetic test source','https://example.invalid','test only','approved')`, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO species(id,name,sensitive) VALUES ('22222222-2222-4222-8222-222222222222','Synthetic sensitive test',true)"); err != nil {
		t.Fatal(err)
	}
	invalid := []string{
		`INSERT INTO road_segments(source_id,native_id,geom,length_m) VALUES ('` + source + `','bad',ST_GeomFromText('LINESTRING(0 0,1 1)',4326),-1)`,
		`INSERT INTO events(source_id,native_id,geom,observed_at,time_precision,uncertainty_m) VALUES ('` + source + `','bad',ST_Point(0,0,4326),now(),'day',-1)`,
		`INSERT INTO aggregate_observations(source_id,native_id,geom,period_start,period_end,event_count) VALUES ('` + source + `','bad',ST_Point(0,0,4326),now(),now(),-1)`,
	}
	for _, query := range invalid {
		if _, err := pool.Exec(ctx, query); err == nil {
			t.Fatal("invalid fixture accepted")
		}
	}
	_, err = pool.Exec(ctx, `INSERT INTO source_artifacts(source_id,object_key,retrieved_at,sha256) VALUES ($1,'synthetic-test',now(),repeat('a',64))`, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE source_artifacts SET object_key='changed'"); err == nil {
		t.Fatal("artifact mutated")
	}
	if _, err := pool.Exec(ctx, "DELETE FROM source_artifacts"); err == nil {
		t.Fatal("artifact deleted")
	}
	_, err = pool.Exec(ctx, `INSERT INTO events(source_id,native_id,species_id,geom,observed_at,time_precision,uncertainty_m,publishable) VALUES ($1,'delayed','22222222-2222-4222-8222-222222222222',ST_Point(0,0,4326),now(),'day',100,true)`, source)
	if err != nil {
		t.Fatal(err)
	}
	store := Store{Pool: pool}
	count, err := store.EventCount(ctx)
	if err != nil || count != 0 {
		t.Fatalf("withheld record leaked %d %v", count, err)
	}
	_, err = pool.Exec(ctx, `UPDATE events SET observed_at=now()-interval '31 days'`)
	if err != nil {
		t.Fatal(err)
	}
	count, err = store.EventCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("released count %d %v", count, err)
	}
}
