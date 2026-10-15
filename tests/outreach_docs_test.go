// tests/outreach_docs_test.go
package tests

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func readDocFile(t *testing.T, filename string) string {
	candidates := []string{
		"../docs/outreach/" + filename,
		"docs/outreach/" + filename,
	}
	var data []byte
	var err error
	for _, p := range candidates {
		data, err = os.ReadFile(p)
		if err == nil {
			return string(data)
		}
	}
	t.Fatalf("failed to read %s from %v: %v", filename, candidates, err)
	return ""
}

func TestOutreachDocumentation(t *testing.T) {
	files := []string{
		"NPS_RANGER_ADOPTION_PLAYBOOK.md",
		"NPS_EMAIL_TEMPLATES.md",
		"FHWA_WCPP_PARTNERSHIP_BRIEF.md",
		"NPS_CONTACT_DIRECTORY.md",
		"NPS_RANGER_QUICKSTART_GUIDE.md",
	}

	for _, f := range files {
		content := readDocFile(t, f)
		if len(content) < 1000 {
			t.Errorf("file %s is too short: %d bytes (expected > 1000)", f, len(content))
		}
		if !strings.Contains(content, "National Park Service") && !strings.Contains(content, "NPS") {
			t.Errorf("file %s missing NPS reference", f)
		}
		if !strings.Contains(content, "Apache-2.0") {
			t.Errorf("file %s missing Apache-2.0 license reference", f)
		}
	}
}

func TestAdoptionPlaybookEightTargetParks(t *testing.T) {
	content := readDocFile(t, "NPS_RANGER_ADOPTION_PLAYBOOK.md")

	targetParks := []string{
		"Yellowstone",
		"Grand Teton",
		"Glacier",
		"Rocky Mountain",
		"Yosemite",
		"Grand Canyon",
		"Olympic",
		"Shenandoah",
	}

	for _, park := range targetParks {
		if !strings.Contains(content, park) {
			t.Errorf("NPS_RANGER_ADOPTION_PLAYBOOK.md missing target park: %s", park)
		}
	}

	if !strings.Contains(content, "eight flagship units") {
		t.Errorf("NPS_RANGER_ADOPTION_PLAYBOOK.md should specify 'eight flagship units'")
	}
}

func TestStatutoryCitationsAndValuations(t *testing.T) {
	files := []string{
		"FHWA_WCPP_PARTNERSHIP_BRIEF.md",
		"NPS_RANGER_ADOPTION_PLAYBOOK.md",
		"NPS_EMAIL_TEMPLATES.md",
		"NPS_RANGER_QUICKSTART_GUIDE.md",
	}

	for _, f := range files {
		content := readDocFile(t, f)
		if !strings.Contains(content, "23 U.S.C. § 171") {
			t.Errorf("file %s missing statutory citation '23 U.S.C. § 171'", f)
		}
		if !strings.Contains(content, "$14,000") {
			t.Errorf("file %s missing USDOT valuation $14,000 for deer", f)
		}
		if !strings.Contains(content, "$45,000") {
			t.Errorf("file %s missing USDOT valuation $45,000 for elk", f)
		}
		if !strings.Contains(content, "$83,000") {
			t.Errorf("file %s missing USDOT valuation $83,000 for moose", f)
		}
	}
}

func TestEmailTemplateAnchors(t *testing.T) {
	content := readDocFile(t, "NPS_EMAIL_TEMPLATES.md")

	// Match markdown links of the form [text](#anchor)
	reLink := regexp.MustCompile(`\[([^\]]+)\]\(#([^\)]+)\)`)
	matches := reLink.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		t.Fatalf("no anchor links found in NPS_EMAIL_TEMPLATES.md")
	}

	for _, match := range matches {
		anchorID := match[2]
		// Verify either an HTML anchor <a id="anchorID"> or matching slug exists
		htmlAnchor := `id="` + anchorID + `"`
		if !strings.Contains(content, htmlAnchor) {
			t.Errorf("anchor link #%s in NPS_EMAIL_TEMPLATES.md has no matching id in file", anchorID)
		}
	}
}

