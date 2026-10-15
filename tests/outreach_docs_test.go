// tests/outreach_docs_test.go
package tests

import (
	"os"
	"strings"
	"testing"
)

func TestOutreachDocumentation(t *testing.T) {
	files := []string{
		"../docs/outreach/NPS_RANGER_ADOPTION_PLAYBOOK.md",
		"../docs/outreach/NPS_EMAIL_TEMPLATES.md",
		"../docs/outreach/FHWA_WCPP_PARTNERSHIP_BRIEF.md",
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		if len(data) < 500 {
			t.Errorf("file %s is too short: %d bytes", f, len(data))
		}
		content := string(data)
		if !strings.Contains(content, "National Park Service") && !strings.Contains(content, "NPS") {
			t.Errorf("file %s missing NPS reference", f)
		}
	}
}
