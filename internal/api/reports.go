package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"strings"

	"corridor/internal/db"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
)

// FieldReportRecord is an alias for the database-level report model.
type FieldReportRecord = db.FieldReportRecord

type defaultClassifier struct{}

func (defaultClassifier) Classify(ctx context.Context, imgBytes []byte) ([]SpeciesSuggestion, error) {
	if len(imgBytes) == 0 {
		return nil, errors.New("empty image bytes")
	}
	_, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid image format: %w", err)
	}

	return []SpeciesSuggestion{
		{Species: "mule_deer", Confidence: 0.88},
		{Species: "elk", Confidence: 0.08},
		{Species: "pronghorn", Confidence: 0.04},
	}, nil
}

// StripEXIF removes all EXIF and GPS metadata from image bytes in pure Go.
// It strips JPEG APP1 segments directly to maintain byte fidelity, or decodes
// and re-encodes to a clean JPEG if necessary.
func StripEXIF(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid or empty image data")
	}

	// For JPEG: strip APP1 segment containing EXIF / GPS
	if data[0] == 0xFF && data[1] == 0xD8 {
		stripped, err := stripJPEGAPP1(data)
		if err == nil {
			_, _, decErr := image.Decode(bytes.NewReader(stripped))
			if decErr == nil {
				return stripped, nil
			}
		}
	}

	// Fallback for other formats or non-standard JPEG streams: decode and re-encode to clean JPEG
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

func stripJPEGAPP1(data []byte) ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte(0xFF)
	out.WriteByte(0xD8)

	offset := 2
	for offset < len(data) {
		if data[offset] != 0xFF {
			return nil, errors.New("corrupt JPEG marker")
		}
		for offset < len(data) && data[offset] == 0xFF {
			offset++
		}
		if offset >= len(data) {
			break
		}
		marker := data[offset]
		offset++

		// End of image
		if marker == 0xD9 {
			out.WriteByte(0xFF)
			out.WriteByte(marker)
			break
		}

		// Start of scan: copy remainder of image data
		if marker == 0xDA {
			out.WriteByte(0xFF)
			out.WriteByte(marker)
			out.Write(data[offset:])
			break
		}

		// Standalone markers without payload
		if (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 {
			out.WriteByte(0xFF)
			out.WriteByte(marker)
			continue
		}

		if offset+2 > len(data) {
			return nil, errors.New("truncated marker length")
		}
		length := int(data[offset])<<8 | int(data[offset+1])
		if length < 2 || offset+length > len(data) {
			return nil, errors.New("invalid marker length")
		}

		// Strip APP1 marker (0xE1: EXIF / GPS)
		if marker == 0xE1 {
			offset += length
			continue
		}

		out.WriteByte(0xFF)
		out.WriteByte(marker)
		out.Write(data[offset : offset+length])
		offset += length
	}

	return out.Bytes(), nil
}

func extractPhotoBytes(req UploadReportPhotoRequestObject) ([]byte, error) {
	if req.Body != nil {
		data, err := io.ReadAll(io.LimitReader(req.Body, 32<<20))
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			return data, nil
		}
	}
	if req.MultipartBody != nil {
		for {
			part, err := req.MultipartBody.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return nil, err
			}
			if part.FormName() == "photo" || part.FormName() == "file" || part.FileName() != "" {
				data, err := io.ReadAll(io.LimitReader(part, 32<<20))
				_ = part.Close()
				if err != nil {
					return nil, err
				}
				if len(data) > 0 {
					return data, nil
				}
			}
			_ = part.Close()
		}
	}
	return nil, errors.New("no photo content provided")
}

