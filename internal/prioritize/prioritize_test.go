package prioritize_test

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"corridor/internal/prioritize"

	"github.com/google/uuid"
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

	// Check species costs
	if math.Abs(cfg.SpeciesCosts["deer"]-14000.0) > 1e-2 {
		t.Fatalf("expected deer cost 14000, got %.2f", cfg.SpeciesCosts["deer"])
	}
	if math.Abs(cfg.SpeciesCosts["elk"]-45000.0) > 1e-2 {
		t.Fatalf("expected elk cost 45000, got %.2f", cfg.SpeciesCosts["elk"])
	}
	if math.Abs(cfg.SpeciesCosts["moose"]-83000.0) > 1e-2 {
		t.Fatalf("expected moose cost 83000, got %.2f", cfg.SpeciesCosts["moose"])
	}
	if math.Abs(cfg.SpeciesCosts["pronghorn"]-16000.0) > 1e-2 {
		t.Fatalf("expected pronghorn cost 16000, got %.2f", cfg.SpeciesCosts["pronghorn"])
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

