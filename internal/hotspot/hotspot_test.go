package hotspot_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"corridor/internal/hotspot"
)

func buildLinearHighway(n int, segmentLengthM float64) *hotspot.RoadGraph {
	g := hotspot.NewRoadGraph()
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("seg-%02d", i)
		_ = g.AddSegment(id, segmentLengthM)
		if i > 0 {
			prev := fmt.Sprintf("seg-%02d", i-1)
			_ = g.AddEdge(prev, id, segmentLengthM)
		}
	}
	return g
}

func TestNKDE_LinearHighway(t *testing.T) {
	// 25 segments, ~0.1 mile (160.934 meters) each
	const numSegments = 25
	const segLength = 160.934
	g := buildLinearHighway(numSegments, segLength)

	// Known cluster at segment 12: 20 events on seg-12, 5 on seg-11, 5 on seg-13
	var events []hotspot.EventPoint
	for i := 0; i < 20; i++ {
		events = append(events, hotspot.EventPoint{
			ID:                    fmt.Sprintf("ev-c-%d", i),
			SegmentID:             "seg-12",
			DistanceAlongSegmentM: segLength / 2,
			Weight:                1.0,
		})
	}
	for i := 0; i < 5; i++ {
		events = append(events, hotspot.EventPoint{
			ID:                    fmt.Sprintf("ev-l-%d", i),
			SegmentID:             "seg-11",
			DistanceAlongSegmentM: segLength / 2,
			Weight:                1.0,
		})
		events = append(events, hotspot.EventPoint{
			ID:                    fmt.Sprintf("ev-r-%d", i),
			SegmentID:             "seg-13",
			DistanceAlongSegmentM: segLength / 2,
			Weight:                1.0,
		})
	}

	bandwidthM := 500.0 // ~3 segments in each direction
	scores, err := g.ComputeNKDE(events, bandwidthM)
	if err != nil {
		t.Fatalf("ComputeNKDE failed: %v", err)
	}

	if len(scores) != numSegments {
		t.Fatalf("expected %d scores, got %d", numSegments, len(scores))
	}

	scoreMap := make(map[string]hotspot.SegmentScore)
	for _, s := range scores {
		scoreMap[s.SegmentID] = s
	}

	// Verify RawCount
	if scoreMap["seg-12"].RawCount != 20 {
		t.Errorf("seg-12 expected RawCount 20, got %d", scoreMap["seg-12"].RawCount)
	}
	if scoreMap["seg-11"].RawCount != 5 {
		t.Errorf("seg-11 expected RawCount 5, got %d", scoreMap["seg-11"].RawCount)
	}
	if scoreMap["seg-13"].RawCount != 5 {
		t.Errorf("seg-13 expected RawCount 5, got %d", scoreMap["seg-13"].RawCount)
	}
	if scoreMap["seg-00"].RawCount != 0 {
		t.Errorf("seg-00 expected RawCount 0, got %d", scoreMap["seg-00"].RawCount)
	}

	// Verify NKDE density peaks at the cluster (seg-12)
	peakDensity := scoreMap["seg-12"].Density
	if peakDensity <= 0 {
		t.Fatalf("seg-12 peak density should be > 0, got %f", peakDensity)
	}

	// Density should fall off along network paths in both directions
	if scoreMap["seg-11"].Density >= peakDensity {
		t.Errorf("seg-11 density (%f) should be less than seg-12 (%f)", scoreMap["seg-11"].Density, peakDensity)
	}
	if scoreMap["seg-10"].Density >= scoreMap["seg-11"].Density {
		t.Errorf("seg-10 density (%f) should be less than seg-11 (%f)", scoreMap["seg-10"].Density, scoreMap["seg-11"].Density)
	}
	if scoreMap["seg-13"].Density >= peakDensity {
		t.Errorf("seg-13 density (%f) should be less than seg-12 (%f)", scoreMap["seg-13"].Density, peakDensity)
	}
	if scoreMap["seg-14"].Density >= scoreMap["seg-13"].Density {
		t.Errorf("seg-14 density (%f) should be less than seg-13 (%f)", scoreMap["seg-14"].Density, scoreMap["seg-13"].Density)
	}

	// Segments far beyond bandwidth should have 0 density
	if scoreMap["seg-00"].Density != 0 {
		t.Errorf("seg-00 should have 0 density beyond bandwidth, got %f", scoreMap["seg-00"].Density)
	}
	if scoreMap["seg-24"].Density != 0 {
		t.Errorf("seg-24 should have 0 density beyond bandwidth, got %f", scoreMap["seg-24"].Density)
	}
}

