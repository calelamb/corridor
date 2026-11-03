package prioritize_test

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corridor/internal/prioritize"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func TestBCR_OverpassFencing_Calculations(t *testing.T) {
	// Structure: overpass_with_fencing
	// Literature effectiveness: 86% (0.86), Rytwinski et al. 2016
	// Lifespan: 25 years
	// Base cost: $3,500,000
	// Fencing cost per meter: $120/m
	// Annual maintenance: 1.5% (0.015)
	structure := prioritize.MitigationStructure{
		ID:                   "overpass_with_fencing",
		Name:                 "Wildlife Overpass with Continuous Fencing",
		Effectiveness:        0.86,
		BaseCostUSD:          3500000.0,
		FencingCostPerMeter:  120.0,
		AnnualMaintenancePct: 0.015,
		LifespanYears:        25,
		Citation:             "Rytwinski et al. 2016",
	}

	lengthM := 1000.0
	annualRate := 10.0
	p10Rate := 6.0
	p90Rate := 15.0

	// Test with Deer: $14,000 / collision
	deerCost := 14000.0
	res := prioritize.CalculateBCR(annualRate, p10Rate, p90Rate, lengthM, structure, deerCost)

	// Capital cost = 3,500,000 + (120 * 1000) = 3,620,000
	expectedCapital := 3620000.0
	if math.Abs(res.CapitalCostUSD-expectedCapital) > 1e-2 {
		t.Fatalf("expected capital cost %.2f, got %.2f", expectedCapital, res.CapitalCostUSD)
	}

	// Annual maintenance = 3,620,000 * 0.015 = 54,300
	// Lifetime cost = 3,620,000 + (54,300 * 25) = 4,977,500
	expectedLifetimeCost := 4977500.0
	if math.Abs(res.LifetimeCostUSD-expectedLifetimeCost) > 1e-2 {
		t.Fatalf("expected lifetime cost %.2f, got %.2f", expectedLifetimeCost, res.LifetimeCostUSD)
	}

	// Avoided collisions = 10.0 * 0.86 * 25 = 215.0
	expectedAvoided := 215.0
	if math.Abs(res.ExpectedAvoidedCollisions-expectedAvoided) > 1e-4 {
		t.Fatalf("expected avoided collisions %.4f, got %.4f", expectedAvoided, res.ExpectedAvoidedCollisions)
	}

	// Expected savings = 215.0 * 14,000 = 3,010,000
	expectedSavings := 3010000.0
	if math.Abs(res.ExpectedSavingsUSD-expectedSavings) > 1e-2 {
		t.Fatalf("expected savings %.2f, got %.2f", expectedSavings, res.ExpectedSavingsUSD)
	}

	// BCR = 3,010,000 / 4,977,500 = 0.604721...
	expectedBCR := expectedSavings / expectedLifetimeCost
	if math.Abs(res.BCR-expectedBCR) > 1e-4 {
		t.Fatalf("expected BCR %.4f, got %.4f", expectedBCR, res.BCR)
	}

	// P10 Avoided = 6.0 * 0.86 * 25 = 129.0
	// P10 Savings = 129.0 * 14,000 = 1,806,000
	// P10 BCR = 1,806,000 / 4,977,500 = 0.362832...
	expectedP10BCR := (6.0 * 0.86 * 25.0 * 14000.0) / expectedLifetimeCost
	if math.Abs(res.P10BCR-expectedP10BCR) > 1e-4 {
		t.Fatalf("expected P10 BCR %.4f, got %.4f", expectedP10BCR, res.P10BCR)
	}

	// P90 Avoided = 15.0 * 0.86 * 25 = 322.5
	// P90 Savings = 322.5 * 14,000 = 4,515,000
	// P90 BCR = 4,515,000 / 4,977,500 = 0.907081...
	expectedP90BCR := (15.0 * 0.86 * 25.0 * 14000.0) / expectedLifetimeCost
	if math.Abs(res.P90BCR-expectedP90BCR) > 1e-4 {
		t.Fatalf("expected P90 BCR %.4f, got %.4f", expectedP90BCR, res.P90BCR)
	}

	// Test Elk: $45,000 / collision -> BCR > 1.0
	elkRes := prioritize.CalculateBCR(annualRate, p10Rate, p90Rate, lengthM, structure, 45000.0)
	expectedElkBCR := (215.0 * 45000.0) / expectedLifetimeCost // 9,675,000 / 4,977,500 = 1.9437
	if math.Abs(elkRes.BCR-expectedElkBCR) > 1e-4 {
		t.Fatalf("expected Elk BCR %.4f, got %.4f", expectedElkBCR, elkRes.BCR)
	}

	// Test Moose: $83,000 / collision -> BCR > 3.5
	mooseRes := prioritize.CalculateBCR(annualRate, p10Rate, p90Rate, lengthM, structure, 83000.0)
	expectedMooseBCR := (215.0 * 83000.0) / expectedLifetimeCost // 17,845,000 / 4,977,500 = 3.585
	if math.Abs(mooseRes.BCR-expectedMooseBCR) > 1e-4 {
		t.Fatalf("expected Moose BCR %.4f, got %.4f", expectedMooseBCR, mooseRes.BCR)
	}
}

