package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"corridor/internal/export"

	"github.com/google/uuid"
)

// MitigationBriefPDF200Response writes the PDF content with proper headers including Content-Disposition.
type MitigationBriefPDF200Response struct {
	PDFBytes []byte
	SiteID   string
}

// VisitMitigationBriefPDFResponse implements MitigationBriefPDFResponseObject.
func (r MitigationBriefPDF200Response) VisitMitigationBriefPDFResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="funding-brief-%s.pdf"`, r.SiteID))
	w.Header().Set("Content-Length", strconv.Itoa(len(r.PDFBytes)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(r.PDFBytes)
	return err
}

func (s service) MitigationBriefPDF(ctx context.Context, req MitigationBriefPDFRequestObject) (MitigationBriefPDFResponseObject, error) {
	if req.Id == uuid.Nil {
		return MitigationBriefPDF400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "id must be a valid non-empty UUID",
		})}, nil
	}

	provider := s.mitigationProvider()
	if provider == nil {
		s.logFailure(ctx, "mitigation_provider_nil")
		return MitigationBriefPDF503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Mitigation prioritization service unavailable",
		})}, nil
	}

	site, err := provider.CandidateSite(ctx, req.Id)
	if err != nil {
		s.logFailure(ctx, "mitigation_candidate_site_lookup")
		return MitigationBriefPDF503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to retrieve mitigation candidate site",
		})}, nil
	}

	if site == nil {
		return MitigationBriefPDF404JSONResponse{NotFoundJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Mitigation candidate site not found",
		})}, nil
	}

	pdfBytes, err := export.GenerateFundingBriefPDF(*site, nil)
	if err != nil {
		s.logFailure(ctx, "pdf_generation_failed")
		return MitigationBriefPDF503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to generate funding brief PDF",
		})}, nil
	}

	return MitigationBriefPDF200Response{
		PDFBytes: pdfBytes,
		SiteID:   req.Id.String(),
	}, nil
}
