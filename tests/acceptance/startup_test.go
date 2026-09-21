//go:build acceptance

package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFreshPostgresWaitsForTCP(t *testing.T) {
	compose := delayedPostgres(t)
	compose("up", "-d", "postgres")
	deadline := time.Now().Add(30 * time.Second)
	for !strings.Contains(compose("logs", "postgres"), "CORRIDOR_TEST_INIT_PAUSED") {
		if time.Now().After(deadline) {
			t.Fatalf("initialization did not reach the test gate: %s", compose("logs", "postgres"))
		}
		time.Sleep(100 * time.Millisecond)
	}
	id := strings.TrimSpace(compose("ps", "-q", "postgres"))
	for {
		out, err := exec.Command("docker", "inspect", "--format", "{{json .State.Health}}", id).Output()
		if err != nil {
			t.Fatal(err)
		}
		var health struct {
			Status string
			Log    []json.RawMessage
		}
		if err := json.Unmarshal(out, &health); err != nil {
			t.Fatal(err)
		}
		if len(health.Log) > 0 {
			if health.Status == "healthy" {
				t.Fatal("socket-only initialization server was marked healthy before TCP startup")
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("health probe did not run")
		}
		time.Sleep(100 * time.Millisecond)
	}
	compose("exec", "-T", "postgres", "touch", "/tmp/release-init")
	compose("up", "-d", "--wait", "postgres")
	compose("up", "--abort-on-container-exit", "--exit-code-from", "migrate", "migrate")
}

func delayedPostgres(t *testing.T) func(...string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "pause.sql")
	if err := os.WriteFile(script, []byte("DO $$ BEGIN RAISE NOTICE 'CORRIDOR_TEST_INIT_PAUSED'; WHILE pg_stat_file('/tmp/release-init',true) IS NULL LOOP PERFORM pg_sleep(0.1); END LOOP; END $$;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	override := filepath.Join(dir, "compose.yaml")
	content, err := json.Marshal(map[string]any{"services": map[string]any{"postgres": map[string]any{"volumes": []string{script + ":/docker-entrypoint-initdb.d/99-test-pause.sql:ro"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(override, content, 0o600); err != nil {
		t.Fatal(err)
	}
	project := fmt.Sprintf("corridor-startup-test-%d", time.Now().UnixNano())
	base := []string{"compose", "--project-directory", "../..", "-p", project, "-f", "../../compose.yaml", "-f", override}
	run := func(args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		return exec.CommandContext(ctx, "docker", append(append([]string{}, base...), args...)...).CombinedOutput()
	}
	t.Cleanup(func() {
		if out, err := run("down", "-v", "--timeout", "1"); err != nil {
			t.Errorf("cleanup isolated startup test: %s %v", out, err)
		}
	})
	return func(args ...string) string {
		out, err := run(args...)
		if err != nil {
			t.Fatalf("startup test command %v: %s %v", args, out, err)
		}
		return string(out)
	}
}