func TestClusterAdjacentHighRiskSegments(t *testing.T) {
	// Create a series of segments along US-191:
	// Seg 1: MP 10.0-10.1, Percentile 92 (High)
	// Seg 2: MP 10.1-10.2, Percentile 95 (High)
	// Seg 3: MP 10.2-10.3, Percentile 91 (High)
	// Seg 4: MP 10.3-10.4, Percentile 80 (Low - gap)
	// Seg 5: MP 10.4-10.5, Percentile 96 (High - isolated)
	// Seg 6: MP 10.5-10.6, Percentile 50 (Low)

	cfg := prioritize.DefaultConfig()

	seg1ID := uuid.New()
	seg2ID := uuid.New()
	seg3ID := uuid.New()
	seg4ID := uuid.New()
	seg5ID := uuid.New()
	seg6ID := uuid.New()

	segments := []prioritize.SegmentRecord{
		{
			ID:             seg1ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.0,
			MilepostEnd:    10.1,
			LengthM:        160.9,
			RiskScore:      2.0,
			PercentileRank: 92.0,
			P10:            1.2,
			P90:            3.1,
			SpeciesBreakdown: map[string]float64{
				"deer": 1.0,
			},
		},
		{
			ID:             seg2ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.1,
			MilepostEnd:    10.2,
			LengthM:        160.9,
			RiskScore:      3.5,
			PercentileRank: 95.0,
			P10:            2.0,
			P90:            5.2,
			SpeciesBreakdown: map[string]float64{
				"deer": 0.5,
				"elk":  0.5,
			},
		},
		{
			ID:             seg3ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.2,
			MilepostEnd:    10.3,
			LengthM:        160.9,
			RiskScore:      1.8,
			PercentileRank: 91.0,
			P10:            1.0,
			P90:            2.8,
			SpeciesBreakdown: map[string]float64{
				"deer": 1.0,
			},
		},
		{
			ID:             seg4ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.3,
			MilepostEnd:    10.4,
			LengthM:        160.9,
			RiskScore:      0.4,
			PercentileRank: 80.0, // NOT > 90th percentile
			P10:            0.2,
			P90:            0.8,
		},
		{
			ID:             seg5ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.4,
			MilepostEnd:    10.5,
			LengthM:        160.9,
			RiskScore:      2.5,
			PercentileRank: 96.0, // High isolated segment
			P10:            1.5,
			P90:            3.8,
			SpeciesBreakdown: map[string]float64{
				"moose": 1.0,
			},
		},
		{
			ID:             seg6ID,
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  10.5,
			MilepostEnd:    10.6,
			LengthM:        160.9,
			RiskScore:      0.1,
			PercentileRank: 50.0,
			P10:            0.05,
			P90:            0.2,
		},
	}

	candidates := prioritize.GenerateCandidatesFromSegments(segments, 10000000.0, cfg)

	// Should produce exactly 2 candidate sites:
	// 1. Cluster of Seg 1, 2, 3 (continuous high-risk stretch MP 10.0-10.3)
	// 2. Cluster of Seg 5 (single isolated high-risk stretch MP 10.4-10.5)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidate sites, got %d", len(candidates))
	}

	// Verify the 3-segment cluster
	var cluster3 *prioritize.CandidateSite
	for i := range candidates {
		if len(candidates[i].Segments) == 3 {
			cluster3 = &candidates[i]
			break
		}
	}
	if cluster3 == nil {
		t.Fatalf("expected cluster with 3 segments")
	}

	if cluster3.HighwayRef != "US-191" || cluster3.State != "WY" {
		t.Fatalf("unexpected highway or state: %s, %s", cluster3.HighwayRef, cluster3.State)
	}
	if math.Abs(cluster3.MilepostStart-10.0) > 1e-4 || math.Abs(cluster3.MilepostEnd-10.3) > 1e-4 {
		t.Fatalf("unexpected mileposts: %.2f - %.2f", cluster3.MilepostStart, cluster3.MilepostEnd)
	}
	expectedLen := 160.9 * 3
	if math.Abs(cluster3.LengthM-expectedLen) > 1e-2 {
		t.Fatalf("expected length %.2f, got %.2f", expectedLen, cluster3.LengthM)
	}
	if cluster3.ExpectedAvoidedCollisions <= 0 {
		t.Fatalf("expected positive avoided collisions, got %.2f", cluster3.ExpectedAvoidedCollisions)
	}
	if cluster3.ExpectedSavingsUSD <= 0 {
		t.Fatalf("expected positive savings, got %.2f", cluster3.ExpectedSavingsUSD)
	}
	if cluster3.BCR <= 0 || cluster3.P10BCR <= 0 || cluster3.P90BCR <= 0 {
		t.Fatalf("expected positive BCRs, got BCR=%.2f, P10=%.2f, P90=%.2f", cluster3.BCR, cluster3.P10BCR, cluster3.P90BCR)
	}
	if cluster3.P10BCR >= cluster3.BCR || cluster3.BCR >= cluster3.P90BCR {
		t.Fatalf("expected P10 < BCR < P90, got P10=%.2f, BCR=%.2f, P90=%.2f", cluster3.P10BCR, cluster3.BCR, cluster3.P90BCR)
	}
	if cluster3.Rationale == "" {
		t.Fatalf("expected non-empty rationale")
	}
}

