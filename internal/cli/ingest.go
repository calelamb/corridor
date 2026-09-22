package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"corridor/internal/config"
	"corridor/internal/db"
	"corridor/internal/ingest"
	"corridor/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ingestAll(ctx context.Context, out io.Writer) error {
	return ingestInputs(ctx, out, "data/raw/sweep-2026-09-21/global-roadkill-v5.csv", ingest.PilotArtifact(), ingestMovementInputs)
}
func ingestInputs(ctx context.Context, out io.Writer, path string, artifact ingest.Artifact, movement func(context.Context, *pgxpool.Pool, *storage.Client) error) error {
	// #nosec G304 -- fixed operator path in production; dependency injection for isolated fixtures.
	file, err := os.Open(path)
	if err != nil {
		return errors.New("pilot artifact missing; see data/SOURCES.md acquisition instructions")
	}
	defer file.Close()
	if err = ingest.Verify(file, artifact.SHA256); err != nil {
		return err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind source: %w", err)
	}
	dataset, err := ingest.ParseRoadkill(file)
	if err != nil {
		return err
	}
	if len(dataset.Records) == 0 {
		return errors.New("pilot has no accepted records")
	}
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	cfg.S3AccessKey = os.Getenv("MINIO_ROOT_USER")
	cfg.S3SecretKey = os.Getenv("MINIO_ROOT_PASSWORD")
	objects, err := storage.New(cfg)
	if err != nil {
		return errors.New("ingestion requires administrator storage credentials")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind archive: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("inspect artifact: %w", err)
	}
	if err = objects.Archive(ctx, artifact.ObjectKey, file, stat.Size()); err != nil {
		return err
	}
	dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
	if dsn == "" {
		return errors.New("ingestion requires CORRIDOR_MIGRATION_DATABASE_URL")
	}
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err = ingest.Import(ctx, pool, dataset, artifact); err != nil {
		return err
	}
	if err = movement(ctx, pool, objects); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Source  string    `json:"source"`
		QA      ingest.QA `json:"qa"`
		Skipped []string  `json:"skipped"`
	}{artifact.Name, dataset.QA, []string{"Montana: rights unresolved", "California derivative: quarantined", "BC WARS: redistribution restricted", "telemetry: CRS/time validation pending", "other research candidates: no adapter or artifact-level clearance"}})
}

func ingestMigration(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client) error {
	const path = "data/raw/exploration/pequop.shp"
	file, err := os.Open(path)
	if err != nil {
		return errors.New("migration shape missing; see data/SOURCES.md")
	}
	defer file.Close()
	a := ingest.MigrationArtifact()
	if err = ingest.Verify(file, a.SHA256); err != nil {
		return err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if err = objects.Archive(ctx, a.ObjectKey, file, stat.Size()); err != nil {
		return err
	}
	_, err = ingest.ImportMigration(ctx, pool, path)
	return err
}