func TestGetisOrdGi_ClusterSignificance(t *testing.T) {
	const numSegments = 25
	const segLength = 160.934
	g := buildLinearHighway(numSegments, segLength)

	// Synthetic scores: high cluster around segment 12, near zero everywhere else
	scores := make([]float64, numSegments)
	for i := 0; i < numSegments; i++ {
		scores[i] = 1.0 // background baseline
	}
	scores[10] = 20.0
	scores[11] = 35.0
	scores[12] = 50.0
	scores[13] = 35.0
	scores[14] = 20.0

	permutations := 999
	results, err := g.ComputeGetisOrdGi(scores, permutations)
	if err != nil {
		t.Fatalf("ComputeGetisOrdGi failed: %v", err)
	}

	if len(results) != numSegments {
		t.Fatalf("expected %d results, got %d", numSegments, len(results))
	}

	resMap := make(map[string]hotspot.GiResult)
	for _, r := range results {
		resMap[r.SegmentID] = r
	}

	clusterRes := resMap["seg-12"]
	if clusterRes.ZScore <= 2.576 {
		t.Errorf("seg-12 ZScore should be > 2.576 (99%% confidence), got %f", clusterRes.ZScore)
	}
	if clusterRes.PValue >= 0.01 {
		t.Errorf("seg-12 PValue should be < 0.01, got %f", clusterRes.PValue)
	}
	if clusterRes.FDRCorrectedPValue >= 0.05 {
		t.Errorf("seg-12 FDRCorrectedPValue should be < 0.05, got %f", clusterRes.FDRCorrectedPValue)
	}
	if !strings.Contains(clusterRes.Significance, "Hotspot") {
		t.Errorf("seg-12 Significance should indicate Hotspot, got %q", clusterRes.Significance)
	}

	// Segments far away should not be significant
	edgeRes := resMap["seg-00"]
	if clusterRes.ZScore <= edgeRes.ZScore {
		t.Errorf("cluster ZScore (%f) should be substantially higher than edge (%f)", clusterRes.ZScore, edgeRes.ZScore)
	}
	if edgeRes.FDRCorrectedPValue < 0.05 {
		t.Errorf("seg-00 should not be significant, got FDR p-value %f", edgeRes.FDRCorrectedPValue)
	}

	// Asymptotic significance test (permutations = 0)
	asympResults, err := g.ComputeGetisOrdGi(scores, 0)
	if err != nil {
		t.Fatalf("asymptotic ComputeGetisOrdGi failed: %v", err)
	}
	asympRes := asympResults[12]
	if asympRes.PValue >= 0.001 {
		t.Errorf("asymptotic PValue should be < 0.001, got %f", asympRes.PValue)
	}
	if asympRes.FDRCorrectedPValue >= 0.01 {
		t.Errorf("asymptotic FDRCorrectedPValue should be < 0.01, got %f", asympRes.FDRCorrectedPValue)
	}
	if asympRes.Significance != "Hotspot (99% confidence)" {
		t.Errorf("expected 99%% confidence hotspot, got %q", asympRes.Significance)
	}
}

