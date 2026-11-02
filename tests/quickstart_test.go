// tests/quickstart_test.go
package tests

import (
	"os"
	"strings"
	"testing"
)

func TestNPSRangerQuickstartGuide(t *testing.T) {
	candidates := []string{
		"../docs/outreach/NPS_RANGER_QUICKSTART_GUIDE.md",
		"docs/outreach/NPS_RANGER_QUICKSTART_GUIDE.md",
	}

	var data []byte
	var readErr error
	var foundPath string
	for _, p := range candidates {
		data, readErr = os.ReadFile(p)
		if readErr == nil {
			foundPath = p
			break
		}
	}

	if readErr != nil {
		t.Fatalf("failed to read NPS ranger quickstart guide from %v: %v", candidates, readErr)
	}

	if len(data) <= 1000 {
		t.Errorf("file %s is too short: got %d bytes, want > 1000 bytes", foundPath, len(data))
	}

	content := string(data)

	requiredSections := []string{
		"1. Zero-Admin Desktop Setup",
		"2. Offline Field Patrol Logging (<10 Seconds)",
		"3. Dual-Resolution Poaching Privacy",
		"4. FHWA Grant Brief Export",
	}

	for _, sec := range requiredSections {
		if !strings.Contains(content, sec) {
			t.Errorf("missing required section: %q", sec)
		}
	}

	// Verify inclusion of keyboard shortcuts
	lower := strings.ToLower(content)
	if !strings.Contains(lower, "keyboard shortcut") {
		t.Errorf("missing keyboard shortcuts section or reference in %s", foundPath)
	}

	// Verify dead-zone queue instructions
	if !strings.Contains(lower, "dead zone") && !strings.Contains(lower, "dead-zone") {
		t.Errorf("missing dead-zone instructions in %s", foundPath)
	}
	if !strings.Contains(lower, "queue") {
		t.Errorf("missing queue instructions in %s", foundPath)
	}
}
