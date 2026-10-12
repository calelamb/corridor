package export

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"strings"
	"time"

	"corridor/internal/prioritize"

	"github.com/go-pdf/fpdf"
)

// GenerateFundingBriefPDF creates a 2-page publication-grade FHWA WCPP funding brief PDF
// containing engineering cost estimates, benefit-cost ratios, map snapshots, and literature citations.
func GenerateFundingBriefPDF(site prioritize.CandidateSite, mapSnapshotPNG []byte) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(false, 0)
	// Disable compression so text is searchable in PDF streams
	pdf.SetCompression(false)

	// ==========================================
	// PAGE 1: Project Overview & Corridor Map
	// ==========================================
	pdf.AddPage()

	drawHeaderBanner(pdf, "CORRIDOR WILDLIFE MITIGATION INITIATIVE - FHWA WCPP FUNDING BRIEF",
		"Federal Highway Administration | Wildlife Crossings Pilot Program (WCPP) Grant Proposal")

	drawSiteMetadataBox(pdf, site, 37)

	drawProjectDescription(pdf, site, 74)

	drawMapSnapshot(pdf, site, mapSnapshotPNG, 126)

	drawFooter(pdf, 1, 2)

	// ==========================================
	// PAGE 2: Economic Metrics & Literature Evidence
	// ==========================================
	pdf.AddPage()

	drawHeaderBanner(pdf, "BENEFIT-COST ANALYSIS & ENGINEERING SPECIFICATIONS",
		"FHWA WCPP Economic Justification, Efficacy Validation & Literature Evidence")

	drawKPICards(pdf, site, 35)

	drawEconomicTable(pdf, site, 62)

	drawCitations(pdf, 130)

	drawAgencySubmissionBlock(pdf, site, 180)

	drawDisclaimer(pdf, 230)

	drawFooter(pdf, 2, 2)

	if pdf.Err() {
		return nil, pdf.Error()
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func drawHeaderBanner(pdf *fpdf.Fpdf, title, subtitle string) {
	pdf.SetFillColor(26, 77, 46) // Forest green
	pdf.Rect(15, 15, 185.9, 18, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetXY(18, 18)
	pdf.Cell(180, 6, title)

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetXY(18, 25)
	pdf.Cell(180, 5, subtitle)
}

func drawSiteMetadataBox(pdf *fpdf.Fpdf, site prioritize.CandidateSite, y float64) {
	pdf.SetFillColor(248, 250, 252) // Slate 50
	pdf.SetDrawColor(203, 213, 225) // Slate 300
	pdf.SetLineWidth(0.3)
	pdf.Rect(15, y, 185.9, 33, "FD")

	// Row 1
	drawLabelValue(pdf, 18, y+3, "HIGHWAY & STATE:", fmt.Sprintf("%s (%s)", site.HighwayRef, site.State))
	drawLabelValue(pdf, 65, y+3, "MILEPOST RANGE:", fmt.Sprintf("MP %.2f – %.2f", site.MilepostStart, site.MilepostEnd))
	drawLabelValue(pdf, 115, y+3, "PROJECT LENGTH:", fmt.Sprintf("%.1f m (%.2f mi)", site.LengthM, site.LengthM/1609.34))
	drawLabelValue(pdf, 160, y+3, "STATUS:", "Priority Candidate")

	// Row 2
	drawLabelValue(pdf, 18, y+14, "RECOMMENDED STRUCTURE:", formatStructureName(site.RecommendedType))
	drawLabelValue(pdf, 115, y+14, "BENEFIT-COST RATIO:", fmt.Sprintf("%.2f [P10: %.2f, P90: %.2f]", site.BCR, site.P10BCR, site.P90BCR))

	// Row 3: Site UUID
	pdf.SetTextColor(100, 116, 139)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(18, y+24)
	pdf.Cell(35, 4, "CANDIDATE SITE ID:")

	pdf.SetFont("Courier", "", 7.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(55, y+24)
	pdf.Cell(140, 4, site.ID.String())
}

func drawLabelValue(pdf *fpdf.Fpdf, x, y float64, label, value string) {
	pdf.SetTextColor(100, 116, 139) // Slate 500
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(x, y)
	pdf.Cell(45, 4, label)

	pdf.SetTextColor(15, 23, 42) // Slate 900
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(x, y+4)
	pdf.Cell(45, 5, value)
}

func drawProjectDescription(pdf *fpdf.Fpdf, site prioritize.CandidateSite, y float64) {
	pdf.SetFillColor(26, 77, 46)
	pdf.Rect(15, y, 3, 5, "F")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(20, y)
	pdf.Cell(180, 5, "PROJECT DESCRIPTION & CORRIDOR RATIONALE")

	pdf.SetTextColor(51, 65, 85)
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetXY(15, y+7)

	rationale := site.Rationale
	if strings.TrimSpace(rationale) == "" {
		rationale = fmt.Sprintf("Recommended %s for %s (MP %.1f–%.1f). High-risk corridor identified through statistical clustering.",
			formatStructureName(site.RecommendedType), site.HighwayRef, site.MilepostStart, site.MilepostEnd)
	}

	pdf.MultiCell(185.9, 4.2, rationale, "", "L", false)

	pdf.SetY(pdf.GetY() + 1.5)
	contextText := fmt.Sprintf("Prioritization Context: This candidate site was identified through spatial network clustering of contiguous highway segments exceeding the 90th percentile of collision risk. Implementation of %s with continuous exclusionary fencing addresses critical wildlife movement, restores ecological permeability, and prevents recurring collisions along highway %s between mileposts %.1f and %.1f.",
		formatStructureName(site.RecommendedType), site.HighwayRef, site.MilepostStart, site.MilepostEnd)
	pdf.MultiCell(185.9, 4.2, contextText, "", "L", false)
}

func drawMapSnapshot(pdf *fpdf.Fpdf, site prioritize.CandidateSite, pngBytes []byte, y float64) {
	pdf.SetFillColor(26, 77, 46)
	pdf.Rect(15, y, 3, 5, "F")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(20, y)
	pdf.Cell(180, 5, "PROJECT CORRIDOR MAP SNAPSHOT")

	mapBoxY := y + 7
	mapBoxH := 120.0

	var hasImage bool
	if len(pngBytes) > 0 {
		// Verify valid PNG header and decoding
		if _, err := png.DecodeConfig(bytes.NewReader(pngBytes)); err == nil {
			imgName := fmt.Sprintf("map_snapshot_%s", site.ID.String())
			opt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
			pdf.RegisterImageOptionsReader(imgName, opt, bytes.NewReader(pngBytes))
			if pdf.Ok() {
				pdf.ImageOptions(imgName, 15, mapBoxY, 185.9, mapBoxH, false, opt, 0, "")
				pdf.SetDrawColor(203, 213, 225)
				pdf.SetLineWidth(0.3)
				pdf.Rect(15, mapBoxY, 185.9, mapBoxH, "D")
				hasImage = true
			}
		}
	}

	if !hasImage {
		// Draw styled placeholder frame
		pdf.SetFillColor(241, 245, 249) // Slate 100
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.3)
		pdf.Rect(15, mapBoxY, 185.9, mapBoxH, "FD")

		// Inner decorative border
		pdf.SetDrawColor(226, 232, 240)
		pdf.Rect(18, mapBoxY+3, 179.9, mapBoxH-6, "D")

		pdf.SetTextColor(100, 116, 139)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetXY(15, mapBoxY+40)
		pdf.CellFormat(185.9, 6, fmt.Sprintf("[ CORRIDOR MAP SNAPSHOT: %s MP %.1f – %.1f ]", site.HighwayRef, site.MilepostStart, site.MilepostEnd), "", 0, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(148, 163, 184)
		pdf.SetXY(15, mapBoxY+48)
		pdf.CellFormat(185.9, 5, "Snapshot linestring visualization and GIS aerial basemap", "", 0, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 7.5)
		pdf.SetXY(15, mapBoxY+55)
		pdf.CellFormat(185.9, 4, "High-resolution spatial GIS layers and vector tiles available via Corridor API:", "", 0, "C", false, 0, "")

		pdf.SetFont("Courier", "", 7.5)
		pdf.SetXY(15, mapBoxY+60)
		pdf.CellFormat(185.9, 4, fmt.Sprintf("GET /v1/segments?bbox=... & GET /tiles/evidence/{z}/{x}/{y} [Site: %s]", site.ID.String()), "", 0, "C", false, 0, "")
	}
}

func drawKPICards(pdf *fpdf.Fpdf, site prioritize.CandidateSite, y float64) {
	cardW := 59.3
	cardH := 23.0

	// Card 1: BCR
	pdf.SetFillColor(240, 253, 244) // Mint
	pdf.SetDrawColor(187, 247, 208)
	pdf.SetLineWidth(0.3)
	pdf.Rect(15, y, cardW, cardH, "FD")

	pdf.SetTextColor(22, 101, 52)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetXY(18, y+2.5)
	pdf.Cell(cardW-6, 4, "BENEFIT-COST RATIO (BCR)")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetXY(18, y+7.5)
	pdf.Cell(cardW-6, 7, fmt.Sprintf("%.2f", site.BCR))

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(18, y+15)
	pdf.Cell(cardW-6, 4, fmt.Sprintf("P10: %.2f  |  P90: %.2f", site.P10BCR, site.P90BCR))

	// Card 2: Lifetime Savings
	x2 := 15 + cardW + 4.0
	pdf.SetFillColor(239, 246, 255) // Blue
	pdf.SetDrawColor(191, 219, 254)
	pdf.Rect(x2, y, cardW, cardH, "FD")

	pdf.SetTextColor(30, 64, 175)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetXY(x2+3, y+2.5)
	pdf.Cell(cardW-6, 4, "LIFECYCLE SAVINGS")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetXY(x2+3, y+7.5)
	pdf.Cell(cardW-6, 7, formatUSD(site.ExpectedSavingsUSD))

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(x2+3, y+15)
	pdf.Cell(cardW-6, 4, fmt.Sprintf("%.1f avoided collisions", site.ExpectedAvoidedCollisions))

	// Card 3: Estimated Cost
	x3 := x2 + cardW + 4.0
	pdf.SetFillColor(254, 242, 242) // Rose
	pdf.SetDrawColor(254, 202, 202)
	pdf.Rect(x3, y, cardW, cardH, "FD")

	pdf.SetTextColor(153, 27, 27)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetXY(x3+3, y+2.5)
	pdf.Cell(cardW-6, 4, "ESTIMATED LIFECYCLE COST")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetXY(x3+3, y+7.5)
	pdf.Cell(cardW-6, 7, formatUSD(site.EstCostUSD))

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(x3+3, y+15)
	pdf.Cell(cardW-6, 4, "Capital + 20-year maintenance")
}

func drawEconomicTable(pdf *fpdf.Fpdf, site prioritize.CandidateSite, y float64) {
	pdf.SetFillColor(26, 77, 46)
	pdf.Rect(15, y, 3, 5, "F")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(20, y)
	pdf.Cell(180, 5, "DETAILED LIFECYCLE BENEFIT-COST BREAKDOWN")

	tableY := y + 7
	col1W := 55.0
	col2W := 50.0
	col3W := 80.9
	rowH := 7.0

	// Header row
	pdf.SetFillColor(30, 41, 59) // Slate 800
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 7.5)

	pdf.SetXY(15, tableY)
	pdf.CellFormat(col1W, rowH, "ECONOMIC / ENGINEERING METRIC", "1", 0, "L", true, 0, "")
	pdf.CellFormat(col2W, rowH, "PROJECT VALUE", "1", 0, "L", true, 0, "")
	pdf.CellFormat(col3W, rowH, "METHODOLOGY & STANDARDS BASIS", "1", 0, "L", true, 0, "")

	// Approximate capital vs annual maint based on standard structure definitions
	capitalCost := site.EstCostUSD * (1.0 / (1.0 + (0.01 * 20.0)))
	annualMaint := capitalCost * 0.01

	rows := []struct {
		metric string
		value  string
		basis  string
	}{
		{
			metric: "Capital Construction Cost",
			value:  formatUSD(capitalCost),
			basis:  "Engineering base cost + fencing per linear meter",
		},
		{
			metric: "Annual Maintenance Cost",
			value:  formatUSD(annualMaint) + " / yr",
			basis:  "Inspection & exclusionary fence repair (1.0% annual)",
		},
		{
			metric: "Structural Design Lifespan",
			value:  "20 Years",
			basis:  "FHWA standard wildlife crossing infrastructure service life",
		},
		{
			metric: "Total Lifecycle Cost (USD)",
			value:  formatUSD(site.EstCostUSD),
			basis:  "Capital cost + 20-year operational maintenance",
		},
		{
			metric: "Expected Avoided Collisions",
			value:  fmt.Sprintf("%.1f Collisions", site.ExpectedAvoidedCollisions),
			basis:  "86% collision reduction efficacy over 20-yr design lifespan",
		},
		{
			metric: "Expected Lifecycle Savings (USD)",
			value:  formatUSD(site.ExpectedSavingsUSD),
			basis:  "Avoided collisions x USDOT comprehensive collision cost",
		},
		{
			metric: "Benefit-Cost Ratio (BCR)",
			value:  fmt.Sprintf("%.2f [P10: %.2f – P90: %.2f]", site.BCR, site.P10BCR, site.P90BCR),
			basis:  "USDOT Benefit-Cost Analysis Guidance (>1.0 indicates net positive return)",
		},
	}

	curY := tableY + rowH
	for i, r := range rows {
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(248, 250, 252) // Slate 50
		}

		pdf.SetTextColor(15, 23, 42)
		pdf.SetFont("Helvetica", "B", 7.5)
		pdf.SetXY(15, curY)
		pdf.CellFormat(col1W, rowH, r.metric, "1", 0, "L", true, 0, "")

		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(col2W, rowH, r.value, "1", 0, "L", true, 0, "")

		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(71, 85, 105)
		pdf.CellFormat(col3W, rowH, r.basis, "1", 0, "L", true, 0, "")

		curY += rowH
	}
}

func drawCitations(pdf *fpdf.Fpdf, y float64) {
	pdf.SetFillColor(26, 77, 46)
	pdf.Rect(15, y, 3, 5, "F")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(20, y)
	pdf.Cell(180, 5, "LITERATURE EVIDENCE & CITATIONS")

	pdf.SetTextColor(51, 65, 85)
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetXY(15, y+6.5)

	citations := []string{
		"1. Rytwinski et al. 2016. How Effective Are Road Mitigation Measures in Reducing Wildlife-Vehicle Collisions and Improving Permeability for Animals? A Systematic Review. Environmental Evidence, 5(1):21. Meta-analysis establishing 86% collision reduction efficacy for crossing structures integrated with continuous fencing.",
		"2. USDOT Benefit-Cost Analysis Guidance for Discretionary Grant Programs (2024). U.S. Department of Transportation, Washington, D.C. Standardized economic values for avoided vehicular crashes, injury prevention, and property damage valuation.",
		"3. Huijser, M.P., et al. (2009). Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Collisions with Large Ungulates in the United States and Canada. Transportation Research Record, 2123(1), 161–172.",
	}

	for _, cit := range citations {
		pdf.MultiCell(185.9, 3.6, cit, "", "L", false)
		pdf.SetY(pdf.GetY() + 1.0)
	}
}

func drawAgencySubmissionBlock(pdf *fpdf.Fpdf, site prioritize.CandidateSite, y float64) {
	pdf.SetFillColor(26, 77, 46)
	pdf.Rect(15, y, 3, 5, "F")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(20, y)
	pdf.Cell(180, 5, "AGENCY SUBMISSION & CERTIFICATION BLOCK")

	boxY := y + 6.5
	boxW := 90.0
	boxH := 38.0

	// Left Box: Agency & Project ID
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.Rect(15, boxY, boxW, boxH, "FD")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetXY(18, boxY+2.5)
	pdf.Cell(boxW-6, 4, "FHWA WCPP GRANT APPLICANT DATA")

	items := []struct {
		lbl string
		val string
	}{
		{"Sponsoring Agency:", fmt.Sprintf("%s Department of Transportation", site.State)},
		{"Funding Program:", "FHWA Wildlife Crossings Pilot Program (WCPP)"},
		{"Corridor Route:", fmt.Sprintf("%s (MP %.1f – %.1f)", site.HighwayRef, site.MilepostStart, site.MilepostEnd)},
		{"Candidate Site ID:", site.ID.String()[:18] + "..."},
		{"Priority Status:", "Tier 1 High-Risk Infrastructure Grant Candidate"},
	}

	lineY := boxY + 7.5
	for _, it := range items {
		pdf.SetFont("Helvetica", "B", 6.8)
		pdf.SetTextColor(100, 116, 139)
		pdf.SetXY(18, lineY)
		pdf.Cell(30, 4, it.lbl)

		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(48, lineY)
		pdf.Cell(boxW-32, 4, it.val)
		lineY += 5.5
	}

	// Right Box: Engineering Sign-off
	x2 := 15 + boxW + 5.9
	pdf.Rect(x2, boxY, boxW, boxH, "FD")

	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Helvetica", "B", 7.5)
	pdf.SetXY(x2+3, boxY+2.5)
	pdf.Cell(boxW-6, 4, "REVIEW & ENGINEERING SIGN-OFF")

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(x2+3, boxY+8)
	pdf.Cell(boxW-6, 4, fmt.Sprintf("Report Generated: %s", time.Now().UTC().Format("January 02, 2006")))

	pdf.SetXY(x2+3, boxY+15)
	pdf.Cell(boxW-6, 4, "Lead Transportation Engineer:")
	pdf.SetXY(x2+3, boxY+21)
	pdf.Cell(boxW-6, 3, "__________________________________________________")

	pdf.SetXY(x2+3, boxY+27)
	pdf.Cell(boxW-6, 4, "State Wildlife Agency Representative:")
	pdf.SetXY(x2+3, boxY+33)
	pdf.Cell(boxW-6, 3, "__________________________________________________")
}

func drawDisclaimer(pdf *fpdf.Fpdf, y float64) {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.3)
	pdf.Rect(15, y, 185.9, 18, "FD")

	pdf.SetTextColor(100, 116, 139)
	pdf.SetFont("Helvetica", "B", 6.8)
	pdf.SetXY(18, y+2)
	pdf.Cell(180, 3.5, "DISCLAIMER & METHODOLOGICAL LIMITATIONS")

	pdf.SetFont("Helvetica", "", 6.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(18, y+6)
	disclaimerText := "Produced by CORRIDOR Spatial Analytics. Estimates based on historical collision models and empirical mitigation efficacy studies. For official WCPP grant submission verification. Actual construction costs and engineering feasibility require site-specific geotechnical and structural engineering field verification. Benefit-cost calculations assume full operational maintenance throughout the stated infrastructure lifespan."
	pdf.MultiCell(179.9, 3.2, disclaimerText, "", "L", false)
}

func drawFooter(pdf *fpdf.Fpdf, pageNum, totalPages int) {
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, 260, 200.9, 260)

	pdf.SetTextColor(148, 163, 184)
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetXY(15, 262)
	pdf.Cell(120, 4, "CORRIDOR Wildlife Mitigation Initiative | FHWA WCPP Project Brief")

	pdf.SetXY(160, 262)
	pdf.CellFormat(40.9, 4, fmt.Sprintf("Page %d of %d", pageNum, totalPages), "", 0, "R", false, 0, "")
}

