//go:build acceptance

package acceptance

import (
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDependencyOutage(t *testing.T) {
	compose := func(args ...string) error {
		return exec.Command("docker", append([]string{"compose", "-p", "corridor-foundation-acceptance"}, args...)...).Run()
	}
	if err := compose("stop", "postgres"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := compose("up", "-d", "--no-deps", "--wait", "postgres"); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Timeout: 5 * time.Second}
	for _, path := range []string{"/readyz", "/v1/coverage", "/v1/sources"} {
		resp, err := client.Get("http://127.0.0.1:8080" + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 503 || strings.Contains(string(body), `"state":"empty"`) {
			t.Fatalf("outage %s: %d %s", path, resp.StatusCode, body)
		}
	}
	resp, err := client.Get("http://127.0.0.1:8080/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("process liveness depends on database")
	}
	if err := compose("exec", "-T", "corridor", "corridorctl", "health"); err == nil {
		t.Fatal("health command accepted dependency outage")
	}
}