func toAPIFieldReport(r FieldReportRecord) FieldReport {
	res := FieldReport{
		Confidence:        r.Confidence,
		CreatedAt:         r.CreatedAt,
		Id:                r.ID,
		IsDuplicate:       r.IsDuplicate,
		Latitude:          r.Latitude,
		Longitude:         r.Longitude,
		ObservedAt:        r.ObservedAt,
		SpeciesCommonName: r.SpeciesCommonName,
		Notes:             r.Notes,
		PhotoKey:          r.PhotoKey,
		SegmentId:         r.SegmentID,
		SnapDistanceM:     r.SnapDistanceM,
		SnappedLatitude:   r.SnappedLatitude,
		SnappedLongitude:  r.SnappedLongitude,
		DedupGroupId:      r.DedupGroupID,
	}
	if len(r.SpeciesSuggestions) > 0 {
		suggs := make([]SpeciesSuggestion, len(r.SpeciesSuggestions))
		for i, s := range r.SpeciesSuggestions {
			suggs[i] = SpeciesSuggestion{
				Confidence: s.Confidence,
				Species:    s.Species,
			}
		}
		res.SpeciesSuggestions = &suggs
	}
	return res
}

func (s service) CreateReport(ctx context.Context, req CreateReportRequestObject) (CreateReportResponseObject, error) {
	if req.Body == nil {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Request body is required",
		})}, nil
	}

	b := req.Body
	if b.Id == uuid.Nil {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "id must be a valid non-empty UUID",
		})}, nil
	}
	if strings.TrimSpace(b.SpeciesCommonName) == "" {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "species_common_name is required",
		})}, nil
	}
	if math.IsNaN(b.Latitude) || math.IsInf(b.Latitude, 0) || b.Latitude < -90 || b.Latitude > 90 {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "latitude must be between -90 and 90",
		})}, nil
	}
	if math.IsNaN(b.Longitude) || math.IsInf(b.Longitude, 0) || b.Longitude < -180 || b.Longitude > 180 {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "longitude must be between -180 and 180",
		})}, nil
	}
	if b.ObservedAt.IsZero() {
		return CreateReport400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "observed_at must be non-zero",
		})}, nil
	}

	store := s.reportStore()
	if store == nil {
		s.logFailure(ctx, "report_store_nil")
		return CreateReport503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Field reporting service unavailable",
		})}, nil
	}

	// Idempotency check: return existing report if previously recorded
	existing, err := store.GetReport(ctx, b.Id)
	if err == nil && existing != nil {
		return CreateReport200JSONResponse{
			Status: ReportEnvelopeStatusSuccess,
			Data:   toAPIFieldReport(*existing),
			Error:  nil,
			Meta:   nil,
		}, nil
	}

	conf := "unconfirmed"
	if b.Confidence != nil && strings.TrimSpace(*b.Confidence) != "" {
		conf = strings.TrimSpace(*b.Confidence)
	}

	record := FieldReportRecord{
		ID:                b.Id,
		SpeciesCommonName: strings.TrimSpace(b.SpeciesCommonName),
		Latitude:          b.Latitude,
		Longitude:         b.Longitude,
		ObservedAt:        b.ObservedAt,
		Confidence:        conf,
		Notes:             b.Notes,
	}

	// Snap coordinates to nearest 0.1-mile road segment (50m tolerance)
	snapper := s.snapper()
	if snapper != nil {
		segID, snapDist, snappedPt, err := snapper(ctx, orb.Point{b.Longitude, b.Latitude}, 50.0)
		if err == nil && segID != uuid.Nil {
			record.SegmentID = &segID
			record.SnapDistanceM = &snapDist
			lat := snappedPt[1]
			lon := snappedPt[0]
			record.SnappedLatitude = &lat
			record.SnappedLongitude = &lon
		}
	}

	// Deduplicate against recent nearby carcass reports (200m / 72h)
	deduplicator := s.deduplicator()
	if deduplicator != nil {
		dedupGroup, isDup, err := deduplicator(ctx, record.SpeciesCommonName, orb.Point{b.Longitude, b.Latitude}, b.ObservedAt, 200.0, 72.0)
		if err == nil && dedupGroup != nil {
			record.DedupGroupID = dedupGroup
			record.IsDuplicate = isDup
		}
	}
	if record.DedupGroupID == nil {
		newGroup := uuid.New()
		record.DedupGroupID = &newGroup
		record.IsDuplicate = false
	}

	createdRecord, isNew, err := store.CreateReport(ctx, record)
	if err != nil {
		s.logFailure(ctx, "create_report_failed")
		return CreateReport503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to record field carcass report",
		})}, nil
	}

	if !isNew {
		return CreateReport200JSONResponse{
			Status: ReportEnvelopeStatusSuccess,
			Data:   toAPIFieldReport(*createdRecord),
			Error:  nil,
			Meta:   nil,
		}, nil
	}

	return CreateReport201JSONResponse{
		Status: ReportEnvelopeStatusSuccess,
		Data:   toAPIFieldReport(*createdRecord),
		Error:  nil,
		Meta:   nil,
	}, nil
}

