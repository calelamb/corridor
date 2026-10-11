package api

import (
	"context"
	"math"
)

func (s service) Hotspots(ctx context.Context, req HotspotsRequestObject) (HotspotsResponseObject, error) {
	if req.Params.Method != nil && !req.Params.Method.Valid() {
		return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "method must be nkde, getis_ord, or density",
		})}, nil
	}

	if req.Params.Bandwidth != nil {
		bw := *req.Params.Bandwidth
		if math.IsNaN(bw) || math.IsInf(bw, 0) || bw <= 0 {
			return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "bandwidth must be a positive number",
			})}, nil
		}
	}

	if req.Params.Bbox != nil {
		_, err := parseBBox(*req.Params.Bbox)
		if err != nil {
			return Hotspots400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  err.Error(),
			})}, nil
		}
	}

	return Hotspots200JSONResponse(HotspotsEnvelope{
		Status: HotspotsEnvelopeStatusSuccess,
		Data:   []HotspotItem{},
	}), nil
}
