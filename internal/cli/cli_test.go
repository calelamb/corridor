package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"corridor/internal/cli"
	"corridor/internal/explore"
	"corridor/internal/prioritize"
)

func TestPrioritize_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"prioritize", "--budget=1000000", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("prioritize command failed: %v, stderr: %s", err, stderr.String())
	}

	var candidates []prioritize.CandidateSite
	if err := json.Unmarshal(stdout.Bytes(), &candidates); err != nil {
		t.Fatalf("failed to parse JSON candidates: %v, raw: %s", err, stdout.String())
	}

	if len(candidates) == 0 {
		t.Fatalf("expected at least one candidate site for budget 1000000, got 0")
	}

	for _, c := range candidates {
		if c.ID.String() == "" {
			t.Errorf("candidate has empty ID")
		}
		if c.State != "WY" {
			t.Errorf("expected state WY, got %s", c.State)
		}
		if c.BCR < 1.0 {
			t.Errorf("expected BCR >= 1.0, got %.2f", c.BCR)
		}
		if c.EstCostUSD > 1000000.0 {
			t.Errorf("candidate cost %.2f exceeds budget 1000000.0", c.EstCostUSD)
		}
		if c.ExpectedAvoidedCollisions <= 0 {
			t.Errorf("expected positive avoided collisions, got %.2f", c.ExpectedAvoidedCollisions)
		}
	}
}

func TestPrioritize_Text(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"prioritize", "--budget=5000000", "--state=WY"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("prioritize text command failed: %v", err)
	}

	out := stdout.String()
	requiredHeaders := []string{"Site ID", "Highway", "Mileposts", "Cost", "Avoided Collisions", "BCR [P10-P90]"}
	for _, h := range requiredHeaders {
		if !strings.Contains(out, h) {
			t.Errorf("expected table header %q in output:\n%s", h, out)
		}
	}

	if !strings.Contains(out, "US-191") {
		t.Errorf("expected US-191 in output:\n%s", out)
	}
}

func TestPrioritize_InvalidOptions(t *testing.T) {
	ctx := context.Background()
	var stdout, stderr bytes.Buffer

	if err := cli.Run(ctx, []string{"prioritize", "--budget=-500"}, &stdout, &stderr); err == nil {
		t.Fatalf("expected error for negative budget")
	}

	if err := cli.Run(ctx, []string{"prioritize", "--min-bcr=-1"}, &stdout, &stderr); err == nil {
		t.Fatalf("expected error for negative min-bcr")
	}
}

func TestHotspot_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"hotspot", "--method=nkde", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("hotspot command failed: %v, stderr: %s", err, stderr.String())
	}

	var results []cli.HotspotOutput
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse JSON hotspots: %v, raw: %s", err, stdout.String())
	}

	if len(results) == 0 {
		t.Fatalf("expected hotspot results, got 0")
	}

	for _, h := range results {
		if h.SegmentID == "" {
			t.Errorf("hotspot item missing segment_id")
		}
		if h.Score < 0 {
			t.Errorf("expected non-negative score, got %f", h.Score)
		}
		if h.Significance == "" {
			t.Errorf("hotspot item missing significance")
		}
	}

	// Verify top hotspot has highest density score
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[0].Score {
			t.Errorf("results not sorted descending: result[0]=%f, result[%d]=%f",
				results[0].Score, i, results[i].Score)
		}
	}
}

func TestHotspot_GetisOrd(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"hotspot", "--method=getis_ord", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("hotspot getis_ord failed: %v", err)
	}

	var results []cli.HotspotOutput
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse Getis-Ord JSON: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected Getis-Ord hotspot results, got 0")
	}

	// Verify ZScore is populated
	if results[0].ZScore == nil {
		t.Fatalf("expected non-nil ZScore for Getis-Ord result")
	}
}

func TestHotspot_Text(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"hotspot", "--method=nkde"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("hotspot text output failed: %v", err)
	}

	out := stdout.String()
	requiredHeaders := []string{"Segment ID", "Score", "Raw Count", "Significance"}
	for _, h := range requiredHeaders {
		if !strings.Contains(out, h) {
			t.Errorf("expected table header %q in output:\n%s", h, out)
		}
	}
}

func TestHotspot_InvalidOptions(t *testing.T) {
	ctx := context.Background()
	var stdout, stderr bytes.Buffer

	if err := cli.Run(ctx, []string{"hotspot", "--method=unknown"}, &stdout, &stderr); err == nil {
		t.Fatalf("expected error for invalid method")
	}

	if err := cli.Run(ctx, []string{"hotspot", "--bandwidth=-100"}, &stdout, &stderr); err == nil {
		t.Fatalf("expected error for negative bandwidth")
	}
}

