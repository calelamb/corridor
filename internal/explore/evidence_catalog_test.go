package explore

import (
	"strings"
	"testing"
)

func TestGetAcademicEvidenceCatalog(t *testing.T) {
	catalog := GetAcademicEvidenceCatalog()

	if len(catalog) < 15 {
		t.Fatalf("expected at least 15 peer-reviewed papers, got %d", len(catalog))
	}

	categoriesFound := make(map[string]int)
	foundRytwinski := false
	foundHuijser := false
	foundClevenger := false
	foundKauffman := false

	for i, paper := range catalog {
		if strings.TrimSpace(paper.ID) == "" {
			t.Errorf("paper [%d] has empty ID", i)
		}
		if strings.TrimSpace(paper.Title) == "" {
			t.Errorf("paper [%d] has empty Title", i)
		}
		if strings.TrimSpace(paper.Authors) == "" {
			t.Errorf("paper [%d] has empty Authors", i)
		}
		if paper.Year < 1995 || paper.Year > 2026 {
			t.Errorf("paper [%s] has invalid Year: %d", paper.ID, paper.Year)
		}
		if strings.TrimSpace(paper.Journal) == "" {
			t.Errorf("paper [%s] has empty Journal", paper.ID)
		}
		if !strings.HasPrefix(paper.DOI, "10.") && !strings.HasPrefix(paper.DOI, "https://doi.org/") {
			t.Errorf("paper [%s] has invalid DOI: %s", paper.ID, paper.DOI)
		}
		if strings.TrimSpace(paper.Category) == "" {
			t.Errorf("paper [%s] has empty Category", paper.ID)
		}
		if strings.TrimSpace(paper.KeyFindings) == "" {
			t.Errorf("paper [%s] has empty KeyFindings", paper.ID)
		}
		if strings.TrimSpace(paper.AppliedInCorridor) == "" {
			t.Errorf("paper [%s] has empty AppliedInCorridor rationale", paper.ID)
		}

		categoriesFound[paper.Category]++

		if strings.Contains(paper.Authors, "Rytwinski") {
			foundRytwinski = true
		}
		if strings.Contains(paper.Authors, "Huijser") {
			foundHuijser = true
		}
		if strings.Contains(paper.Authors, "Clevenger") {
			foundClevenger = true
		}
		if strings.Contains(paper.Authors, "Kauffman") {
			foundKauffman = true
		}
	}

	requiredCategories := []string{"mitigation", "bias", "ecology", "economics"}
	for _, req := range requiredCategories {
		if categoriesFound[req] == 0 {
			t.Errorf("expected at least 1 paper in category %q, got 0", req)
		}
	}

	if !foundRytwinski {
		t.Errorf("missing foundational paper by Rytwinski et al.")
	}
	if !foundHuijser {
		t.Errorf("missing foundational paper by Huijser et al.")
	}
	if !foundClevenger {
		t.Errorf("missing foundational paper by Clevenger et al.")
	}
	if !foundKauffman {
		t.Errorf("missing foundational paper by Kauffman et al.")
	}
}
