package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"corridor/internal/cli"
	"corridor/internal/prioritize"
)

func TestPrioritize_MultiPark(t *testing.T) {
	ctx := context.Background()

	parks := []struct {
		name      string
		parkFlag  string
		wantState string
		wantHwy   string
	}{
		{"Yosemite by code", "YOSE", "CA", "CA-120"},
		{"Yosemite by name", "Yosemite", "CA", "CA-120"},
		{"Yellowstone by code", "YELL", "WY", "US-191"},
		{"Yellowstone by name", "Yellowstone", "WY", "US-191"},
		{"Grand Teton by code", "GRTE", "WY", "US-89"},
		{"Grand Teton by name", "Grand Teton", "WY", "US-89"},
		{"Glacier by code", "GLAC", "MT", "US-2"},
		{"Glacier by name", "Glacier", "MT", "US-2"},
		{"Rocky Mountain by code", "ROMO", "CO", "US-34"},
		{"Rocky Mountain by name", "Rocky Mountain", "CO", "US-34"},
	}

	for _, tc := range parks {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Run(ctx, []string{"prioritize", "--park=" + tc.parkFlag, "--json"}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("prioritize --park=%s failed: %v, stderr: %s", tc.parkFlag, err, stderr.String())
			}

			var candidates []prioritize.CandidateSite
			if err := json.Unmarshal(stdout.Bytes(), &candidates); err != nil {
				t.Fatalf("unmarshal candidates JSON failed: %v, raw: %s", err, stdout.String())
			}

			if len(candidates) == 0 {
				t.Fatalf("expected at least 1 candidate for park %s, got 0", tc.parkFlag)
			}

			c := candidates[0]
			if c.State != tc.wantState {
				t.Errorf("expected state %s, got %s", tc.wantState, c.State)
			}
			if !strings.EqualFold(c.HighwayRef, tc.wantHwy) {
				t.Errorf("expected highway %s, got %s", tc.wantHwy, c.HighwayRef)
			}
			if c.BCR < 1.0 {
				t.Errorf("expected BCR >= 1.0, got %.2f", c.BCR)
			}
			if c.ExpectedAvoidedCollisions <= 0 {
				t.Errorf("expected positive avoided collisions, got %.2f", c.ExpectedAvoidedCollisions)
			}
			if c.EstCostUSD <= 0 {
				t.Errorf("expected positive estimated cost, got %.2f", c.EstCostUSD)
			}
		})
	}
}

func TestPrioritize_StateResolution(t *testing.T) {
	ctx := context.Background()

	states := []struct {
		state   string
		wantHwy string
	}{
		{"CA", "CA-120"},
		{"MT", "US-2"},
		{"CO", "US-34"},
		{"WY", "US-191"},
	}

	for _, tc := range states {
		t.Run(tc.state, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Run(ctx, []string{"prioritize", "--state=" + tc.state, "--json"}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("prioritize --state=%s failed: %v", tc.state, err)
			}

			var candidates []prioritize.CandidateSite
			if err := json.Unmarshal(stdout.Bytes(), &candidates); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if len(candidates) == 0 {
				t.Fatalf("expected candidates for state %s, got 0", tc.state)
			}

			for _, c := range candidates {
				if c.State != tc.state {
					t.Errorf("expected candidate state %s, got %s", tc.state, c.State)
				}
				if c.BCR < 1.0 {
					t.Errorf("expected BCR >= 1.0, got %.2f", c.BCR)
				}
			}
		})
	}
}

func TestHotspot_MultiPark(t *testing.T) {
	ctx := context.Background()

	parks := []struct {
		name     string
		parkFlag string
		prefixID string
	}{
		{"YOSE", "YOSE", "ca-120-seg-"},
		{"YELL", "YELL", "seg-"},
		{"GRTE", "GRTE", "grte-seg-"},
		{"GLAC", "GLAC", "glac-seg-"},
		{"ROMO", "ROMO", "romo-seg-"},
	}

	for _, tc := range parks {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Run(ctx, []string{"hotspot", "--park=" + tc.parkFlag, "--json"}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("hotspot --park=%s failed: %v", tc.parkFlag, err)
			}

			var results []cli.HotspotOutput
			if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
				t.Fatalf("unmarshal hotspots failed: %v", err)
			}

			if len(results) == 0 {
				t.Fatalf("expected hotspots for park %s, got 0", tc.parkFlag)
			}

			if !strings.HasPrefix(results[0].SegmentID, tc.prefixID) {
				t.Errorf("expected segment ID prefix %s, got %s", tc.prefixID, results[0].SegmentID)
			}
			if results[0].Score <= 0 {
				t.Errorf("expected positive score, got %f", results[0].Score)
			}
			if results[0].RawCount <= 0 {
				t.Errorf("expected positive raw count, got %d", results[0].RawCount)
			}
		})
	}
}

func TestHotspot_RequestedParkWithoutCorridorIsUnavailable(t *testing.T) {
	ctx := context.Background()
	for _, args := range [][]string{
		{"hotspot", "--park=OLYM", "--json"},
		{"hotspot", "--park=Olympic", "--json"},
		{"hotspot", "--state=WA", "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := cli.Run(ctx, args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("expected unavailable hotspot data, got %s", stdout.String())
			}
			if !strings.Contains(strings.ToLower(err.Error()), "unavailable") {
				t.Fatalf("expected unavailable error, got %v", err)
			}
			if strings.Contains(stdout.String(), "seg-") {
				t.Fatalf("unknown request returned Yellowstone segments: %s", stdout.String())
			}
		})
	}
}