func TestEvidence_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"evidence", "--category=mitigation", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("evidence command failed: %v, stderr: %s", err, stderr.String())
	}

	var papers []explore.AcademicPaper
	if err := json.Unmarshal(stdout.Bytes(), &papers); err != nil {
		t.Fatalf("failed to parse JSON evidence: %v, raw: %s", err, stdout.String())
	}

	if len(papers) == 0 {
		t.Fatalf("expected mitigation papers, got 0")
	}

	for _, p := range papers {
		if p.Category != "mitigation" {
			t.Errorf("expected category 'mitigation', got %q", p.Category)
		}
		if p.Title == "" {
			t.Errorf("paper has empty title: %+v", p)
		}
		if p.KeyFindings == "" {
			t.Errorf("paper has empty key findings: %+v", p)
		}
		if p.AppliedInCorridor == "" {
			t.Errorf("paper has empty applied in corridor: %+v", p)
		}
	}
}

func TestEvidence_Text(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"evidence", "--category=mitigation"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("evidence text command failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Key Findings:") {
		t.Errorf("expected 'Key Findings:' in text output:\n%s", out)
	}
	if !strings.Contains(out, "Corridor Application:") {
		t.Errorf("expected 'Corridor Application:' in text output:\n%s", out)
	}
	if !strings.Contains(out, "Rytwinski") {
		t.Errorf("expected citation for Rytwinski in output:\n%s", out)
	}
}

func TestEvidence_AllAndEmpty(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	// All categories
	if err := cli.Run(ctx, []string{"evidence", "--category=all", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("evidence all categories failed: %v", err)
	}

	var allPapers []explore.AcademicPaper
	if err := json.Unmarshal(stdout.Bytes(), &allPapers); err != nil {
		t.Fatalf("failed to parse JSON evidence: %v", err)
	}
	if len(allPapers) < 15 {
		t.Errorf("expected at least 15 papers for category 'all', got %d", len(allPapers))
	}

	// Non-matching category
	stdout.Reset()
	if err := cli.Run(ctx, []string{"evidence", "--category=nonexistent"}, &stdout, &stderr); err != nil {
		t.Fatalf("evidence nonexistent category should handle gracefully: %v", err)
	}
	if !strings.Contains(stdout.String(), "No academic evidence found") {
		t.Errorf("expected 'No academic evidence found' message, got:\n%s", stdout.String())
	}
}

func TestHelpText(t *testing.T) {
	ctx := context.Background()

	// 1. Test corridorctl --help
	var helpOut, helpErr bytes.Buffer
	if err := cli.Run(ctx, []string{"--help"}, &helpOut, &helpErr); err != nil {
		t.Fatalf("corridorctl --help failed: %v", err)
	}
	helpStr := helpOut.String()
	for _, term := range []string{"corridorctl", "prioritize", "hotspot", "evidence", "help"} {
		if !strings.Contains(helpStr, term) {
			t.Errorf("expected %q in --help output:\n%s", term, helpStr)
		}
	}

	// 2. Test corridorctl -h
	helpOut.Reset()
	if err := cli.Run(ctx, []string{"-h"}, &helpOut, &helpErr); err != nil {
		t.Fatalf("corridorctl -h failed: %v", err)
	}
	if !strings.Contains(helpOut.String(), "corridorctl") {
		t.Errorf("expected corridorctl in -h output")
	}

	// 3. Test corridorctl help
	helpOut.Reset()
	if err := cli.Run(ctx, []string{"help"}, &helpOut, &helpErr); err != nil {
		t.Fatalf("corridorctl help failed: %v", err)
	}
	if !strings.Contains(helpOut.String(), "corridorctl") {
		t.Errorf("expected corridorctl in help output")
	}

	// 4. Test corridorctl unknown command outputs help text and returns error
	var unkOut, unkErr bytes.Buffer
	err := cli.Run(ctx, []string{"unknown-command"}, &unkOut, &unkErr)
	if err == nil {
		t.Fatalf("expected error for unknown-command, got nil")
	}
	unkStr := unkOut.String()
	if !strings.Contains(unkStr, "corridorctl") {
		t.Errorf("expected help text in output for unknown command, got:\n%s", unkStr)
	}
	if !strings.Contains(unkStr, "prioritize") {
		t.Errorf("expected prioritize in help text for unknown command")
	}

	// 5. Test corridorctl with empty args outputs help text and returns error
	var emptyOut, emptyErr bytes.Buffer
	if err := cli.Run(ctx, []string{}, &emptyOut, &emptyErr); err == nil {
		t.Fatalf("expected error for empty args, got nil")
	}
	if !strings.Contains(emptyOut.String(), "corridorctl") {
		t.Errorf("expected help text in output for empty args")
	}
}
