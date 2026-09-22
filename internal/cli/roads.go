package cli

import (
	"context"
	"corridor/internal/ingest"
	"corridor/internal/storage"
	"errors"
	"io"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ingestMovementInputs(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client) error {
	if err := ingestMigration(ctx, pool, objects); err != nil {
		return err
	}
	return ingestRoads(ctx, pool, objects)
}
func ingestRoads(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client) error {
	return ingestRoadInputs(ctx, pool, objects, "data/raw/prediction/pequop-roads-get.json", ingest.RoadArtifact())
}
func ingestRoadInputs(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client, path string, a ingest.Artifact) error {
	// #nosec G304 -- fixed operator path in production; injected synthetic fixture in tests.
	f, err := os.Open(path)
	if err != nil {
		return errors.New("pinned I-80 roads missing; see data/SOURCES.md")
	}
	defer f.Close()
	if err = ingest.Verify(f, a.SHA256); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	roads, err := ingest.ParseRoads(f)
	if err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if err = objects.Archive(ctx, a.ObjectKey, f, stat.Size()); err != nil {
		return err
	}
	return ingest.ImportRoads(ctx, pool, roads, a)
}
