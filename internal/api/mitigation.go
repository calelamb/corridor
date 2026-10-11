package api

import (
	"context"
	"math"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s service) Mitigation(ctx context.Context, req MitigationRequestObject) (MitigationResponseObject, error) {
	budget := -1.0
	if req.Params.Budget != nil {
		b := *req.Params.Budget
		if math.IsNaN(b) || math.IsInf(b, 0) || b < 0 {
			return Mitigation400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "budget must be a non-negative finite number",
			})}, nil
		}
		budget = b
	}

	var minBCR *float64
	if req.Params.MinBcr != nil {
		m := *req.Params.MinBcr
		if math.IsNaN(m) || math.IsInf(m, 0) || m < 0 {
			return Mitigation400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "min_bcr must be a non-negative finite number",
			})}, nil
		}
		minBCR = &m
	}

	var state *string
	if req.Params.State != nil {
		st := strings.ToUpper(strings.TrimSpace(*req.Params.State))
		if len(st) != 2 && len(st) != 0 {
			return Mitigation400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "state must be a 2-letter state code",
			})}, nil
		}
		if st != "" {
			state = &st
		}
	}

	provider := s.mitigationProvider()
	if provider == nil {
		s.logFailure(ctx, "mitigation_provider_nil")
		return Mitigation503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Mitigation prioritization service unavailable",
		})}, nil
	}

	candidates, err := provider.GenerateCandidates(ctx, budget, minBCR, state)
	if err != nil {
		s.logFailure(ctx, "mitigation_generate_candidates")
		return Mitigation503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to generate mitigation candidates",
		})}, nil
	}

	items := make([]CandidateSite, 0, len(candidates))
	for _, c := range candidates {
		segUUIDs := make([]openapi_types.UUID, len(c.Segments))
		for j, id := range c.Segments {
			segUUIDs[j] = id
		}

		items = append(items, CandidateSite{
			Id:                        c.ID,
			Segments:                  segUUIDs,
			State:                     c.State,
			HighwayRef:                c.HighwayRef,
			MilepostStart:             c.MilepostStart,
			MilepostEnd:               c.MilepostEnd,
			LengthM:                   c.LengthM,
			RecommendedType:           c.RecommendedType,
			EstCostUsd:                c.EstCostUSD,
			ExpectedAvoidedCollisions: c.ExpectedAvoidedCollisions,
			ExpectedSavingsUsd:        c.ExpectedSavingsUSD,
			Bcr:                       c.BCR,
			P10Bcr:                    c.P10BCR,
			P90Bcr:                    c.P90BCR,
			Rationale:                 c.Rationale,
		})
	}

	return Mitigation200JSONResponse(MitigationEnvelope{
		Status: MitigationEnvelopeStatusSuccess,
		Data:   items,
		Error:  nil,
		Meta:   nil,
	}), nil
}