func TestHighwayFilterBeforeKnapsack(t *testing.T) {
	cfg := prioritize.DefaultConfig()
	expensive := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-191",
		State:          "WY",
		MilepostStart:  12.0,
		MilepostEnd:    12.2,
		LengthM:        200.0,
		RiskScore:      8.0,
		PercentileRank: 99.0,
		P10:            5.0,
		P90:            11.0,
		SpeciesBreakdown: map[string]float64{
			"moose": 1.0,
		},
	}
	parkHighway := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-89",
		State:          "WY",
		MilepostStart:  18.0,
		MilepostEnd:    18.2,
		LengthM:        200.0,
		RiskScore:      6.0,
		PercentileRank: 98.0,
		P10:            4.0,
		P90:            8.0,
		SpeciesBreakdown: map[string]float64{
			"elk": 1.0,
		},
	}
	segments := []prioritize.SegmentRecord{expensive, parkHighway}
	open := prioritize.GenerateCandidatesFromSegmentsWithOptions(segments, prioritize.PrioritizeOptions{
		Budget:        5000000,
		MinPercentile: 90,
	}, cfg)
	if len(open) == 0 || open[0].HighwayRef != "US-191" {
		t.Fatalf("statewide budget should select the higher-BCR US-191 project, got %+v", open)
	}
	park := prioritize.GenerateCandidatesFromSegmentsWithOptions(segments, prioritize.PrioritizeOptions{
		Budget:        5000000,
		Highway:       "US-89",
		MinPercentile: 90,
	}, cfg)
	if len(park) != 1 || park[0].HighwayRef != "US-89" {
		t.Fatalf("park highway filter ran after the knapsack, got %+v", park)
	}
}

