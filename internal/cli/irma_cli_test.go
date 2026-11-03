package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"corridor/internal/cli"
)

func TestIRMACLI(t *testing.T) {
	recent := time.Now().UTC().Add(-48 * time.Hour).Format("2006-01-02")
	const (
		sensitiveLat = "37.882145"
		sensitiveLon = "-119.261287"
		otherParkLat = "44.821012"
		otherParkLon = "-110.650034"
	)
	csvData := strings.Join([]string{
		"Park Code,Observation Date,Common Name,Road/Route Name,Milepost/Location Description,Latitude,Longitude,Cause of Death,Observation ID,Comments",
		"YOSE,2024-06-15,Mule deer,Big Oak Flat Road,MP 3.5,37.7511,-119.7891,Vehicle Strike,obs-yose-deer,synthetic mule deer",
		"YOSE," + recent + ",Sierra Nevada bighorn sheep,Tioga Pass Road,MP 14.2," + sensitiveLat + "," + sensitiveLon + ",Disease,obs-yose-bighorn,synthetic recent sensitive",
		"YELL,2024-05-10,Rocky Mountain elk,US-191,MP 22.0," + otherParkLat + "," + otherParkLon + ",Vehicle Strike,obs-yell-elk,synthetic other park",
		"",
	}, "\n")

	path := filepath.Join(t.TempDir(), "whmmo-synthetic.csv")
	if err := os.WriteFile(path, []byte(csvData), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := cli.Run(context.Background(), []string{
		"ingest", "irma", "--file", path, "--park", "YOSE", "--json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("ingest irma failed: %v\nstderr: %s", err, stderr.String())
	}

	raw := stdout.String()
	var summary struct {
		Source          string           `json:"source"`
		Park            string           `json:"park"`
		RecordsRead     int64            `json:"records_read"`
		Accepted        int64            `json:"accepted"`
		Rejected        int64            `json:"rejected"`
		Duplicates      int64            `json:"duplicates"`
		SensitiveCount  int64            `json:"sensitive_count"`
		RoadkillCount   int64            `json:"roadkill_count"`
		SpeciesCounts   map[string]int64 `json:"species_counts"`
		QuarantineDelay string           `json:"quarantine_delay"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("decode JSON summary: %v\nraw: %s", err, raw)
	}

	if summary.Park != "YOSE" {
		t.Errorf("park = %q, want YOSE", summary.Park)
	}
	if summary.RecordsRead != 3 {
		t.Errorf("records_read = %d, want 3", summary.RecordsRead)
	}
	if summary.Accepted != 2 {
		t.Errorf("accepted = %d, want 2 after YOSE filter", summary.Accepted)
	}
	if summary.Rejected != 0 || summary.Duplicates != 0 {
		t.Errorf("rejected=%d duplicates=%d, want 0 and 0", summary.Rejected, summary.Duplicates)
	}
	if summary.SensitiveCount != 1 {
		t.Errorf("sensitive_count = %d, want 1", summary.SensitiveCount)
	}
	if summary.RoadkillCount != 1 {
		t.Errorf("roadkill_count = %d, want 1 (YOSE mule deer only)", summary.RoadkillCount)
	}
	if summary.SpeciesCounts["mule deer"] != 1 {
		t.Errorf("species_counts[mule deer] = %d, want 1", summary.SpeciesCounts["mule deer"])
	}
	if summary.SpeciesCounts["sierra nevada bighorn sheep"] != 1 {
		t.Errorf("species_counts[sierra nevada bighorn sheep] = %d, want 1; counts=%v", summary.SpeciesCounts["sierra nevada bighorn sheep"], summary.SpeciesCounts)
	}
	if summary.SpeciesCounts["elk"] != 0 {
		t.Errorf("park filter kept YELL elk: %v", summary.SpeciesCounts)
	}
	if !strings.Contains(summary.QuarantineDelay, "30-day") {
		t.Errorf("quarantine_delay = %q, want 30-day notice", summary.QuarantineDelay)
	}

	for _, secret := range []string{sensitiveLat, sensitiveLon, otherParkLat, otherParkLon} {
		if strings.Contains(raw, secret) {
			t.Errorf("public JSON leaked coordinate %s:\n%s", secret, raw)
		}
	}
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "latitude") || strings.Contains(lower, "longitude") {
		t.Errorf("public JSON printed coordinate fields:\n%s", raw)
	}
}
