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

// ReferenceYellowstoneCorridor returns the Yellowstone US-191 Gallatin Canyon reference corridor.
func ReferenceYellowstoneCorridor() ReferenceCorridor {
	const numSegments = 25
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	// US-191 corridor coordinates around MP 10 (Yellowstone Gallatin Canyon)
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

	addCluster("seg-12", 20, 1.0, "elk")
	addCluster("seg-11", 8, 1.0, "moose")
	addCluster("seg-13", 7, 1.0, "grizzly_bear")
	addCluster("seg-10", 3, 1.0, "elk")
	addCluster("seg-14", 2, 1.0, "moose")

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

// ReferenceWyomingCorridor returns the Yellowstone US-191 reference corridor (alias for backward compatibility).
func ReferenceWyomingCorridor() ReferenceCorridor {
	return ReferenceYellowstoneCorridor()
}

// ReferenceGrandTetonCorridor returns the Grand Teton US-89 Jackson Hole corridor with mule deer and pronghorn clusters.
func ReferenceGrandTetonCorridor() ReferenceCorridor {
	const numSegments = 25
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	baseLon := -110.6600
	baseLat := 43.7000
	dLat := 0.00145

	for i := 0; i < numSegments; i++ {
		id := fmt.Sprintf("grte-seg-%02d", i)
		_ = g.AddSegment(id, segLengthM)
		if i > 0 {
			prev := fmt.Sprintf("grte-seg-%02d", i-1)
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
	addCluster := func(segID string, count int, weight float64, species string) {
		for i := 0; i < count; i++ {
			events = append(events, EventPoint{
				ID:                    fmt.Sprintf("ev-grte-%s-%s-%d", species, segID, i),
				SegmentID:             segID,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                weight,
			})
		}
	}

	// Blacktail Butte / Moose to Moran corridor cluster
	addCluster("grte-seg-08", 18, 1.0, "mule_deer")
	addCluster("grte-seg-09", 10, 1.0, "pronghorn")
	addCluster("grte-seg-07", 7, 1.0, "elk")
	addCluster("grte-seg-10", 3, 1.0, "mule_deer")

	return ReferenceCorridor{
		ID:          "wy-us89-teton",
		Name:        "Grand Teton US-89 Jackson Hole Corridor",
		HighwayRef:  "US-89",
		State:       "WY",
		Park:        "Grand Teton",
		BBox:        [4]float64{-110.75, 43.65, -110.55, 43.85},
		Graph:       g,
		Events:      events,
		SegmentGeos: geos,
	}
}

// ReferenceGlacierCorridor returns the Glacier US-2 Southern Boundary corridor with grizzly bear and lynx clusters.
func ReferenceGlacierCorridor() ReferenceCorridor {
	const numSegments = 30
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	startLon, startLat := -113.9800, 48.5000
	endLon, endLat := -113.3500, 48.3100
	dLon := (endLon - startLon) / float64(numSegments)
	dLat := (endLat - startLat) / float64(numSegments)

	for i := 0; i < numSegments; i++ {
		id := fmt.Sprintf("glac-seg-%02d", i)
		_ = g.AddSegment(id, segLengthM)
		if i > 0 {
			prev := fmt.Sprintf("glac-seg-%02d", i-1)
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
				ID:                    fmt.Sprintf("ev-glac-%s-%s-%d", species, segID, i),
				SegmentID:             segID,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                weight,
			})
		}
	}

	// Goat Lick / Marias Pass crossing hotspot
	addCluster("glac-seg-12", 20, 1.0, "grizzly_bear")
	addCluster("glac-seg-13", 10, 1.0, "bighorn_sheep")
	addCluster("glac-seg-11", 8, 1.0, "canada_lynx")
	addCluster("glac-seg-14", 4, 1.0, "grizzly_bear")

	return ReferenceCorridor{
		ID:          "mt-us2-glacier",
		Name:        "Glacier US-2 Southern Boundary Corridor",
		HighwayRef:  "US-2",
		State:       "MT",
		Park:        "Glacier",
		BBox:        [4]float64{-114.00, 48.25, -113.20, 48.55},
		Graph:       g,
		Events:      events,
		SegmentGeos: geos,
	}
}

// ReferenceRockyMountainCorridor returns the Rocky Mountain US-34 Trail Ridge Road corridor with elk and bighorn clusters.
func ReferenceRockyMountainCorridor() ReferenceCorridor {
	const numSegments = 25
	const segLengthM = 160.934

	g := NewRoadGraph()
	geos := make(map[string][][2]float64, numSegments)

	startLon, startLat := -105.8600, 40.3500
	endLon, endLat := -105.7400, 40.4300
	dLon := (endLon - startLon) / float64(numSegments)
	dLat := (endLat - startLat) / float64(numSegments)

	for i := 0; i < numSegments; i++ {
		id := fmt.Sprintf("romo-seg-%02d", i)
		_ = g.AddSegment(id, segLengthM)
		if i > 0 {
			prev := fmt.Sprintf("romo-seg-%02d", i-1)
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
				ID:                    fmt.Sprintf("ev-romo-%s-%s-%d", species, segID, i),
				SegmentID:             segID,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                weight,
			})
		}
	}

	// Milner Pass / Trail Ridge crossing hotspot
	addCluster("romo-seg-10", 19, 1.0, "elk")
	addCluster("romo-seg-11", 9, 1.0, "mule_deer")
	addCluster("romo-seg-09", 7, 1.0, "bighorn_sheep")
	addCluster("romo-seg-12", 3, 1.0, "elk")

	return ReferenceCorridor{
		ID:          "co-us34-romo",
		Name:        "Rocky Mountain US-34 Trail Ridge Road Corridor",
		HighwayRef:  "US-34",
		State:       "CO",
		Park:        "Rocky Mountain",
		BBox:        [4]float64{-105.90, 40.30, -105.65, 40.48},
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

// AllReferenceCorridors returns all 5 reference national park highway corridors.
func AllReferenceCorridors() []ReferenceCorridor {
	return []ReferenceCorridor{
		ReferenceYellowstoneCorridor(),
		ReferenceGrandTetonCorridor(),
		ReferenceGlacierCorridor(),
		ReferenceRockyMountainCorridor(),
		ReferenceYosemiteCorridor(),
	}
}

// FindReferenceCorridor locates the reference corridor matching state or park name/code.
func FindReferenceCorridor(state, park string) *ReferenceCorridor {
	normState := strings.ToUpper(strings.TrimSpace(state))
	normPark := strings.ToLower(strings.TrimSpace(park))
	normPark = strings.ReplaceAll(normPark, "-", " ")
	normPark = strings.ReplaceAll(normPark, "_", " ")

	// Yosemite (CA)
	if normPark == "yose" || strings.Contains(normPark, "yosemite") || (normState == "CA" && normPark == "") {
		c := ReferenceYosemiteCorridor()
		return &c
	}
	// Grand Teton (WY)
	if normPark == "grte" || strings.Contains(normPark, "grand teton") || strings.Contains(normPark, "teton") {
		c := ReferenceGrandTetonCorridor()
		return &c
	}
	// Glacier (MT)
	if normPark == "glac" || strings.Contains(normPark, "glacier") || (normState == "MT" && normPark == "") {
		c := ReferenceGlacierCorridor()
		return &c
	}
	// Rocky Mountain (CO)
	if normPark == "romo" || strings.Contains(normPark, "rocky") || (normState == "CO" && normPark == "") {
		c := ReferenceRockyMountainCorridor()
		return &c
	}
	// Yellowstone (WY)
	if normPark == "yell" || strings.Contains(normPark, "yellowstone") || (normState == "WY" && normPark == "") {
		c := ReferenceYellowstoneCorridor()
		return &c
	}
	return nil
}