func TestBudgetOptimization_KnapsackGreedy(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	// Create 3 candidate clusters with known costs and BCRs:
	// Project A: High BCR, Cost ~$200k
	// Project B: Medium BCR, Cost ~$300k
	// Project C: High Cost ~$5M, High BCR
	segA := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-89",
		State:          "WY",
		MilepostStart:  1.0,
		MilepostEnd:    1.1,
		LengthM:        160.0,
		RiskScore:      5.0,
		PercentileRank: 98.0,
		P10:            3.0,
		P90:            8.0,
		SpeciesBreakdown: map[string]float64{
			"moose": 1.0, // High savings
		},
	}
	segB := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-191",
		State:          "WY",
		MilepostStart:  20.0,
		MilepostEnd:    20.1,
		LengthM:        160.0,
		RiskScore:      1.5,
		PercentileRank: 93.0,
		P10:            0.8,
		P90:            2.5,
		SpeciesBreakdown: map[string]float64{
			"deer": 1.0,
		},
	}

	// Test with zero budget: should return empty slice
	zeroBudget := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segA, segB}, 0.0, cfg)
	if len(zeroBudget) != 0 {
		t.Fatalf("expected 0 candidates for 0 budget, got %d", len(zeroBudget))
	}

	// Test with ample budget: should return all candidate sites sorted by BCR descending
	allCandidates := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segA, segB}, 100000000.0, cfg)
	if len(allCandidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(allCandidates))
	}
	if allCandidates[0].BCR < allCandidates[1].BCR {
		t.Fatalf("candidates should be sorted by BCR descending: %.2f vs %.2f", allCandidates[0].BCR, allCandidates[1].BCR)
	}

	// Test with budget that only allows the top 1 candidate
	topCost := allCandidates[0].EstCostUSD
	budgetOnlyOne := topCost + 100.0
	// If second candidate cost exceeds remaining budget:
	if topCost+allCandidates[1].EstCostUSD > budgetOnlyOne {
		filtered := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segA, segB}, budgetOnlyOne, cfg)
		if len(filtered) != 1 {
			t.Fatalf("expected exactly 1 candidate within budget, got %d", len(filtered))
		}
		if filtered[0].ID != allCandidates[0].ID {
			t.Fatalf("expected top BCR candidate to be selected")
		}
		totalCost := 0.0
		for _, c := range filtered {
			totalCost += c.EstCostUSD
		}
		if totalCost > budgetOnlyOne {
			t.Fatalf("total cost %.2f exceeds budget %.2f", totalCost, budgetOnlyOne)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	// 1. Empty segments list
	empty := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{}, 1000000.0, cfg)
	if len(empty) != 0 {
		t.Fatalf("expected empty result for empty input")
	}

	// 2. No high-risk segments (all < 90th percentile)
	lowRisk := []prioritize.SegmentRecord{
		{
			ID:             uuid.New(),
			HighwayRef:     "I-80",
			State:          "WY",
			MilepostStart:  50.0,
			MilepostEnd:    50.1,
			LengthM:        160.0,
			RiskScore:      0.2,
			PercentileRank: 65.0,
		},
	}
	resLow := prioritize.GenerateCandidatesFromSegments(lowRisk, 1000000.0, cfg)
	if len(resLow) != 0 {
		t.Fatalf("expected 0 candidates when no high-risk segments, got %d", len(resLow))
	}

	// 3. Single segment project
	singleHigh := []prioritize.SegmentRecord{
		{
			ID:             uuid.New(),
			HighwayRef:     "I-80",
			State:          "WY",
			MilepostStart:  50.0,
			MilepostEnd:    50.1,
			LengthM:        160.0,
			RiskScore:      4.0,
			PercentileRank: 94.0,
			P10:            2.5,
			P90:            6.2,
			SpeciesBreakdown: map[string]float64{
				"elk": 1.0,
			},
		},
	}
	resSingle := prioritize.GenerateCandidatesFromSegments(singleHigh, 5000000.0, cfg)
	if len(resSingle) != 1 {
		t.Fatalf("expected 1 candidate for single segment, got %d", len(resSingle))
	}
	if len(resSingle[0].Segments) != 1 || resSingle[0].Segments[0] != singleHigh[0].ID {
		t.Fatalf("unexpected segment ID in candidate site")
	}

	// 4. GenerateCandidates with nil pool should return error
	_, err := prioritize.GenerateCandidates(context.Background(), nil, 1000000.0)
	if err == nil {
		t.Fatalf("expected error when pool is nil")
	}
}

