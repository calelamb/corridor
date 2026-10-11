package hotspot

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// GiResult represents the Getis-Ord Gi* statistic and significance for a road segment.
type GiResult struct {
	SegmentID          string
	ZScore             float64
	PValue             float64
	FDRCorrectedPValue float64
	Significance       string
}

// ComputeGetisOrdGi calculates Getis-Ord Gi* with immediate adjacent neighbors (1-hop including self).
func (g *RoadGraph) ComputeGetisOrdGi(scores []float64, permutations int) ([]GiResult, error) {
	return g.ComputeGetisOrdGiWithThreshold(scores, 0, permutations)
}

// ComputeGetisOrdGiWithThreshold calculates Getis-Ord Gi* using a network distance threshold.
// If thresholdMeters <= 0, immediate 1-hop adjacency (including self) is used.
func (g *RoadGraph) ComputeGetisOrdGiWithThreshold(scores []float64, thresholdMeters float64, permutations int) ([]GiResult, error) {
	segments := g.Segments()
	n := len(segments)
	if len(scores) != n {
		return nil, fmt.Errorf("number of scores (%d) must match number of segments in graph (%d)", len(scores), n)
	}

	if n == 0 {
		return []GiResult{}, nil
	}

	// Single segment edge case
	if n == 1 {
		return []GiResult{
			{
				SegmentID:          segments[0].ID,
				ZScore:             0.0,
				PValue:             1.0,
				FDRCorrectedPValue: 1.0,
				Significance:       "Not Significant",
			},
		}, nil
	}

	// Calculate sample mean and standard deviation
	var sum, sumSq float64
	for _, x := range scores {
		sum += x
		sumSq += x * x
	}
	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - (mean * mean)

	// If variance is 0 (all scores identical / uniform)
	if variance <= 1e-12 {
		results := make([]GiResult, n)
		for i, seg := range segments {
			results[i] = GiResult{
				SegmentID:          seg.ID,
				ZScore:             0.0,
				PValue:             1.0,
				FDRCorrectedPValue: 1.0,
				Significance:       "Not Significant",
			}
		}
		return results, nil
	}
	stdDev := math.Sqrt(variance)

	// Build index mapping for segments
	idToIndex := make(map[string]int, n)
	for i, seg := range segments {
		idToIndex[seg.ID] = i
	}

	// Precompute spatial weights for each segment
	// neighborIndices[i] contains indices of neighbors j where w_{ij} = 1 (including i itself)
	neighborIndices := make([][]int, n)
	for i, seg := range segments {
		neighbors := []int{i} // Gi* includes self (w_ii = 1)
		if thresholdMeters > 0 {
			dists, err := g.ShortestDistances(seg.ID, thresholdMeters)
			if err == nil {
				for targetID, d := range dists {
					if targetID != seg.ID && d <= thresholdMeters {
						if idx, ok := idToIndex[targetID]; ok {
							neighbors = append(neighbors, idx)
						}
					}
				}
			}
		} else {
			// Immediate 1-hop adjacency
			for _, edge := range g.adj[seg.ID] {
				if idx, ok := idToIndex[edge.To]; ok {
					neighbors = append(neighbors, idx)
				}
			}
		}
		neighborIndices[i] = neighbors
	}

	zScores := make([]float64, n)
	rawPValues := make([]float64, n)
	observedSums := make([]float64, n)
	wStars := make([]float64, n)
	denoms := make([]float64, n)

	fn := float64(n)
	for i := 0; i < n; i++ {
		neighbors := neighborIndices[i]
		var localSum float64
		for _, j := range neighbors {
			localSum += scores[j]
		}
		observedSums[i] = localSum

		wStar := float64(len(neighbors))
		wStars[i] = wStar
		s1 := wStar // for binary weights, sum(w^2) == sum(w) == wStar

		num := localSum - (wStar * mean)
		denomRad := (fn*s1 - wStar*wStar) / (fn - 1.0)
		if denomRad <= 0 {
			zScores[i] = 0.0
			denoms[i] = 0.0
			rawPValues[i] = 1.0
			continue
		}
		denom := stdDev * math.Sqrt(denomRad)
		denoms[i] = denom

		if denom > 0 {
			z := num / denom
			zScores[i] = z
		} else {
			zScores[i] = 0.0
		}
	}

	// Compute p-values
	if permutations <= 0 {
		// Asymptotic standard normal distribution directional p-values
		for i := 0; i < n; i++ {
			if denoms[i] <= 0 {
				rawPValues[i] = 1.0
			} else {
				z := zScores[i]
				if z >= 0 {
					// Upper-tail p-value for hotspot: 1 - Phi(z) = 0.5 * erfc(z / sqrt(2))
					rawPValues[i] = 0.5 * math.Erfc(z/math.Sqrt2)
				} else {
					// Lower-tail p-value for coldspot: Phi(z) = 0.5 * erfc(-z / sqrt(2))
					rawPValues[i] = 0.5 * math.Erfc(-z/math.Sqrt2)
				}
				if rawPValues[i] > 1.0 {
					rawPValues[i] = 1.0
				}
			}
		}
	} else {
		// Monte Carlo permutation test
		rng := rand.New(rand.NewSource(42))
		extremeCounts := make([]int, n)

		// Pre-allocate slice for permuted values
		permScores := make([]float64, n)
		copy(permScores, scores)

		for p := 0; p < permutations; p++ {
			// Fisher-Yates shuffle
			rng.Shuffle(n, func(i, j int) {
				permScores[i], permScores[j] = permScores[j], permScores[i]
			})

			for i := 0; i < n; i++ {
				if denoms[i] <= 0 {
					continue
				}
				var permLocalSum float64
				for _, j := range neighborIndices[i] {
					permLocalSum += permScores[j]
				}

				if zScores[i] >= 0 {
					// Testing upper tail for hotspot
					if permLocalSum >= observedSums[i]-1e-12 {
						extremeCounts[i]++
					}
				} else {
					// Testing lower tail for coldspot
					if permLocalSum <= observedSums[i]+1e-12 {
						extremeCounts[i]++
					}
				}
			}
		}

		for i := 0; i < n; i++ {
			if denoms[i] <= 0 {
				rawPValues[i] = 1.0
			} else {
				// Pseudo p-value: (extreme_count + 1) / (permutations + 1)
				rawPValues[i] = float64(extremeCounts[i]+1) / float64(permutations+1)
			}
		}
	}

	// Apply Benjamini-Hochberg FDR correction
	fdrPValues := BenjaminiHochberg(rawPValues)

	// Assemble results
	results := make([]GiResult, n)
	for i, seg := range segments {
		z := zScores[i]
		q := fdrPValues[i]

		var sig string
		if q <= 0.01 && z > 0 {
			sig = "Hotspot (99% confidence)"
		} else if q <= 0.05 && z > 0 {
			sig = "Hotspot (95% confidence)"
		} else if q <= 0.10 && z > 0 {
			sig = "Hotspot (90% confidence)"
		} else if q <= 0.01 && z < 0 {
			sig = "Coldspot (99% confidence)"
		} else if q <= 0.05 && z < 0 {
			sig = "Coldspot (95% confidence)"
		} else if q <= 0.10 && z < 0 {
			sig = "Coldspot (90% confidence)"
		} else {
			sig = "Not Significant"
		}

		results[i] = GiResult{
			SegmentID:          seg.ID,
			ZScore:             z,
			PValue:             rawPValues[i],
			FDRCorrectedPValue: q,
			Significance:       sig,
		}
	}

	return results, nil
}

