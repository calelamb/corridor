// tests/contact_directory_test.go
package tests

import (
	"os"
	"strings"
	"testing"
)

func TestNPSContactDirectory(t *testing.T) {
	candidates := []string{
		"../docs/outreach/NPS_CONTACT_DIRECTORY.md",
		"docs/outreach/NPS_CONTACT_DIRECTORY.md",
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
		t.Fatalf("failed to read NPS contact directory from %v: %v", candidates, readErr)
	}

	if len(data) <= 1000 {
		t.Errorf("file %s is too short: expected > 1000 bytes, got %d bytes", foundPath, len(data))
	}

	content := string(data)

	// 1. Verify 8 target park units
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
			t.Errorf("missing target park unit: %s", park)
		}
	}

	// 2. Verify divisions and required sections
	requiredSections := []string{
		"Division of Resource Management & Science",
		"Division of Visitor and Resource Protection",
		"Dispatch",
		"State DOT Wildlife Crossings Liaisons",
	}
	for _, sec := range requiredSections {
		if !strings.Contains(content, sec) {
			t.Errorf("missing required section/division: %s", sec)
		}
	}

	// 3. Verify all 7 State DOT Wildlife Crossings Liaisons
	requiredDOTs := []string{
		"WYDOT",
		"MDT",
		"CDOT",
		"Caltrans",
		"ADOT",
		"WSDOT",
		"VDOT",
	}
	for _, dot := range requiredDOTs {
		if !strings.Contains(content, dot) {
			t.Errorf("missing state DOT liaison: %s", dot)
		}
	}

	// 4. Verify specific role profiles
	requiredRoles := []string{
		"Chief of Resource Management",
		"Supervisory Wildlife Biologist",
		"Chief Ranger",
		"Facility & Transportation Manager",
	}
	for _, role := range requiredRoles {
		if !strings.Contains(content, role) {
			t.Errorf("missing role profile: %s", role)
		}
	}

	// 5. Verify FHWA and WCPP
	requiredPrograms := []string{
		"FHWA",
		"Wildlife Crossings Pilot Program",
		"WCPP",
	}
	for _, prog := range requiredPrograms {
		if !strings.Contains(content, prog) {
			t.Errorf("missing program reference: %s", prog)
		}
	}

	// 6. Verify non-placeholder email guidelines and format
	emailGuidelines := []string{
		"@nps.gov",
		"first_last@nps.gov",
	}
	for _, eg := range emailGuidelines {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(eg)) {
			t.Errorf("missing email formatting guidance: %s", eg)
		}
	}

	// 7. Verify outbound campaign tracking checklist
	if !strings.Contains(content, "Outbound Campaign Tracking") {
		t.Errorf("missing 'Outbound Campaign Tracking' section")
	}
}