func TestLoadConfigurationFiles(t *testing.T) {
	// Verify loading from config/ files
	effPath := filepath.Join("..", "..", "config", "mitigation_effectiveness.yaml")
	speciesPath := filepath.Join("..", "..", "config", "species_costs.yaml")

	cfg, err := prioritize.LoadConfigFromFiles(effPath, speciesPath)
	if err != nil {
		t.Fatalf("failed to load configs: %v", err)
	}

	// Check structures
	overpass, ok := cfg.Structures["overpass_with_fencing"]
	if !ok {
		t.Fatalf("missing overpass_with_fencing")
	}
	if math.Abs(overpass.Effectiveness-0.86) > 1e-4 {
		t.Fatalf("expected overpass effectiveness 0.86, got %.2f", overpass.Effectiveness)
	}
	if overpass.Citation == "" {
		t.Fatalf("expected academic citation for overpass")
	}

	retro, ok := cfg.Structures["small_carnivore_culvert_retrofit"]
	if !ok {
		t.Fatalf("missing small_carnivore_culvert_retrofit")
	}
	if math.Abs(retro.Effectiveness-0.70) > 1e-4 {
		t.Fatalf("expected retrofit effectiveness 0.70, got %.2f", retro.Effectiveness)
	}
	if math.Abs(retro.BaseCostUSD-45000.0) > 1e-2 {
		t.Fatalf("expected retrofit base cost 45000, got %.2f", retro.BaseCostUSD)
	}
	if math.Abs(retro.FencingCostPerMeter-40.0) > 1e-2 {
		t.Fatalf("expected retrofit fencing cost 40, got %.2f", retro.FencingCostPerMeter)
	}
	if math.Abs(retro.AnnualMaintenancePct-0.02) > 1e-4 {
		t.Fatalf("expected retrofit maintenance 0.02, got %.2f", retro.AnnualMaintenancePct)
	}
	if retro.LifespanYears != 15 {
		t.Fatalf("expected retrofit lifespan 15, got %d", retro.LifespanYears)
	}
	if !strings.Contains(retro.Citation, "Clevenger & Huijser") {
		t.Fatalf("expected citation with Clevenger & Huijser, got %s", retro.Citation)
	}

	// Check species costs
	expectedCosts := map[string]float64{
		"deer":                 14000.0,
		"mule_deer":            14000.0,
		"elk":                  45000.0,
		"moose":                83000.0,
		"pronghorn":            16000.0,
		"bighorn_sheep":        25000.0,
		"sierra_bighorn_sheep": 25000.0,
		"black_bear":           20000.0,
		"mountain_lion":        18000.0,
		"fisher":               15000.0,
		"pacific_fisher":       15000.0,
	}
	for sp, expCost := range expectedCosts {
		if math.Abs(cfg.SpeciesCosts[sp]-expCost) > 1e-2 {
			t.Fatalf("expected %s cost %.2f, got %.2f", sp, expCost, cfg.SpeciesCosts[sp])
		}
	}
	if math.Abs(cfg.DefaultSpeciesCost-15000.0) > 1e-2 {
		t.Fatalf("expected default species cost 15000, got %.2f", cfg.DefaultSpeciesCost)
	}
}