type pValRank struct {
	index int
	p     float64
}

// BenjaminiHochberg adjusts raw p-values using the Benjamini-Hochberg False Discovery Rate (FDR) procedure.
func BenjaminiHochberg(pValues []float64) []float64 {
	m := len(pValues)
	if m == 0 {
		return []float64{}
	}
	if m == 1 {
		res := make([]float64, 1)
		res[0] = math.Min(1.0, math.Max(0.0, pValues[0]))
		return res
	}

	ranks := make([]pValRank, m)
	for i, p := range pValues {
		ranks[i] = pValRank{index: i, p: p}
	}

	// Sort p-values ascending
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].p < ranks[j].p
	})

	qVals := make([]float64, m)
	fm := float64(m)
	for k := 0; k < m; k++ {
		rank := float64(k + 1) // 1-based rank
		q := ranks[k].p * fm / rank
		if q > 1.0 {
			q = 1.0
		}
		qVals[k] = q
	}

	// Enforce monotonicity backwards: q[k] = min(q[k], q[k+1])
	for k := m - 2; k >= 0; k-- {
		if qVals[k] > qVals[k+1] {
			qVals[k] = qVals[k+1]
		}
	}

	// Enforce q >= p and map back to original indices
	adjusted := make([]float64, m)
	for k := 0; k < m; k++ {
		origIdx := ranks[k].index
		finalQ := qVals[k]
		if finalQ < ranks[k].p {
			finalQ = ranks[k].p
		}
		if finalQ > 1.0 {
			finalQ = 1.0
		}
		adjusted[origIdx] = finalQ
	}

	return adjusted
}