func TestNKDE_EmptyEvents(t *testing.T) {
	g := buildLinearHighway(10, 160.934)
	scores, err := g.ComputeNKDE([]hotspot.EventPoint{}, 500.0)
	if err != nil {
		t.Fatalf("ComputeNKDE on empty events failed: %v", err)
	}
	if len(scores) != 10 {
		t.Fatalf("expected 10 scores, got %d", len(scores))
	}
	for _, s := range scores {
		if s.Density != 0 || s.RawCount != 0 {
			t.Errorf("segment %s: expected 0 density and count, got %f and %d", s.SegmentID, s.Density, s.RawCount)
		}
	}
}

func TestGetisOrdGi_UniformScores(t *testing.T) {
	g := buildLinearHighway(15, 160.934)
	scores := make([]float64, 15)
	for i := range scores {
		scores[i] = 10.0
	}

	results, err := g.ComputeGetisOrdGi(scores, 500)
	if err != nil {
		t.Fatalf("ComputeGetisOrdGi on uniform scores failed: %v", err)
	}
	if len(results) != 15 {
		t.Fatalf("expected 15 results, got %d", len(results))
	}
	for _, r := range results {
		if r.ZScore != 0.0 {
			t.Errorf("expected ZScore 0 for uniform scores, got %f", r.ZScore)
		}
		if r.PValue != 1.0 {
			t.Errorf("expected PValue 1.0 for uniform scores, got %f", r.PValue)
		}
		if !strings.Contains(strings.ToLower(r.Significance), "not significant") {
			t.Errorf("expected Significance to be not significant, got %q", r.Significance)
		}
	}
}

func TestSingleSegment(t *testing.T) {
	g := hotspot.NewRoadGraph()
	if err := g.AddSegment("solo", 160.934); err != nil {
		t.Fatalf("AddSegment failed: %v", err)
	}

	// NKDE on single segment
	events := []hotspot.EventPoint{
		{ID: "e1", SegmentID: "solo", DistanceAlongSegmentM: 80.0, Weight: 1.0},
		{ID: "e2", SegmentID: "solo", DistanceAlongSegmentM: 80.0, Weight: 2.0},
	}
	nkdeScores, err := g.ComputeNKDE(events, 500.0)
	if err != nil {
		t.Fatalf("NKDE single segment failed: %v", err)
	}
	if len(nkdeScores) != 1 || nkdeScores[0].RawCount != 2 || nkdeScores[0].Density <= 0 {
		t.Errorf("unexpected NKDE single segment result: %+v", nkdeScores)
	}

	// Getis-Ord on single segment
	giResults, err := g.ComputeGetisOrdGi([]float64{10.0}, 100)
	if err != nil {
		t.Fatalf("Getis-Ord single segment failed: %v", err)
	}
	if len(giResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(giResults))
	}
	if giResults[0].ZScore != 0 || giResults[0].PValue != 1.0 {
		t.Errorf("expected 0 z-score and 1.0 p-value for single segment, got %+v", giResults[0])
	}
}

func TestDisconnectedComponents(t *testing.T) {
	g := hotspot.NewRoadGraph()
	// Component A: seg-a1, seg-a2
	_ = g.AddSegment("seg-a1", 160.0)
	_ = g.AddSegment("seg-a2", 160.0)
	_ = g.AddEdge("seg-a1", "seg-a2", 160.0)

	// Component B: seg-b1, seg-b2 (disconnected from A)
	_ = g.AddSegment("seg-b1", 160.0)
	_ = g.AddSegment("seg-b2", 160.0)
	_ = g.AddEdge("seg-b1", "seg-b2", 160.0)

	// Events only in Component A
	events := []hotspot.EventPoint{
		{ID: "e-a", SegmentID: "seg-a1", DistanceAlongSegmentM: 80.0, Weight: 5.0},
	}

	scores, err := g.ComputeNKDE(events, 500.0)
	if err != nil {
		t.Fatalf("ComputeNKDE failed: %v", err)
	}

	scoreMap := make(map[string]hotspot.SegmentScore)
	for _, s := range scores {
		scoreMap[s.SegmentID] = s
	}

	if scoreMap["seg-a1"].Density <= 0 {
		t.Errorf("expected seg-a1 density > 0, got %f", scoreMap["seg-a1"].Density)
	}
	if scoreMap["seg-b1"].Density != 0 || scoreMap["seg-b2"].Density != 0 {
		t.Errorf("expected component B density to be 0, got b1=%f b2=%f", scoreMap["seg-b1"].Density, scoreMap["seg-b2"].Density)
	}
}

