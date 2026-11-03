package cli

import (
	"bytes"
	"context"
	"corridor/internal/ingest"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestIngestNPSRoadsMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := IngestNPSRoads(context.Background(), nil, nil, "nonexistent.json"); err == nil {
		t.Fatal("missing nps roads accepted")
	}
}

func TestIngestSubcommands(t *testing.T) {
	ctx := context.Background()

	// Test help
	if err := Run(ctx, []string{"ingest", "help"}, io.Discard, io.Discard); err != nil {
		t.Fatalf("ingest help failed: %v", err)
	}

	// Test unknown subcommand
	if err := Run(ctx, []string{"ingest", "unknown"}, io.Discard, io.Discard); err == nil {
		t.Fatal("expected error for unknown subcommand")
	}

	// Test irma missing file flag
	if err := Run(ctx, []string{"ingest", "irma"}, io.Discard, io.Discard); err == nil {
		t.Fatal("expected error for missing --file flag")
	}

	// Test mdt missing file flag
	if err := Run(ctx, []string{"ingest", "mdt"}, io.Discard, io.Discard); err == nil {
		t.Fatal("expected error for missing --file flag")
	}

	// Test caltrans missing file and url flag
	if err := Run(ctx, []string{"ingest", "caltrans"}, io.Discard, io.Discard); err == nil {
		t.Fatal("expected error for missing --file or --url flag")
	}

	for _, name := range []string{"wydot", "udot", "idfg"} {
		if err := Run(ctx, []string{"ingest", name}, io.Discard, io.Discard); err == nil {
			t.Fatalf("expected error for missing --file flag on %s", name)
		}
	}
}

func TestIngestStateFeedSummaryOmitsCoordinates(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Now().UTC().UnixMilli()
	idfgPath := filepath.Join(dir, "idfg.json")
	body := fmt.Sprintf(`{"features":[{"attributes":{"OBJECTID":2,"species":"GRAY WOLF","observed":%d,"highway":"US-20","latitude":44.41,"longitude":-111.35}}]}`, now)
	if err := os.WriteFile(idfgPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := Run(ctx, []string{"ingest", "idfg", "--file", idfgPath, "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("idfg ingest: %v stderr %s", err, stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "-111.35") || strings.Contains(out, "latitude") {
		t.Fatalf("summary leaked a coordinate: %s", out)
	}
	if !strings.Contains(out, `"withheld_sensitive": 1`) {
		t.Fatalf("recent wolf was not withheld: %s", out)
	}
}
