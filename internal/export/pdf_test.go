package export_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"corridor/internal/export"
	"corridor/internal/prioritize"

	"github.com/google/uuid"
)

func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Draw a simple green background with a blue line
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 34, G: 139, B: 34, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func sampleCandidateSite() prioritize.CandidateSite {
	return prioritize.CandidateSite{
		ID:                        uuid.MustParse("a1b2c3d4-e5f6-4a1b-8c2d-3e4f5a6b7c8d"),
		Segments:                  []uuid.UUID{uuid.New(), uuid.New()},
		State:                     "WY",
		HighwayRef:                "US-191",
		MilepostStart:             10.0,
		MilepostEnd:               10.5,
		LengthM:                   804.67,
		RecommendedType:           "overpass_with_fencing",
		EstCostUSD:                4977500.0,
		ExpectedAvoidedCollisions: 215.4,
		ExpectedSavingsUSD:        9675000.0,
		BCR:                       1.94,
		P10BCR:                    1.16,
		P90BCR:                    2.91,
		Rationale:                 "Critical migration corridor connecting high-density seasonal ungulate range. High historical collision frequency.",
	}
}

func TestGenerateFundingBriefPDF_ValidStructure(t *testing.T) {
	site := sampleCandidateSite()
	mapPNG := createTestPNG(200, 150)

	pdfBytes, err := export.GenerateFundingBriefPDF(site, mapPNG)
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF failed: %v", err)
	}

	// Verify header (%PDF-)
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("expected PDF header %%PDF-, got %q", string(pdfBytes[:min(len(pdfBytes), 10)]))
	}

	// Verify EOF (%%EOF)
	if !bytes.Contains(pdfBytes, []byte("%%EOF")) {
		t.Errorf("expected PDF to contain %%EOF footer")
	}

	// Verify size > 1KB
	if len(pdfBytes) < 1024 {
		t.Errorf("expected PDF bytes > 1024 bytes, got %d", len(pdfBytes))
	}
}

func TestGenerateFundingBriefPDF_ExpectedContent(t *testing.T) {
	site := sampleCandidateSite()
	pdfBytes, err := export.GenerateFundingBriefPDF(site, nil)
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF with nil map failed: %v", err)
	}

	pdfText := string(pdfBytes)

	expectedSnippets := []string{
		"CORRIDOR WILDLIFE MITIGATION INITIATIVE",
		"FHWA WCPP FUNDING BRIEF",
		"US-191",
		"WY",
		"10.0",
		"10.5",
		"1.94",
		"215.4",
		"9,675,000",
		"Rytwinski et al. 2016",
		"USDOT",
		"disclaimer",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(strings.ToLower(pdfText), strings.ToLower(snippet)) {
			t.Errorf("PDF output does not contain expected snippet %q", snippet)
		}
	}
}

func TestGenerateFundingBriefPDF_WithAndWithoutSnapshot(t *testing.T) {
	site := sampleCandidateSite()

	// 1. With valid PNG snapshot
	pngBytes := createTestPNG(400, 300)
	pdfWithMap, err := export.GenerateFundingBriefPDF(site, pngBytes)
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF with map snapshot failed: %v", err)
	}
	if len(pdfWithMap) < 1024 {
		t.Fatalf("expected PDF with map to be > 1KB, got %d", len(pdfWithMap))
	}

	// 2. With nil snapshot
	pdfNilMap, err := export.GenerateFundingBriefPDF(site, nil)
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF with nil map snapshot failed: %v", err)
	}
	if len(pdfNilMap) < 1024 {
		t.Fatalf("expected PDF with nil map to be > 1KB, got %d", len(pdfNilMap))
	}

	// 3. With empty snapshot
	pdfEmptyMap, err := export.GenerateFundingBriefPDF(site, []byte{})
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF with empty map snapshot failed: %v", err)
	}
	if len(pdfEmptyMap) < 1024 {
		t.Fatalf("expected PDF with empty map to be > 1KB, got %d", len(pdfEmptyMap))
	}

	// 4. With corrupt/invalid image bytes (should gracefully render placeholder)
	pdfCorruptMap, err := export.GenerateFundingBriefPDF(site, []byte("invalid-png-header-data"))
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF with corrupt image should gracefully succeed, got: %v", err)
	}
	if len(pdfCorruptMap) < 1024 {
		t.Fatalf("expected PDF with corrupt map to be > 1KB, got %d", len(pdfCorruptMap))
	}
}

func TestGenerateFundingBriefPDF_EdgeCases(t *testing.T) {
	site := sampleCandidateSite()
	site.Rationale = ""
	site.RecommendedType = "custom_underpass"

	pdfBytes, err := export.GenerateFundingBriefPDF(site, nil)
	if err != nil {
		t.Fatalf("GenerateFundingBriefPDF failed on empty rationale: %v", err)
	}

	pdfText := string(pdfBytes)
	if !strings.Contains(pdfText, "Page 1 of 2") || !strings.Contains(pdfText, "Page 2 of 2") {
		t.Errorf("expected 2-page document footer indicators")
	}
	if !strings.Contains(pdfText, "Custom Underpass") {
		t.Errorf("expected title-cased recommended type in output")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

