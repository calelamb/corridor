package api

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"

	"corridor/internal/db"
)

func parseBBox(raw string) (*[4]float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, errors.New("bounding box must contain 4 comma-separated values: west,south,east,north")
	}
	var vals [4]float64
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("invalid float in bounding box")
		}
		vals[i] = v
	}
	west, south, east, north := vals[0], vals[1], vals[2], vals[3]
	if west < -180 || west > 180 || east < -180 || east > 180 {
		return nil, errors.New("longitude must be between -180 and 180")
	}
	if south < -90 || south > 90 || north < -90 || north > 90 {
		return nil, errors.New("latitude must be between -90 and 90")
	}
	if west >= east {
		return nil, errors.New("west bound must be strictly less than east bound")
	}
	if south >= north {
		return nil, errors.New("south bound must be strictly less than north bound")
	}
	return &vals, nil
}

func (s service) Segments(ctx context.Context, req SegmentsRequestObject) (SegmentsResponseObject, error) {
	var filter db.SegmentFilter

	if req.Params.Bbox != nil {
		bbox, err := parseBBox(*req.Params.Bbox)
		if err != nil {
			return Segments400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{Status: Error, Error: err.Error()})}, nil
		}
		filter.BBox = bbox
	}

	if req.Params.MinRisk != nil {
		if math.IsNaN(*req.Params.MinRisk) || math.IsInf(*req.Params.MinRisk, 0) || *req.Params.MinRisk < 0 {
			return Segments400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{Status: Error, Error: "min_risk must be a non-negative finite number"})}, nil
		}
		filter.MinRisk = req.Params.MinRisk
	}

	store := s.segmentStore()
	if store == nil {
		s.logFailure(ctx, "segments_store_nil")
		return Segments503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment data unavailable"})}, nil
	}

	rows, err := store.Segments(ctx, filter)
	if err != nil {
		s.logFailure(ctx, "segments_query")
		return Segments503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment data unavailable"})}, nil
	}

	items := make([]SegmentSummary, 0, len(rows))
	for _, row := range rows {
		coords := row.Coordinates
		if coords == nil {
			coords = [][]float64{}
		}
		items = append(items, SegmentSummary{
			Id:             row.ID,
			RiskScore:      row.RiskScore,
			PercentileRank: row.PercentileRank,
			P10:            row.P10,
			P90:            row.P90,
			HighwayRef:     row.HighwayRef,
			State:          row.State,
			Geometry: LineStringGeometry{
				Type:        LineString,
				Coordinates: coords,
			},
		})
	}

	return Segments200JSONResponse{
		Status: SegmentsEnvelopeStatusSuccess,
		Data:   items,
		Error:  nil,
		Meta:   nil,
	}, nil
}

func (s service) SegmentDetail(ctx context.Context, req SegmentDetailRequestObject) (SegmentDetailResponseObject, error) {
	store := s.segmentStore()
	if store == nil {
		s.logFailure(ctx, "segment_detail_store_nil")
		return SegmentDetail503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment detail unavailable"})}, nil
	}

	detail, err := store.Segment(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return SegmentDetail404JSONResponse{NotFoundJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment not found"})}, nil
		}
		s.logFailure(ctx, "segment_detail_query")
		return SegmentDetail503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment detail unavailable"})}, nil
	}
	if detail == nil {
		return SegmentDetail404JSONResponse{NotFoundJSONResponse(ErrorEnvelope{Status: Error, Error: "Segment not found"})}, nil
	}

	drivers := make([]ShapDriver, 0, len(detail.ShapDrivers))
	for _, d := range detail.ShapDrivers {
		drivers = append(drivers, ShapDriver{
			Feature: d.Feature,
			Impact:  d.Impact,
			Weight:  d.Weight,
		})
	}

	monthly := detail.MonthlyProfile
	if monthly == nil {
		monthly = make(map[string]float64)
	}

	species := detail.SpeciesBreakdown
	if species == nil {
		species = make(map[string]float64)
	}

	coords := detail.Coordinates
	if coords == nil {
		coords = [][]float64{}
	}

	data := SegmentDetail{
		Id:               detail.ID,
		RiskScore:        detail.RiskScore,
		PercentileRank:   detail.PercentileRank,
		P10:              detail.P10,
		P90:              detail.P90,
		HighwayRef:       detail.HighwayRef,
		State:            detail.State,
		MonthlyProfile:   monthly,
		SpeciesBreakdown: species,
		ShapDrivers:      drivers,
		Geometry: LineStringGeometry{
			Type:        LineString,
			Coordinates: coords,
		},
	}

	return SegmentDetail200JSONResponse{
		Status: SegmentDetailEnvelopeStatusSuccess,
		Data:   data,
		Error:  nil,
		Meta:   nil,
	}, nil
}
