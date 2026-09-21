package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type brokenEntropy struct{}

func (brokenEntropy) Read([]byte) (int, error) { return 0, errors.New("no entropy") }
func TestInitDev(t *testing.T) {
	target := filepath.Join(t.TempDir(), ".env")
	var out bytes.Buffer
	if err := initDev(target, strings.NewReader(strings.Repeat("x", 128)), &out); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("mode: %v %v", info, err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("POSTGRES_PASSWORD=")) {
		t.Fatal("missing configuration")
	}
	if err := initDev(target, strings.NewReader(strings.Repeat("y", 128)), &out); err == nil {
		t.Fatal("overwrote existing env")
	}
	if strings.Contains(out.String(), "787878") {
		t.Fatal("printed secret")
	}
	if err := initDev(filepath.Join(t.TempDir(), ".env"), brokenEntropy{}, io.Discard); err == nil {
		t.Fatal("entropy error ignored")
	}
}
func TestCommands(t *testing.T) {
	if Run(context.Background(), []string{"help"}, io.Discard, io.Discard) != nil {
		t.Fatal("help")
	}
	for _, args := range [][]string{{}, {"ingest"}, {"migrate"}} {
		if Run(context.Background(), args, io.Discard, io.Discard) == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestProbeFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Run(ctx, []string{"health"}, io.Discard, io.Discard) == nil {
		t.Fatal("canceled probe succeeded")
	}
}

func TestSetupValidation(t *testing.T) {
	if Run(context.Background(), []string{"setup"}, io.Discard, io.Discard) == nil {
		t.Fatal("missing setup config accepted")
	}
}
func TestInitDevCommand(t *testing.T) {
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	if err := Run(context.Background(), []string{"init-dev"}, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationFailure(t *testing.T) {
	t.Setenv("CORRIDOR_MIGRATION_DATABASE_URL", "invalid")
	if Run(context.Background(), []string{"migrate"}, io.Discard, io.Discard) == nil {
		t.Fatal("invalid migration connection accepted")
	}
}
