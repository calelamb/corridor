package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Artifact struct {
	SourceID, Name, URL, License, SHA256, ObjectKey string
	RetrievedAt                                     time.Time
}

func PilotArtifact() Artifact {
	return Artifact{
		SourceID: "e94ccfd0-2a16-4549-a6ee-90c6ff60a00a", Name: "Global Roadkill v5 · US", URL: "https://doi.org/10.6084/m9.figshare.25714233.v5", License: "CC BY 4.0 · Grilo et al. · modified: US subset and H3 aggregation",
		SHA256: "ba005179e7f02bae33ba4b9374b6957cdf795f04a8ada773ef4887fe61a46c16", ObjectKey: "sha256/ba005179e7f02bae33ba4b9374b6957cdf795f04a8ada773ef4887fe61a46c16/global-roadkill-v5.csv", RetrievedAt: time.Date(2026, 9, 21, 21, 4, 30, 0, time.UTC)}
}
func Verify(input io.Reader, expected string) error {
	hash := sha256.New()
	if _, err := io.Copy(hash, input); err != nil {
		return fmt.Errorf("hash artifact: %w", err)
	}
	if hex.EncodeToString(hash.Sum(nil)) != expected {
		return errors.New("artifact checksum mismatch")
	}
	return nil
}
func Import(ctx context.Context, pool *pgxpool.Pool, data Dataset, a Artifact) error {
	if len(a.SHA256) != 64 || a.ObjectKey == "" || a.RetrievedAt.IsZero() {
		return errors.New("artifact metadata required")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin import: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827365)"); err != nil {
		return fmt.Errorf("lock import: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO sources(id,name,url,license,status,public_url) VALUES($1,$2,$3,$4,'approved',$3) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,license=EXCLUDED.license,public_url=EXCLUDED.public_url`, a.SourceID, a.Name, a.URL, a.License); err != nil {
		return fmt.Errorf("register source: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO source_artifacts(source_id,object_key,retrieved_at,sha256) VALUES($1,$2,$3,$4) ON CONFLICT(source_id,sha256) DO NOTHING`, a.SourceID, a.ObjectKey, a.RetrievedAt, a.SHA256); err != nil {
		return fmt.Errorf("register artifact: %w", err)
	}
	var artifactID string
	if err = tx.QueryRow(ctx, "SELECT id::text FROM source_artifacts WHERE source_id=$1 AND sha256=$2", a.SourceID, a.SHA256).Scan(&artifactID); err != nil {
		return fmt.Errorf("resolve artifact: %w", err)
	}
	if err = insertRecords(ctx, tx, data.Records, a.SourceID, artifactID); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	return nil
}
func insertRecords(ctx context.Context, tx pgx.Tx, records []Record, source, artifact string) error {
	batch := &pgx.Batch{}
	for _, r := range records {
		batch.Queue(`INSERT INTO wildlife_observations(source_id,artifact_id,native_id,species,road,survey,geom,uncertainty_m,period_start,period_end,time_precision,quantity,reference_text)
 VALUES($1,$2,$3,$4,$5,$6,ST_SetSRID(ST_MakePoint($7,$8),4326),$9,$10,$11,$12,$13,$14) ON CONFLICT(source_id,native_id) DO NOTHING`, source, artifact, r.ID, r.Species, r.Road, r.Survey, r.Longitude, r.Latitude, r.Uncertainty, r.Start, r.End, r.Precision, r.Quantity, r.Reference)
	}
	results := tx.SendBatch(ctx, batch)
	if err := results.Close(); err != nil {
		return fmt.Errorf("store normalized observations: %w", err)
	}
	return nil
}