func TestCoordinateClusteringAndOptions(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	// Two segments touching by coordinate endpoints, zero mileposts
	segA := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "WY-22",
		State:          "WY",
		MilepostStart:  0,
		MilepostEnd:    0,
		LengthM:        160.0,
		Coordinates:    [][]float64{{-110.8, 43.5}, {-110.801, 43.501}},
		RiskScore:      2.0,
		PercentileRank: 95.0,
		P10:            1.0,
		P90:            3.0,
		SpeciesBreakdown: map[string]float64{
			"elk": 1.0,
		},
	}
	segB := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "WY-22",
		State:          "WY",
		MilepostStart:  0,
		MilepostEnd:    0,
		LengthM:        160.0,
		Coordinates:    [][]float64{{-110.801, 43.501}, {-110.802, 43.502}}, // touches segA
		RiskScore:      3.0,
		PercentileRank: 96.0,
		P10:            1.5,
		P90:            4.5,
		SpeciesBreakdown: map[string]float64{
			"elk": 1.0,
		},
	}
	segUT := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-89",
		State:          "UT", // Different state
		MilepostStart:  10.0,
		MilepostEnd:    10.1,
		LengthM:        160.0,
		RiskScore:      4.0,
		PercentileRank: 98.0,
		P10:            2.0,
		P90:            6.0,
		SpeciesBreakdown: map[string]float64{
			"deer": 1.0,
		},
	}

	// 1. Without state filter
	candsAll := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segA, segB, segUT}, -1.0, cfg)
	if len(candsAll) != 2 {
		t.Fatalf("expected 2 candidates (one clustered WY, one UT), got %d", len(candsAll))
	}

	// Verify WY cluster contains both segA and segB
	var wyCand *prioritize.CandidateSite
	for i := range candsAll {
		if candsAll[i].State == "WY" {
			wyCand = &candsAll[i]
			break
		}
	}
	if wyCand == nil {
		t.Fatalf("missing WY candidate")
	}
	if len(wyCand.Segments) != 2 {
		t.Fatalf("expected 2 segments in WY cluster, got %d", len(wyCand.Segments))
	}

	// 2. With state filter "UT"
	utState := "UT"
	candsUT := prioritize.GenerateCandidatesFromSegmentsWithOptions([]prioritize.SegmentRecord{segA, segB, segUT}, prioritize.PrioritizeOptions{
		State:         &utState,
		MinPercentile: 90.0,
		Budget:        -1.0,
	}, cfg)
	if len(candsUT) != 1 || candsUT[0].State != "UT" {
		t.Fatalf("expected 1 UT candidate, got %d", len(candsUT))
	}

	// 3. With MinBCR filter
	highBCR := 1000.0 // higher than any project
	candsHighBCR := prioritize.GenerateCandidatesFromSegmentsWithOptions([]prioritize.SegmentRecord{segA, segB, segUT}, prioritize.PrioritizeOptions{
		MinBCR:        &highBCR,
		MinPercentile: 90.0,
		Budget:        -1.0,
	}, cfg)
	if len(candsHighBCR) != 0 {
		t.Fatalf("expected 0 candidates with very high MinBCR, got %d", len(candsHighBCR))
	}
}

func TestSpeciesCostNormalizationAndAliases(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	testCases := []struct {
		inputSpecies string
		expectedCost float64
	}{
		{"mule_deer", 14000.0},
		{"Mule deer", 14000.0},
		{"MULE_DEER", 14000.0},
		{"  mule-deer  ", 14000.0},
		{"deer", 14000.0},
		{"Deer", 14000.0},
		{"bighorn_sheep", 25000.0},
		{"Bighorn Sheep", 25000.0},
		{"sierra_bighorn_sheep", 25000.0},
		{"Sierra Nevada bighorn sheep", 25000.0},
		{"black_bear", 20000.0},
		{"California black bear", 20000.0},
		{"fisher", 15000.0},
		{"pacific_fisher", 15000.0},
		{"Pacific Fisher", 15000.0},
		{"mountain_lion", 18000.0},
		{"Mountain Lion", 18000.0},
		{"elk", 45000.0},
		{"moose", 83000.0},
		{"pronghorn", 16000.0},
		{"unknown_species", 15000.0}, // Default cost
	}

	for _, tc := range testCases {
		seg := prioritize.SegmentRecord{
			ID:             uuid.New(),
			HighwayRef:     "US-395",
			State:          "CA",
			MilepostStart:  10.0,
			MilepostEnd:    10.1,
			LengthM:        160.0,
			RiskScore:      1.0,
			PercentileRank: 95.0,
			P10:            0.5,
			P90:            1.5,
			SpeciesBreakdown: map[string]float64{
				tc.inputSpecies: 1.0,
			},
		}

		cands := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{seg}, -1.0, cfg)
		if len(cands) != 1 {
			t.Fatalf("expected 1 candidate for species %s, got %d", tc.inputSpecies, len(cands))
		}

		cand := cands[0]
		// Avoided collisions = 1.0 * structure.Effectiveness * structure.LifespanYears
		// ExpectedSavingsUSD = avoided * cost
		expectedSavings := cand.ExpectedAvoidedCollisions * tc.expectedCost
		if math.Abs(cand.ExpectedSavingsUSD-expectedSavings) > 1e-2 {
			t.Errorf("for species %q: expected savings %.2f (based on cost %.2f), got %.2f",
				tc.inputSpecies, expectedSavings, tc.expectedCost, cand.ExpectedSavingsUSD)
		}
	}

	// Verify fallback mapping when mule_deer is not directly in SpeciesCosts map but deer is
	cfgWithoutMuleDeer := cfg
	delete(cfgWithoutMuleDeer.SpeciesCosts, "mule_deer")
	segMule := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "US-395",
		State:          "CA",
		MilepostStart:  10.0,
		MilepostEnd:    10.1,
		LengthM:        160.0,
		RiskScore:      1.0,
		PercentileRank: 95.0,
		SpeciesBreakdown: map[string]float64{
			"mule_deer": 1.0,
		},
	}
	candsMule := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segMule}, -1.0, cfgWithoutMuleDeer)
	if len(candsMule) != 1 {
		t.Fatalf("expected 1 candidate for mule_deer fallback")
	}
	expectedSavingsMule := candsMule[0].ExpectedAvoidedCollisions * 14000.0
	if math.Abs(candsMule[0].ExpectedSavingsUSD-expectedSavingsMule) > 1e-2 {
		t.Errorf("expected mule_deer to fallback to deer cost (14000), got savings %.2f", candsMule[0].ExpectedSavingsUSD)
	}
}

