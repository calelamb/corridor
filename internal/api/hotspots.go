package api

import (
	"context"
	"fmt"
	"math"
	"sort"

	"corridor/internal/db"
	"corridor/internal/hotspot"
)

func (s service) Hotspots(ctx context.Context, req HotspotsRequestObject) (HotspotsResponseObject, error) {
	if req.Params.Method != nil && !req.Params.Method.Valid() {
		return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "method must be nkde, getis_ord, or density",
		})}, nil
	}

	bw := 500.0
	if req.Params.Bandwidth != nil {
		bw = *req.Params.Bandwidth
		if math.IsNaN(bw) || math.IsInf(bw, 0) || bw <= 0 {
			return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "bandwidth must be a positive number",
			})}, nil
		}
	}

	var bboxVals *[4]float64
	if req.Params.Bbox != nil {
		b, err := parseBBox(*req.Params.Bbox)
		if err != nil {
			return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  err.Error(),
			})}, nil
		}
		bboxVals = b
	}

	methodStr := "nkde"
	if req.Params.Method != nil {
		methodStr = string(*req.Params.Method)
	}

	// 1. Delegated hotspot provider if configured
	if hp := s.hotspotProvider(); hp != nil {
		items, err := hp.Hotspots(ctx, req.Params)
		if err != nil {
			s.logFailure(ctx, "hotspots_provider")
			return Hotspots503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "Hotspot analysis service unavailable",
			})}, nil
		}
		if items == nil {
			items = []HotspotItem{}
		}
		return Hotspots200JSONResponse(HotspotsEnvelope{
			Status: HotspotsEnvelopeStatusSuccess,
			Data:   items,
		}), nil
	}

	// 2. Database segments if SegmentStore is available
	var dbSegments []db.SegmentSummary
	if store := s.segmentStore(); store != nil {
		var filter db.SegmentFilter
		filter.BBox = bboxVals
		if segs, err := store.Segments(ctx, filter); err == nil && len(segs) > 0 {
			dbSegments = segs
		}
	}

	if len(dbSegments) > 0 {
		items, err := computeHotspotsFromDBSegments(dbSegments, methodStr, bw)
		if err != nil {
			s.logFailure(ctx, "hotspots_db_compute")
			return Hotspots503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "Hotspot analysis calculation failed",
			})}, nil
		}
		return Hotspots200JSONResponse(HotspotsEnvelope{
			Status: HotspotsEnvelopeStatusSuccess,
			Data:   items,
		}), nil
	}

	// 3. Fallback to reference corridors
	items := computeHotspotsFromReferenceCorridors(bboxVals, methodStr, bw)

	return Hotspots200JSONResponse(HotspotsEnvelope{
		Status: HotspotsEnvelopeStatusSuccess,
		Data:   items,
	}), nil
}

func computeHotspotsFromReferenceCorridors(bboxVals *[4]float64, method string, bandwidth float64) []HotspotItem {
	corridors := hotspot.AllReferenceCorridors()
	var allItems []HotspotItem

	for _, c := range corridors {
		if bboxVals != nil {
			if !hotspot.CorridorOverlapsBBox(c.BBox, *bboxVals) {
				continue
			}
		}

		results, err := c.Graph.AnalyzeHotspots(c.Events, method, bandwidth)
		if err != nil {
			continue
		}

		for _, r := range results {
			if bboxVals != nil {
				geo, ok := c.SegmentGeos[r.SegmentID]
				if !ok || !hotspot.SegmentInBBox(geo, *bboxVals) {
					continue
				}
			}

			sig := r.Significance
			m := method
			item := HotspotItem{
				SegmentId:        r.SegmentID,
				Score:            r.Score,
				RawCount:         &r.RawCount,
				ZScore:           r.ZScore,
				PValue:           r.PValue,
				FdrPValue:        r.FDRPValue,
				Significance:     &sig,
				SignificanceTier: &sig,
				Method:           &m,
			}
			allItems = append(allItems, item)
		}
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Score > allItems[j].Score
	})

	if allItems == nil {
		allItems = []HotspotItem{}
	}
	return allItems
}

func computeHotspotsFromDBSegments(segs []db.SegmentSummary, method string, bandwidth float64) ([]HotspotItem, error) {
	g := hotspot.NewRoadGraph()
	var events []hotspot.EventPoint

	const segLengthM = 160.934
	for i, seg := range segs {
		id := seg.ID.String()
		if err := g.AddSegment(id, segLengthM); err != nil {
			return nil, err
		}
		if i > 0 {
			prevID := segs[i-1].ID.String()
			_ = g.AddEdge(prevID, id, segLengthM)
		}

		count := int(math.Round(seg.RiskScore * 5))
		if count == 0 && seg.RiskScore > 0 {
			count = 1
		}
		for j := 0; j < count; j++ {
			events = append(events, hotspot.EventPoint{
				ID:                    fmt.Sprintf("ev-%s-%d", id, j),
				SegmentID:             id,
				DistanceAlongSegmentM: segLengthM / 2.0,
				Weight:                seg.RiskScore,
			})
		}
	}

	results, err := g.AnalyzeHotspots(events, method, bandwidth)
	if err != nil {
		return nil, err
	}

	items := make([]HotspotItem, len(results))
	for i, r := range results {
		sig := r.Significance
		m := method
		items[i] = HotspotItem{
			SegmentId:        r.SegmentID,
			Score:            r.Score,
			RawCount:         &r.RawCount,
			ZScore:           r.ZScore,
			PValue:           r.PValue,
			FdrPValue:        r.FDRPValue,
			Significance:     &sig,
			SignificanceTier: &sig,
			Method:           &m,
		}
	}
	return items, nil
}