func (s service) UploadReportPhoto(ctx context.Context, req UploadReportPhotoRequestObject) (UploadReportPhotoResponseObject, error) {
	if req.Id == uuid.Nil {
		return UploadReportPhoto400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "id must be a valid non-empty UUID",
		})}, nil
	}

	store := s.reportStore()
	if store == nil {
		s.logFailure(ctx, "report_store_nil")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Field reporting service unavailable",
		})}, nil
	}

	_, err := store.GetReport(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return UploadReportPhoto404JSONResponse{NotFoundJSONResponse(ErrorEnvelope{
				Status: Error,
				Error:  "Field report not found",
			})}, nil
		}
		s.logFailure(ctx, "get_report_failed")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to check field report existence",
		})}, nil
	}

	rawBytes, err := extractPhotoBytes(req)
	if err != nil || len(rawBytes) == 0 {
		return UploadReportPhoto400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Invalid or empty photo data",
		})}, nil
	}

	cleanBytes, err := StripEXIF(rawBytes)
	if err != nil {
		return UploadReportPhoto400JSONResponse{BadRequestJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to process image: " + err.Error(),
		})}, nil
	}

	photoKey := fmt.Sprintf("corridor-raw/photos/%s.jpg", req.Id)
	pStore := s.photoStore()
	if pStore == nil {
		s.logFailure(ctx, "photo_store_nil")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Photo storage service unavailable",
		})}, nil
	}

	if err := pStore.PutPhoto(ctx, photoKey, cleanBytes); err != nil {
		s.logFailure(ctx, "store_photo_failed")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to store photo in private storage",
		})}, nil
	}

	classifier := s.speciesClassifier()
	suggestions, err := classifier.Classify(ctx, cleanBytes)
	if err != nil {
		s.logFailure(ctx, "species_inference_failed")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Species classification inference failed",
		})}, nil
	}

	_, err = store.UpdateReportPhoto(ctx, req.Id, photoKey, suggestions)
	if err != nil {
		s.logFailure(ctx, "update_report_photo_failed")
		return UploadReportPhoto503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Failed to update report with photo metadata",
		})}, nil
	}

	return UploadReportPhoto200JSONResponse{
		Status: PhotoUploadEnvelopeStatusSuccess,
		Data: PhotoUploadData{
			PhotoKey:           photoKey,
			ReportId:           req.Id,
			SpeciesSuggestions: suggestions,
		},
		Error: nil,
		Meta:  nil,
	}, nil
}

type dbReportStoreAdapter struct {
	store db.Store
}

func (a dbReportStoreAdapter) CreateReport(ctx context.Context, report FieldReportRecord) (*FieldReportRecord, bool, error) {
	return a.store.CreateReport(ctx, report)
}

func (a dbReportStoreAdapter) GetReport(ctx context.Context, id uuid.UUID) (*FieldReportRecord, error) {
	return a.store.GetReport(ctx, id)
}

func (a dbReportStoreAdapter) UpdateReportPhoto(ctx context.Context, id uuid.UUID, photoKey string, suggestions []SpeciesSuggestion) (*FieldReportRecord, error) {
	dbSuggs := make([]db.SpeciesSuggestion, len(suggestions))
	for i, s := range suggestions {
		dbSuggs[i] = db.SpeciesSuggestion{
			Species:    s.Species,
			Confidence: s.Confidence,
		}
	}
	return a.store.UpdateReportPhoto(ctx, id, photoKey, dbSuggs)
}
