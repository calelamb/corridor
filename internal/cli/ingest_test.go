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

	for _, name := range []string{"wydot", "udot", "idfg", "banff", "ncdot", "vdot"} {
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

func TestIngestStateFeedBanffSummaryOmitsGeometry(t *testing.T) {
	// Synthetic fixture for this test only: one data row, explicit 0 deer, blank grizzly.
	csv := "Region,Location,Year images were taken,Date of first image,Date of last image," +
		"Number of hikers and runners,Number of horse riders,Number of bicyclists,Number of cougars," +
		"Number of elk,Number of grizzly bears,Number of lynx,Number of wolves,Number of deer\n" +
		"Synthetic Region,Synthetic Structure,2015,1/6/2015,30/9/2015,1,2,3,4,5,,7,8,0\n"
	path := filepath.Join(t.TempDir(), "banff.csv")
	if err := os.WriteFile(path, []byte(csv), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"ingest", "banff", "--file", path, "--json"}, &stdout, io.Discard); err != nil {
		t.Fatalf("banff ingest: %v", err)
	}
	out := stdout.String()
	for _, banned := range []string{"latitude", "longitude", "geometry"} {
		if strings.Contains(strings.ToLower(out), banned) {
			t.Fatalf("banff summary contains %q: %s", banned, out)
		}
	}
	if !strings.Contains(out, `"records_read": 1`) || !strings.Contains(out, "Open Government Licence") {
		t.Fatalf("unexpected banff summary: %s", out)
	}
	if !strings.Contains(out, `"grizzly_counts_known": 0`) {
		t.Fatalf("blank grizzly count reported as known: %s", out)
	}
}

func TestIngestStateFeedVDOTWithholdsRecentBear(t *testing.T) {
	// Synthetic fixture for this test only; coordinates are not real locations.
	const inputLongitude = "-100.123456"
	now := time.Now().UTC().UnixMilli()
	body := fmt.Sprintf(`{"features":[{"attributes":{"objectid":1,"starttime":%d,"wildlife":"Bear","rtenamedir":"SYNTHETIC-ROUTE","latitude_d":10.123456,"longitude_d":%s}}]}`, now, inputLongitude)
	path := filepath.Join(t.TempDir(), "vdot.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"ingest", "vdot", "--file", path, "--json"}, &stdout, io.Discard); err != nil {
		t.Fatalf("vdot ingest: %v", err)
	}
	out := stdout.String()
	if strings.Contains(out, inputLongitude) || strings.Contains(out, "10.123456") || strings.Contains(out, "longitude") {
		t.Fatalf("vdot summary leaked a coordinate: %s", out)
	}
	if !strings.Contains(out, `"withheld_sensitive": 1`) || !strings.Contains(out, `"public_records": 0`) {
		t.Fatalf("recent bear was not withheld: %s", out)
	}
}
