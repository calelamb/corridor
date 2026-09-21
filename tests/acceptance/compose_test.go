//go:build acceptance

package acceptance

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCompose(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	for _, path := range []string{"/", "/data/", "/readyz", "/v1/coverage"} {
		resp, err := client.Get("http://127.0.0.1:8080" + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("%s: %d %v", path, resp.StatusCode, err)
		}
		if path == "/v1/coverage" {
			var result struct {
				Status string
				Data   struct {
					State          string
					IngestedEvents int `json:"ingested_events"`
				}
			}
			if err := json.Unmarshal(body, &result); err != nil || result.Status != "success" || result.Data.State != "empty" || result.Data.IngestedEvents != 0 {
				t.Fatalf("coverage %s %v", body, err)
			}
		}
	}
	if out, err := exec.Command("docker", "compose", "-p", "corridor-foundation-acceptance", "exec", "-T", "corridor", "corridorctl", "health").CombinedOutput(); err != nil {
		t.Fatalf("health probe: %s %v", out, err)
	}
	for _, path := range []string{"/v1/missing", "/tiles/missing.mvt", "/_app/missing.js"} {
		resp, err := client.Get("http://127.0.0.1:8080" + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 404 || strings.Contains(string(body), "<html") {
			t.Fatalf("fallback %s", path)
		}
	}
}
