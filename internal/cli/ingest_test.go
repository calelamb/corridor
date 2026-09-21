package cli

import (
	"context"
	"corridor/internal/ingest"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestIngestMissingAndChangedArtifact(t *testing.T) {
	t.Chdir(t.TempDir())
	if Run(context.Background(), []string{"ingest", "all"}, io.Discard, io.Discard) == nil {
		t.Fatal("missing artifact accepted")
	}
	path := filepath.Join(t.TempDir(), "bad.csv")
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if ingestInputs(context.Background(), io.Discard, path, ingest.PilotArtifact(), nil) == nil {
		t.Fatal("bad checksum accepted")
	}
}

func TestIngestInputFailures(t *testing.T) {
	for _, raw := range []string{"broken,header\n", "occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\n", "occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\nx,US,Synthetic deer,-115,43,30,1,2014,2,2,Synthetic road,Systematic,\n"} {
		path := filepath.Join(t.TempDir(), "synthetic.csv")
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256([]byte(raw))
		a := ingest.PilotArtifact()
		a.SHA256 = hex.EncodeToString(digest[:])
		t.Setenv("CORRIDOR_DATABASE_URL", "")
		if err := ingestInputs(context.Background(), io.Discard, path, a, nil); err == nil {
			t.Fatal("malformed source or missing configuration accepted")
		}
	}
}
func TestMigrationArtifactFailures(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := ingestMigration(context.Background(), nil, nil); err == nil {
		t.Fatal("missing migration accepted")
	}
	if err := os.MkdirAll("data/raw/exploration", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("data/raw/exploration/pequop.shp", []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ingestMigration(context.Background(), nil, nil); err == nil {
		t.Fatal("changed migration accepted")
	}
}