func TestBenjaminiHochbergFDR(t *testing.T) {
	// Verify Benjamini-Hochberg ranking and monotonicity
	// For raw p-values [0.01, 0.04, 0.05, 0.20], m = 4:
	// k=1: 0.01 * 4 / 1 = 0.04
	// k=2: 0.04 * 4 / 2 = 0.08
	// k=3: 0.05 * 4 / 3 = 0.0667 -> bounded by q(4) or adjusted monotonicity
	// In standard BH:
	// q(k) = min_{j >= k} (p(j) * m / j)
	rawP := []float64{0.01, 0.04, 0.05, 0.20}
	fdr := hotspot.BenjaminiHochberg(rawP)
	if len(fdr) != len(rawP) {
		t.Fatalf("expected %d FDR values, got %d", len(rawP), len(fdr))
	}
	for i := 0; i < len(fdr)-1; i++ {
		if fdr[i] > fdr[i+1]+1e-9 {
			t.Errorf("FDR monotonicity violated: fdr[%d]=%f > fdr[%d]=%f", i, fdr[i], i+1, fdr[i+1])
		}
	}
	for i, q := range fdr {
		if q < rawP[i]-1e-9 {
			t.Errorf("FDR q-value (%f) cannot be smaller than raw p-value (%f)", q, rawP[i])
		}
		if q > 1.0 {
			t.Errorf("FDR q-value cannot exceed 1.0, got %f", q)
		}
	}
}

func TestNKDE_KernelSelection(t *testing.T) {
	g := buildLinearHighway(5, 100.0)
	events := []hotspot.EventPoint{
		{ID: "e1", SegmentID: "seg-02", DistanceAlongSegmentM: 50.0, Weight: 1.0},
	}

	epanScores, err := g.ComputeNKDEWithKernel(events, 250.0, hotspot.KernelEpanechnikov)
	if err != nil {
		t.Fatalf("Epanechnikov failed: %v", err)
	}

	gaussScores, err := g.ComputeNKDEWithKernel(events, 250.0, hotspot.KernelGaussian)
	if err != nil {
		t.Fatalf("Gaussian failed: %v", err)
	}

	if len(epanScores) != 5 || len(gaussScores) != 5 {
		t.Fatal("unexpected score lengths")
	}

	// Both should peak at seg-02
	if epanScores[2].Density <= 0 || gaussScores[2].Density <= 0 {
		t.Errorf("expected positive density at center")
	}
	if math.Abs(epanScores[2].Density-gaussScores[2].Density) < 1e-6 {
		t.Errorf("different kernels should yield distinct density values")
	}
}

