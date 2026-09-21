package api

import (
	"context"
	"errors"
)

func (s service) Health(context.Context, HealthRequestObject) (HealthResponseObject, error) {
	return Health200JSONResponse{Status: HealthEnvelopeStatusSuccess, Data: HealthData{Available: true}}, nil
}
func (s service) Ready(ctx context.Context, _ ReadyRequestObject) (ReadyResponseObject, error) {
	if s.deps.Ready == nil {
		return Ready503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Dependencies unavailable"})}, nil
	}
	if err := s.deps.Ready(ctx); err != nil {
		s.logFailure(ctx, "readiness")
		return Ready503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{Status: Error, Error: "Dependencies unavailable"})}, nil
	}
	return Ready200JSONResponse{Status: HealthEnvelopeStatusSuccess, Data: HealthData{Available: true}}, nil
}
func (s service) logFailure(ctx context.Context, operation string) {
	if s.deps.Logger != nil {
		s.deps.Logger.ErrorContext(ctx, "dependency operation failed", "operation", operation)
	}
}

var errNoStore = errors.New("store unavailable")
