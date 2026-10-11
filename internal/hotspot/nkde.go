package hotspot

import (
	"errors"
	"math"
)

// KernelType defines the kernel weighting function for NKDE.
type KernelType string

const (
	KernelEpanechnikov KernelType = "epanechnikov"
	KernelGaussian     KernelType = "gaussian"
)

// EventPoint represents a collision or observation on a road segment.
type EventPoint struct {
	ID                    string
	SegmentID             string
	DistanceAlongSegmentM float64
	Weight                float64
}

// SegmentScore represents the computed density and raw count for a road segment.
type SegmentScore struct {
	SegmentID string
	Density   float64
	RawCount  int
}

// ComputeNKDE calculates Network Kernel Density Estimation using the Epanechnikov kernel.
func (g *RoadGraph) ComputeNKDE(events []EventPoint, bandwidthMeters float64) ([]SegmentScore, error) {
	return g.ComputeNKDEWithKernel(events, bandwidthMeters, KernelEpanechnikov)
}

// ComputeNKDEWithKernel calculates Network Kernel Density Estimation with a specified kernel function.
func (g *RoadGraph) ComputeNKDEWithKernel(events []EventPoint, bandwidthMeters float64, kernel KernelType) ([]SegmentScore, error) {
	if bandwidthMeters <= 0 {
		return nil, errors.New("bandwidth must be positive")
	}

	segments := g.Segments()
	densityMap := make(map[string]float64, len(segments))
	countMap := make(map[string]int, len(segments))

	for _, seg := range segments {
		densityMap[seg.ID] = 0.0
		countMap[seg.ID] = 0
	}

	if len(segments) == 0 || len(events) == 0 {
		res := make([]SegmentScore, len(segments))
		for i, seg := range segments {
			res[i] = SegmentScore{
				SegmentID: seg.ID,
				Density:   0.0,
				RawCount:  0,
			}
		}
		return res, nil
	}

	// Cache shortest path distances for segments that contain events
	distancesCache := make(map[string]map[string]float64)

	for _, ev := range events {
		if _, exists := densityMap[ev.SegmentID]; !exists {
			// Event is on a segment not present in the graph; skip
			continue
		}

		countMap[ev.SegmentID]++

		w := ev.Weight
		if w <= 0 {
			w = 1.0
		}

		dists, cached := distancesCache[ev.SegmentID]
		if !cached {
			var err error
			dists, err = g.ShortestDistances(ev.SegmentID, bandwidthMeters)
			if err != nil {
				return nil, err
			}
			distancesCache[ev.SegmentID] = dists
		}

		segInfo, _ := g.Segment(ev.SegmentID)
		segMid := 0.0
		if segInfo != nil && segInfo.LengthM > 0 {
			segMid = segInfo.LengthM / 2.0
		}

		for targetID, netDist := range dists {
			var d float64
			if targetID == ev.SegmentID {
				// Distance along segment relative to segment midpoint
				if segInfo != nil && segInfo.LengthM > 0 && ev.DistanceAlongSegmentM > 0 {
					d = math.Abs(ev.DistanceAlongSegmentM - segMid)
				} else {
					d = 0.0
				}
			} else {
				d = netDist
			}

			if d > bandwidthMeters {
				continue
			}

			var k float64
			u := d / bandwidthMeters
			switch kernel {
			case KernelGaussian:
				// Gaussian kernel scaled by 1/h: (1 / (h * sqrt(2*pi))) * exp(-0.5 * u^2)
				k = (1.0 / (bandwidthMeters * math.Sqrt(2*math.Pi))) * math.Exp(-0.5*u*u)
			case KernelEpanechnikov:
				fallthrough
			default:
				// Epanechnikov kernel: (3 / (4 * h)) * (1 - u^2) for u <= 1
				if u <= 1.0 {
					k = (3.0 / (4.0 * bandwidthMeters)) * (1.0 - u*u)
				}
			}

			densityMap[targetID] += w * k
		}
	}

	results := make([]SegmentScore, len(segments))
	for i, seg := range segments {
		results[i] = SegmentScore{
			SegmentID: seg.ID,
			Density:   densityMap[seg.ID],
			RawCount:  countMap[seg.ID],
		}
	}

	return results, nil
}
