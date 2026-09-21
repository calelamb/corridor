//go:build integration

package cli

import (
	"bytes"
	"context"
	"corridor/internal/db"
	"corridor/internal/ingest"
	"corridor/internal/storage"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIngestPipelinePrivateAndIdempotent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg, err := postgres.Run(ctx, "corridor-postgres:foundation", postgres.WithDatabase("corridor_test"), postgres.WithUsername("test_admin"), postgres.WithPassword("synthetic-test-only"), postgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Terminate(context.Background())
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	mini, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Image: "corridor-minio:foundation", ExposedPorts: []string{"9000/tcp"}, Env: map[string]string{"MINIO_ROOT_USER": "synthetic-admin", "MINIO_ROOT_PASSWORD": "synthetic-admin-password"}, WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp")}, Started: true})
	if err != nil {
		t.Fatal(err)
	}
	defer mini.Terminate(context.Background())
	host, err := mini.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := mini.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{"CORRIDOR_ENV": "development", "CORRIDOR_DATABASE_URL": dsn, "CORRIDOR_MIGRATION_DATABASE_URL": dsn, "CORRIDOR_DB_PASSWORD": "synthetic-application-password-long", "CORRIDOR_S3_ENDPOINT": "http://" + host + ":" + port.Port(), "CORRIDOR_S3_BUCKET": "synthetic-private", "CORRIDOR_S3_ACCESS_KEY": "synthetic-app", "CORRIDOR_S3_SECRET_KEY": "synthetic-app-password", "MINIO_ROOT_USER": "synthetic-admin", "MINIO_ROOT_PASSWORD": "synthetic-admin-password"} {
		t.Setenv(k, v)
	}
	if err := Run(ctx, []string{"migrate"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := setup(ctx); err != nil {
		t.Fatal(err)
	}
	raw := []byte("occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\nsynth,US,Synthetic deer,-115,43,30,2,2014,2,2,Synthetic road,Systematic,\n")
	path := filepath.Join(t.TempDir(), "synthetic.csv")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	a := ingest.PilotArtifact()
	a.SHA256 = hex.EncodeToString(digest[:])
	a.ObjectKey = "sha256/" + a.SHA256 + "/synthetic.csv"
	movement := func(context.Context, *pgxpool.Pool, *storage.Client) error { return nil }
	for range 2 {
		var out bytes.Buffer
		if err = ingestInputs(ctx, &out, path, a, movement); err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(out.Bytes(), []byte(`"Accepted":1`)) {
			t.Fatalf("missing QA %s", out.String())
		}
	}
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var count int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM wildlife_observations").Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate import %d %v", count, err)
	}
	t.Setenv("CORRIDOR_MIGRATION_DATABASE_URL", "")
	if ingestInputs(ctx, io.Discard, path, a, movement) == nil {
		t.Fatal("missing migration DSN accepted")
	}
	t.Setenv("CORRIDOR_MIGRATION_DATABASE_URL", "invalid")
	if ingestInputs(ctx, io.Discard, path, a, movement) == nil {
		t.Fatal("invalid migration DSN accepted")
	}
	t.Setenv("MINIO_ROOT_PASSWORD", "")
	if ingestInputs(ctx, io.Discard, path, a, movement) == nil {
		t.Fatal("missing ingestion credentials accepted")
	}

}