func TestRecommendStructureType_LargeMammals(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	// High collision rate (risk 1.6, length 200m).
	// For deer (not in large mammal list), largeMammalFrac is 0.0, risk < 2.5, length < 1200m -> recommends underpass.
	segDeer := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "CA-120",
		State:          "CA",
		MilepostStart:  20.0,
		MilepostEnd:    20.2,
		LengthM:        200.0,
		RiskScore:      1.6,
		PercentileRank: 95.0,
		SpeciesBreakdown: map[string]float64{
			"deer": 1.0,
		},
	}
	candsDeer := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segDeer}, -1.0, cfg)
	if len(candsDeer) != 1 {
		t.Fatalf("expected 1 candidate for deer")
	}
	if candsDeer[0].RecommendedType != "underpass_with_fencing" {
		t.Errorf("expected underpass_with_fencing for deer with risk 1.6, got %s", candsDeer[0].RecommendedType)
	}

	// For bighorn sheep, largeMammalFrac >= 0.25 and risk >= 1.5 -> recommends overpass_with_fencing!
	largeMammals := []string{"bighorn_sheep", "sierra_bighorn_sheep", "black_bear"}
	for _, sp := range largeMammals {
		seg := prioritize.SegmentRecord{
			ID:             uuid.New(),
			HighwayRef:     "CA-120",
			State:          "CA",
			MilepostStart:  20.0,
			MilepostEnd:    20.2,
			LengthM:        200.0,
			RiskScore:      1.6,
			PercentileRank: 95.0,
			SpeciesBreakdown: map[string]float64{
				sp: 1.0,
			},
		}
		cands := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{seg}, -1.0, cfg)
		if len(cands) != 1 {
			t.Fatalf("expected 1 candidate for %s", sp)
		}
		if cands[0].RecommendedType != "overpass_with_fencing" {
			t.Errorf("for large mammal %s: expected overpass_with_fencing, got %s", sp, cands[0].RecommendedType)
		}
	}
}