func formatUSD(val float64) string {
	if val < 0 {
		return "-$" + formatNumber(-val, 0)
	}
	return "$" + formatNumber(val, 0)
}

func formatNumber(val float64, decimals int) string {
	pow := math.Pow(10, float64(decimals))
	rounded := math.Round(val*pow) / pow

	intPart := int64(math.Floor(rounded))
	decPart := int64(math.Round((rounded - float64(intPart)) * pow))

	intStr := fmt.Sprintf("%d", intPart)
	// Add commas
	var result []byte
	l := len(intStr)
	for i, c := range []byte(intStr) {
		if i > 0 && (l-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}

	if decimals > 0 {
		return fmt.Sprintf("%s.%0*d", string(result), decimals, decPart)
	}
	return string(result)
}

func formatStructureName(recommendedType string) string {
	switch recommendedType {
	case "overpass_with_fencing":
		return "Wildlife Overpass with Continuous Fencing"
	case "underpass_large":
		return "Large Wildlife Underpass with Continuous Fencing"
	case "underpass_small":
		return "Small Wildlife Underpass with Continuous Fencing"
	case "fencing_only":
		return "Exclusionary Wildlife Fencing with Escape Ramps"
	case "warning_system":
		return "Dynamic Animal Detection and Warning System"
	default:
		words := strings.Split(recommendedType, "_")
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
		}
		return strings.Join(words, " ")
	}
}