func TestRoadGraph_ValidationErrors(t *testing.T) {
	g := hotspot.NewRoadGraph()

	// Empty ID
	if err := g.AddSegment("", 100); err == nil {
		t.Error("expected error for empty segment ID")
	}
	// Non-positive length
	if err := g.AddSegment("s1", 0); err == nil {
		t.Error("expected error for zero segment length")
	}
	if err := g.AddSegment("s1", -5); err == nil {
		t.Error("expected error for negative segment length")
	}

	_ = g.AddSegment("s1", 100)
	_ = g.AddSegment("s2", 150)

	// Updating segment length
	_ = g.AddSegment("s1", 120)
	seg, ok := g.Segment("s1")
	if !ok || seg.LengthM != 120 {
		t.Errorf("expected updated length 120, got %+v", seg)
	}

	// Non-existent segment lookup
	if _, ok := g.Segment("missing"); ok {
		t.Error("expected false for missing segment")
	}

	// Self-loop
	if err := g.AddEdge("s1", "s1", 50); err == nil {
		t.Error("expected error for self-loop edge")
	}

	// Missing segments in edge
	if err := g.AddEdge("s1", "missing", 50); err == nil {
		t.Error("expected error for missing edge endpoint")
	}
	if err := g.AddEdge("missing", "s2", 50); err == nil {
		t.Error("expected error for missing edge startpoint")
	}

	// Negative distance
	if err := g.AddEdge("s1", "s2", -10); err == nil {
		t.Error("expected error for negative edge distance")
	}

	// AddConnection
	if err := g.AddConnection("s1", "s2"); err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}
	if err := g.AddConnection("s1", "missing"); err == nil {
		t.Error("expected error for AddConnection with missing segment")
	}

	// ShortestDistance
	d, err := g.ShortestDistance("s1", "s2")
	if err != nil {
		t.Fatalf("ShortestDistance failed: %v", err)
	}
	expectedDist := (120.0 + 150.0) / 2.0
	if math.Abs(d-expectedDist) > 1e-6 {
		t.Errorf("expected distance %f, got %f", expectedDist, d)
	}

	// Distance to self
	dSelf, err := g.ShortestDistance("s1", "s1")
	if err != nil || dSelf != 0 {
		t.Errorf("distance to self should be 0, got %f, err: %v", dSelf, err)
	}

	// Unreachable segment
	_ = g.AddSegment("island", 100)
	_, err = g.ShortestDistance("s1", "island")
	if err == nil {
		t.Error("expected error for unreachable segment")
	}
}

func TestNKDE_ValidationErrors(t *testing.T) {
	g := hotspot.NewRoadGraph()
	_ = g.AddSegment("s1", 100)

	// Bandwidth <= 0
	_, err := g.ComputeNKDE([]hotspot.EventPoint{}, 0)
	if err == nil {
		t.Error("expected error for 0 bandwidth")
	}
	_, err = g.ComputeNKDE([]hotspot.EventPoint{}, -100)
	if err == nil {
		t.Error("expected error for negative bandwidth")
	}
}

func TestGetisOrdGi_ValidationErrors(t *testing.T) {
	g := hotspot.NewRoadGraph()
	_ = g.AddSegment("s1", 100)
	_ = g.AddSegment("s2", 100)

	// Mismatched score length
	_, err := g.ComputeGetisOrdGi([]float64{1.0}, 100)
	if err == nil {
		t.Error("expected error for score length mismatch")
	}

	// Empty graph
	emptyG := hotspot.NewRoadGraph()
	emptyRes, err := emptyG.ComputeGetisOrdGi([]float64{}, 100)
	if err != nil || len(emptyRes) != 0 {
		t.Errorf("expected empty results for empty graph, got %v, err: %v", emptyRes, err)
	}
}

func TestGetisOrdGi_WithDistanceThreshold(t *testing.T) {
	// 10-segment linear highway
	g := buildLinearHighway(10, 100.0)
	scores := []float64{1, 1, 1, 1, 20, 20, 1, 1, 1, 1}

	// Threshold = 150m (covers immediate neighbors)
	results, err := g.ComputeGetisOrdGiWithThreshold(scores, 150.0, 500)
	if err != nil {
		t.Fatalf("ComputeGetisOrdGiWithThreshold failed: %v", err)
	}
	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
	if results[4].ZScore <= 0 || results[5].ZScore <= 0 {
		t.Errorf("expected positive z-scores at cluster, got %f, %f", results[4].ZScore, results[5].ZScore)
	}
}