func TestRecommendStructureType_SmallCarnivoreRetrofit(t *testing.T) {
	cfg := prioritize.DefaultConfig()

	// 1. Dominant small carnivore species should recommend small_carnivore_culvert_retrofit
	smallCarnivores := []string{"fisher", "pacific_fisher", "raccoon"}
	for _, sp := range smallCarnivores {
		seg := prioritize.SegmentRecord{
			ID:             uuid.New(),
			HighwayRef:     "CA-41",
			State:          "CA",
			MilepostStart:  5.0,
			MilepostEnd:    5.2,
			LengthM:        200.0,
			RiskScore:      1.5,
			PercentileRank: 95.0,
			SpeciesBreakdown: map[string]float64{
				sp:     0.8,
				"deer": 0.2,
			},
		}
		cands := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{seg}, -1.0, cfg)
		if len(cands) != 1 {
			t.Fatalf("expected 1 candidate for %s", sp)
		}
		if cands[0].RecommendedType != "small_carnivore_culvert_retrofit" {
			t.Errorf("for dominant species %s: expected small_carnivore_culvert_retrofit, got %s",
				sp, cands[0].RecommendedType)
		}
	}

	// 2. Non-dominant small carnivore should NOT recommend culvert retrofit
	segNonDominant := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "CA-41",
		State:          "CA",
		MilepostStart:  5.0,
		MilepostEnd:    5.2,
		LengthM:        200.0,
		RiskScore:      1.5,
		PercentileRank: 95.0,
		SpeciesBreakdown: map[string]float64{
			"deer":   0.8,
			"fisher": 0.2,
		},
	}
	candsNonDom := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segNonDominant}, -1.0, cfg)
	if len(candsNonDom) != 1 {
		t.Fatalf("expected 1 candidate for non-dominant")
	}
	if candsNonDom[0].RecommendedType == "small_carnivore_culvert_retrofit" {
		t.Errorf("expected non-dominant fisher to not trigger retrofit, got %s", candsNonDom[0].RecommendedType)
	}

	// 3. If small_carnivore_culvert_retrofit is NOT available in config, should fall back
	cfgWithoutRetro := cfg
	delete(cfgWithoutRetro.Structures, "small_carnivore_culvert_retrofit")
	segFisher := prioritize.SegmentRecord{
		ID:             uuid.New(),
		HighwayRef:     "CA-41",
		State:          "CA",
		MilepostStart:  5.0,
		MilepostEnd:    5.2,
		LengthM:        200.0,
		RiskScore:      1.5,
		PercentileRank: 95.0,
		SpeciesBreakdown: map[string]float64{
			"fisher": 1.0,
		},
	}
	candsFallback := prioritize.GenerateCandidatesFromSegments([]prioritize.SegmentRecord{segFisher}, -1.0, cfgWithoutRetro)
	if len(candsFallback) != 1 {
		t.Fatalf("expected 1 candidate for fallback")
	}
	if candsFallback[0].RecommendedType == "small_carnivore_culvert_retrofit" {
		t.Errorf("expected fallback when retrofit structure is not configured")
	}
}

func TestSensitiveSpeciesConfigFile(t *testing.T) {
	path := filepath.Join("..", "..", "config", "sensitive_species.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read sensitive_species.yaml: %v", err)
	}

	var doc struct {
		Privacy struct {
			Generalization struct {
				System     string `yaml:"system"`
				Resolution int    `yaml:"resolution"`
			} `yaml:"generalization"`
			Quarantine struct {
				DelayDays int `yaml:"delay_days"`
			} `yaml:"quarantine"`
		} `yaml:"privacy"`
		Species []struct {
			CommonName     string   `yaml:"common_name"`
			ScientificName string   `yaml:"scientific_name"`
			Status         string   `yaml:"status"`
			Aliases        []string `yaml:"aliases"`
		} `yaml:"species"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("failed to unmarshal sensitive_species.yaml: %v", err)
	}

	// Verify Uber H3 resolution 6 and 30-day quarantine delay
	if doc.Privacy.Generalization.System != "Uber H3" {
		t.Errorf("expected Uber H3, got %s", doc.Privacy.Generalization.System)
	}
	if doc.Privacy.Generalization.Resolution != 6 {
		t.Errorf("expected H3 resolution 6, got %d", doc.Privacy.Generalization.Resolution)
	}
	if doc.Privacy.Quarantine.DelayDays != 30 {
		t.Errorf("expected 30 day quarantine delay, got %d", doc.Privacy.Quarantine.DelayDays)
	}

	// Verify required species
	requiredSpecies := map[string]bool{
		"Sierra Nevada bighorn sheep": false,
		"Pacific fisher":              false,
		"California black bear":       false,
		"Grizzly bear":                false,
		"Canada lynx":                 false,
		"Gray wolf":                   false,
		"Wolverine":                   false,
	}

	for _, s := range doc.Species {
		for req := range requiredSpecies {
			if strings.EqualFold(s.CommonName, req) {
				requiredSpecies[req] = true
			}
		}
	}

	for req, found := range requiredSpecies {
		if !found {
			t.Errorf("missing required sensitive species %q in config/sensitive_species.yaml", req)
		}
	}
}
