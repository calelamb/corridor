package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestSeedSegmentsCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	// Calling seed-segments without environment variables should fail gracefully
	t.Setenv("CORRIDOR_MIGRATION_DATABASE_URL", "")
	t.Setenv("CORRIDOR_DATABASE_URL", "")
	err := Run(context.Background(), []string{"seed-segments"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when running seed-segments without DB URL")
	}
}

func TestSeedSegmentsNilInputs(t *testing.T) {
	ctx := context.Background()
	var stdout bytes.Buffer

	if err := SeedSegments(nil, nil, &stdout); err == nil {
		t.Fatal("expected error with nil context")
	}

	if err := SeedSegments(ctx, nil, &stdout); err == nil {
		t.Fatal("expected error with nil pool")
	}
}

func TestSeedSegmentsCommandInvalidDSN(t *testing.T) {
	var stdout, stderr bytes.Buffer
	t.Setenv("CORRIDOR_MIGRATION_DATABASE_URL", "postgres://invalid:5432/nonexistent?connect_timeout=1")
	err := Run(context.Background(), []string{"seed-segments"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when running seed-segments with invalid DSN")
	}
}
