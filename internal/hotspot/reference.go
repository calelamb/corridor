package hotspot

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// HotspotResult represents the combined hotspot calculation output for a segment.
type HotspotResult struct {
	SegmentID    string
	Score        float64
	RawCount     int
	ZScore       *float64
	PValue       *float64
	FDRPValue    *float64
	Significance string
	HighwayRef   string
	State        string
}

// AnalyzeHotspots runs statistical hotspot analysis (NKDE and Getis-Ord Gi*) on the graph.
func (g *RoadGraph) AnalyzeHotspots(events []EventPoint, method string, bandwidthMeters float64) ([]HotspotResult, error) {
	if bandwidthMeters <= 0 {
		return nil, errors.New("bandwidth must be positive")
	}

	normMethod := strings.ToLower(strings.TrimSpace(method))
	if normMethod == "" {
		normMethod = "nkde"
	}
	if normMethod != "nkde" && normMethod != "getis_ord" && normMethod != "getis-ord" && normMethod != "density" {
		return nil, fmt.Errorf("invalid method %q: must be 'nkde', 'getis_ord', or 'density'", method)
	}

	segments := g.Segments()
	if len(segments) == 0 {
		return []HotspotResult{}, nil
	}

	// Always compute NKDE to obtain network kernel densities and raw counts per segment
	nkdeScores, err := g.ComputeNKDE(events, bandwidthMeters)
	if err != nil {
		return nil, err
	}

	scoreMap := make(map[string]SegmentScore, len(nkdeScores))
	for _, s := range nkdeScores {
		scoreMap[s.SegmentID] = s
	}

	densities := make([]float64, len(segments))
	rawCounts := make([]float64, len(segments))
	for i, seg := range segments {
		s := scoreMap[seg.ID]
		densities[i] = s.Density
		rawCounts[i] = float64(s.RawCount)
	}

	var giScores []float64
	if normMethod == "nkde" {
		giScores = densities
	} else {
		giScores = rawCounts
	}

	giResults, err := g.ComputeGetisOrdGiWithThreshold(giScores, bandwidthMeters, 500)
	if err != nil {
		return nil, err
	}

	giMap := make(map[string]GiResult, len(giResults))
	for _, r := range giResults {
		giMap[r.SegmentID] = r
	}

	results := make([]HotspotResult, len(segments))
	for i, seg := range segments {
		s := scoreMap[seg.ID]
		gi := giMap[seg.ID]

		scoreVal := s.Density
		if normMethod == "getis_ord" || normMethod == "getis-ord" {
			scoreVal = gi.ZScore
		} else if normMethod == "density" {
			scoreVal = float64(s.RawCount)
		}

		z := gi.ZScore
		p := gi.PValue
		fdr := gi.FDRCorrectedPValue

		results[i] = HotspotResult{
			SegmentID:    seg.ID,
			Score:        scoreVal,
			RawCount:     s.RawCount,
			ZScore:       &z,
			PValue:       &p,
			FDRPValue:    &fdr,
			Significance: gi.Significance,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// ReferenceCorridor holds graph, events, and geographic metadata for a reference highway corridor.
type ReferenceCorridor struct {
	ID          string
	Name        string
	HighwayRef  string
	State       string
	Park        string
	BBox        [4]float64 // [west, south, east, north]
	Graph       *RoadGraph
	Events      []EventPoint
	SegmentGeos map[string][][2]float64
}

// SegmentInBBox checks if any point of a segment's coordinates lies inside the bounding box.
func SegmentInBBox(coords [][2]float64, bbox [4]float64) bool {
	west, south, east, north := bbox[0], bbox[1], bbox[2], bbox[3]
	for _, pt := range coords {
		lon, lat := pt[0], pt[1]
		if lon >= west && lon <= east && lat >= south && lat <= north {
			return true
		}
	}
	return false
}

// CorridorOverlapsBBox checks if two bounding boxes overlap.
func CorridorOverlapsBBox(a, b [4]float64) bool {
	// a: [west, south, east, north], b: [west, south, east, north]
	return a[0] <= b[2] && a[2] >= b[0] && a[1] <= b[3] && a[3] >= b[1]
}

// ReferenceWyomingCorridor returns the Yellowstone / Grand Teton US-191 reference corridor.
func ReferenceWyomingCorridor() ReferenceCorridor {
	const numSegments = 25
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	// US-191 corridor coordinates around MP 10 (Yellowstone / Grand Teton)
	baseLon := -110.7000
	baseLat := 44.4000
	dLat := 0.00145 // ~161 meters per step in latitude

	for i := 0; i < numSegments; i++ {
		id := fmt.Sprintf("seg-%02d", i)
		_ = g.AddSegment(id, segLengthM)
		if i > 0 {
			prev := fmt.Sprintf("seg-%02d", i-1)
			_ = g.AddEdge(prev, id, segLengthM)
		}
		lat0 := baseLat + float64(i)*dLat
		lat1 := baseLat + float64(i+1)*dLat
		geos[id] = [][2]float64{
			{baseLon, lat0},
			{baseLon, lat1},
		}
	}

	var events []EventPoint
	addCluster := func(segID string, count int, weight float64) {
		for i := 0; i < count; i++ {
			events = append(events, EventPoint{
				ID:                    fmt.Sprintf("ev-%s-%d", segID, i),
				SegmentID:             segID,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                weight,
			})
		}
	}

	addCluster("seg-12", 20, 1.0)
	addCluster("seg-11", 8, 1.0)
	addCluster("seg-13", 7, 1.0)
	addCluster("seg-10", 3, 1.0)
	addCluster("seg-14", 2, 1.0)

	return ReferenceCorridor{
		ID:          "wy-us191",
		Name:        "Yellowstone US-191 Gallatin Corridor",
		HighwayRef:  "US-191",
		State:       "WY",
		Park:        "Yellowstone",
		BBox:        [4]float64{-110.75, 44.38, -110.65, 44.46},
		Graph:       g,
		Events:      events,
		SegmentGeos: geos,
	}
}

// ReferenceYosemiteCorridor returns the CA-120 Crane Flat to Tuolumne Meadows corridor with black bear and mule deer event clusters.
func ReferenceYosemiteCorridor() ReferenceCorridor {
	const numSegments = 30
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	// CA-120 between Crane Flat (-119.8142, 37.7505) and Tuolumne Meadows (-119.3563, 37.8732)
	startLon, startLat := -119.8142, 37.7505
	endLon, endLat := -119.3563, 37.8732
	dLon := (endLon - startLon) / float64(numSegments)
	dLat := (endLat - startLat) / float64(numSegments)

	for i := 0; i < numSegments; i++ {
		id := fmt.Sprintf("ca-120-seg-%02d", i)
		_ = g.AddSegment(id, segLengthM)
		if i > 0 {
			prev := fmt.Sprintf("ca-120-seg-%02d", i-1)
			_ = g.AddEdge(prev, id, segLengthM)
		}
		lon0 := startLon + float64(i)*dLon
		lat0 := startLat + float64(i)*dLat
		lon1 := startLon + float64(i+1)*dLon
		lat1 := startLat + float64(i+1)*dLat
		geos[id] = [][2]float64{
			{lon0, lat0},
			{lon1, lat1},
		}
	}

	var events []EventPoint
	addCluster := func(segID string, count int, weight float64, species string) {
		for i := 0; i < count; i++ {
			events = append(events, EventPoint{
				ID:                    fmt.Sprintf("ev-%s-%s-%d", species, segID, i),
				SegmentID:             segID,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                weight,
			})
		}
	}

	// Crane Flat black bear crossings (segments 3-6, peak segment 4)
	addCluster("ca-120-seg-04", 18, 1.0, "bear")
	addCluster("ca-120-seg-05", 10, 1.0, "bear")
	addCluster("ca-120-seg-03", 7, 1.0, "bear")
	addCluster("ca-120-seg-06", 4, 1.0, "bear")

	// Tuolumne Meadows mule deer crossings (segments 23-26, peak segment 24)
	addCluster("ca-120-seg-24", 16, 1.0, "deer")
	addCluster("ca-120-seg-25", 9, 1.0, "deer")
	addCluster("ca-120-seg-23", 6, 1.0, "deer")
	addCluster("ca-120-seg-26", 3, 1.0, "deer")

	// Sporadic baseline observations along central Tioga Road
	addCluster("ca-120-seg-12", 1, 1.0, "deer")
	addCluster("ca-120-seg-16", 1, 1.0, "bear")

	return ReferenceCorridor{
		ID:          "ca-120-yosemite",
		Name:        "Yosemite CA-120 Tioga Road Corridor (Crane Flat to Tuolumne Meadows)",
		HighwayRef:  "CA-120",
		State:       "CA",
		Park:        "Yosemite",
		BBox:        [4]float64{-119.85, 37.74, -119.34, 37.89},
		Graph:       g,
		Events:      events,
		SegmentGeos: geos,
	}
}

// AllReferenceCorridors returns both reference highway corridors.
func AllReferenceCorridors() []ReferenceCorridor {
	return []ReferenceCorridor{
		ReferenceWyomingCorridor(),
		ReferenceYosemiteCorridor(),
	}
}

// FindReferenceCorridor locates the reference corridor matching state or park name.
func FindReferenceCorridor(state, park string) *ReferenceCorridor {
	normState := strings.ToUpper(strings.TrimSpace(state))
	normPark := strings.ToLower(strings.TrimSpace(park))

	if normState == "CA" || strings.Contains(normPark, "yosemite") {
		c := ReferenceYosemiteCorridor()
		return &c
	}
	if normState == "WY" || strings.Contains(normPark, "yellowstone") || strings.Contains(normPark, "teton") {
		c := ReferenceWyomingCorridor()
		return &c
	}
	return nil
}
