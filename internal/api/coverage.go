package api

import "context"

func (s service) Coverage(ctx context.Context, _ CoverageRequestObject) (CoverageResponseObject, error) {
	var count int64
	err := errNoStore
	if s.deps.Store != nil {
		count, err = s.deps.Store.EventCount(ctx)
	}
	if err != nil || count < 0 {
		s.logFailure(ctx, "coverage")
		return Coverage503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Data unavailable"})}, nil
	}
	state := Empty
	if count > 0 {
		state = Unmodeled
	}
	return Coverage200JSONResponse{Status: CoverageEnvelopeStatusSuccess, Data: CoverageData{State: state, IngestedEvents: count, ModelAvailable: False}}, nil
}
func (s service) Sources(ctx context.Context, req SourcesRequestObject) (SourcesResponseObject, error) {
	limit, offset := 25, 0
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return Sources400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{Status: Error, Error: "Pagination out of range"})}, nil
	}
	unavailable := Sources503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Data unavailable"})}
	if s.deps.Store == nil {
		return unavailable, nil
	}
	rows, total, err := s.deps.Store.Sources(ctx, int32(limit), int32(offset))
	if err != nil {
		s.logFailure(ctx, "sources")
		return unavailable, nil
	}
	sources := make([]SourceSummary, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, SourceSummary{Id: row.ID, Name: row.Name, Url: row.URL, License: row.License, Status: row.Status})
	}
	return Sources200JSONResponse{Status: SourcesEnvelopeStatusSuccess, Data: sources, Meta: Page{Total: total, Limit: limit, Offset: offset}}, nil
}
