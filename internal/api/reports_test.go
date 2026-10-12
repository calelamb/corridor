package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"

	"corridor/internal/db"
	"corridor/internal/geo"
)

type mockReportStore struct {
	mu      sync.Mutex
	reports map[uuid.UUID]*FieldReportRecord
	err     error
}

func newMockReportStore() *mockReportStore {
	return &mockReportStore{
		reports: make(map[uuid.UUID]*FieldReportRecord),
	}
}

func (m *mockReportStore) CreateReport(ctx context.Context, report FieldReportRecord) (*FieldReportRecord, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, false, m.err
	}
	if existing, ok := m.reports[report.ID]; ok {
		return existing, false, nil
	}
	r := report
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	m.reports[r.ID] = &r
	return &r, true, nil
}

func (m *mockReportStore) GetReport(ctx context.Context, id uuid.UUID) (*FieldReportRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	r, ok := m.reports[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (m *mockReportStore) UpdateReportPhoto(ctx context.Context, id uuid.UUID, photoKey string, suggestions []SpeciesSuggestion) (*FieldReportRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	r, ok := m.reports[id]
	if !ok {
		return nil, errors.New("not found")
	}
	r.PhotoKey = &photoKey
	dbSuggs := make([]db.SpeciesSuggestion, len(suggestions))
	for i, s := range suggestions {
		dbSuggs[i] = db.SpeciesSuggestion{
			Species:    s.Species,
			Confidence: s.Confidence,
		}
	}
	r.SpeciesSuggestions = dbSuggs
	return r, nil
}

type mockPhotoStore struct {
	mu     sync.Mutex
	photos map[string][]byte
	err    error
}

func newMockPhotoStore() *mockPhotoStore {
	return &mockPhotoStore{
		photos: make(map[string][]byte),
	}
}

func (m *mockPhotoStore) PutPhoto(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	cpy := make([]byte, len(data))
	copy(cpy, data)
	m.photos[key] = cpy
	return nil
}

func (m *mockPhotoStore) GetPhoto(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.photos[key]
	if !ok {
		return nil, errors.New("photo not found")
	}
	return data, nil
}

// createTestJPEGWithEXIF generates a valid JPEG containing an APP1 EXIF segment with GPS data.
func createTestJPEGWithEXIF(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var rawBuf bytes.Buffer
	if err := jpeg.Encode(&rawBuf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("failed to encode test jpeg: %v", err)
	}
	raw := rawBuf.Bytes()
	if len(raw) < 2 || raw[0] != 0xFF || raw[1] != 0xD8 {
		t.Fatal("invalid test jpeg header")
	}

	// Craft APP1 marker: 0xFF, 0xE1, length (2 bytes), "Exif\x00\x00", followed by dummy GPS data
	app1Payload := []byte("Exif\x00\x00GPS:LAT=43.8,LON=-110.2,ALT=2000m")
	app1Len := len(app1Payload) + 2
	var app1 bytes.Buffer
	app1.WriteByte(0xFF)
	app1.WriteByte(0xE1)
	app1.WriteByte(byte(app1Len >> 8))
	app1.WriteByte(byte(app1Len & 0xFF))
	app1.Write(app1Payload)

	// Insert APP1 right after SOI (0xFF, 0xD8)
	var withEXIF bytes.Buffer
	withEXIF.Write(raw[:2])
	withEXIF.Write(app1.Bytes())
	withEXIF.Write(raw[2:])

	return withEXIF.Bytes()
}

func TestFieldReportsEndpoint(t *testing.T) {
	reportStore := newMockReportStore()
	photoStore := newMockPhotoStore()

	targetSegmentID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	var knownReports []FieldReportRecord

	snapper := func(ctx context.Context, pt orb.Point, maxDist float64) (uuid.UUID, float64, orb.Point, error) {
		// If within 50m of test segment coordinate (-110.20, 43.80)
		if pt[0] >= -110.21 && pt[0] <= -110.19 && pt[1] >= 43.79 && pt[1] <= 43.81 {
			return targetSegmentID, 12.5, orb.Point{-110.2001, 43.8001}, nil
		}
		return uuid.Nil, 0, orb.Point{}, geo.ErrExceedsSnapTolerance
	}

	deduplicator := func(ctx context.Context, species string, pt orb.Point, obs time.Time, distM, hours float64) (*uuid.UUID, bool, error) {
		for _, r := range knownReports {
			if r.SpeciesCommonName == species && mathAbs(r.Latitude-pt[1]) < 0.002 && mathAbs(r.Longitude-pt[0]) < 0.002 {
				return r.DedupGroupID, true, nil
			}
		}
		newGroup := uuid.New()
		return &newGroup, false, nil
	}

	h := NewRouter(Dependencies{
		ReportStore:  reportStore,
		PhotoStore:   photoStore,
		Snapper:      snapper,
		Deduplicator: deduplicator,
		Ready:        func(context.Context) error { return nil },
	})

	reportID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	obsTime := time.Date(2026, 10, 11, 14, 30, 0, 0, time.UTC)

	t.Run("POST /v1/reports creates new report with segment snapping and dedup group", func(t *testing.T) {
		payload := map[string]any{
			"id":                  reportID.String(),
			"species_common_name": "mule_deer",
			"latitude":            43.8000,
			"longitude":           -110.2000,
			"observed_at":         obsTime.Format(time.RFC3339),
			"confidence":          "confirmed",
			"notes":               "Adult male carcass near milepost 142",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var res struct {
			Status string      `json:"status"`
			Data   FieldReport `json:"data"`
			Error  any         `json:"error"`
			Meta   any         `json:"meta"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.Status != "success" {
			t.Fatalf("expected status success, got %q", res.Status)
		}
		if res.Data.Id != reportID {
			t.Fatalf("expected report id %s, got %s", reportID, res.Data.Id)
		}
		if res.Data.SpeciesCommonName != "mule_deer" {
			t.Fatalf("expected species mule_deer, got %s", res.Data.SpeciesCommonName)
		}
		if res.Data.SegmentId == nil || *res.Data.SegmentId != targetSegmentID {
			t.Fatalf("expected segment id %s, got %v", targetSegmentID, res.Data.SegmentId)
		}
		if res.Data.SnapDistanceM == nil || *res.Data.SnapDistanceM <= 0 {
			t.Fatalf("expected positive snap distance, got %v", res.Data.SnapDistanceM)
		}
		if res.Data.SnappedLatitude == nil || res.Data.SnappedLongitude == nil {
			t.Fatal("expected non-nil snapped coordinates")
		}
		if res.Data.IsDuplicate {
			t.Fatal("expected first report to not be duplicate")
		}
		if res.Data.DedupGroupId == nil {
			t.Fatal("expected non-nil dedup group id")
		}

		// Save to known reports for subsequent dedup check
		knownReports = append(knownReports, FieldReportRecord{
			ID:                res.Data.Id,
			SpeciesCommonName: res.Data.SpeciesCommonName,
			Latitude:          res.Data.Latitude,
			Longitude:         res.Data.Longitude,
			DedupGroupID:      res.Data.DedupGroupId,
		})
	})

	t.Run("POST /v1/reports is idempotent with identical client UUID", func(t *testing.T) {
		payload := map[string]any{
			"id":                  reportID.String(),
			"species_common_name": "mule_deer",
			"latitude":            43.8000,
			"longitude":           -110.2000,
			"observed_at":         obsTime.Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Fatalf("expected 200 or 201 for idempotent request, got %d: %s", rec.Code, rec.Body.String())
		}

		var res struct {
			Status string      `json:"status"`
			Data   FieldReport `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if res.Data.Id != reportID {
			t.Fatalf("expected report id %s, got %s", reportID, res.Data.Id)
		}
	})

	t.Run("POST /v1/reports deduplicates nearby observation within 200m / 72h", func(t *testing.T) {
		dupID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		payload := map[string]any{
			"id":                  dupID.String(),
			"species_common_name": "mule_deer",
			"latitude":            43.8001,
			"longitude":           -110.2001,
			"observed_at":         obsTime.Add(2 * time.Hour).Format(time.RFC3339),
			"confidence":          "unconfirmed",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var res struct {
			Status string      `json:"status"`
			Data   FieldReport `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !res.Data.IsDuplicate {
			t.Fatal("expected report to be flagged as duplicate")
		}
		if res.Data.DedupGroupId == nil || *res.Data.DedupGroupId != *knownReports[0].DedupGroupID {
			t.Fatalf("expected dedup group id %v, got %v", knownReports[0].DedupGroupID, res.Data.DedupGroupId)
		}
	})

	t.Run("POST /v1/reports beyond snapping tolerance leaves segment_id nil", func(t *testing.T) {
		farID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		payload := map[string]any{
			"id":                  farID.String(),
			"species_common_name": "elk",
			"latitude":            44.5000,
			"longitude":           -111.5000,
			"observed_at":         obsTime.Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var res struct {
			Data FieldReport `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if res.Data.SegmentId != nil {
			t.Fatalf("expected nil segment_id for point beyond tolerance, got %v", res.Data.SegmentId)
		}
	})

	t.Run("POST /v1/reports rejects invalid coordinates with 400", func(t *testing.T) {
		for _, coords := range [][2]float64{
			{95.0, -110.0},  // lat > 90
			{-95.0, -110.0}, // lat < -90
			{45.0, 190.0},   // lon > 180
			{45.0, -190.0},  // lon < -180
		} {
			badID := uuid.New()
			payload := map[string]any{
				"id":                  badID.String(),
				"species_common_name": "mule_deer",
				"latitude":            coords[0],
				"longitude":           coords[1],
				"observed_at":         obsTime.Format(time.RFC3339),
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 Bad Request for coords %v, got %d", coords, rec.Code)
			}
		}
	})

	t.Run("POST /v1/reports returns 503 when store fails", func(t *testing.T) {
		failingStore := newMockReportStore()
		failingStore.err = errors.New("db connection lost")
		hFail := NewRouter(Dependencies{
			ReportStore: failingStore,
			Ready:       func(context.Context) error { return nil },
		})
		payload := map[string]any{
			"id":                  uuid.New().String(),
			"species_common_name": "mule_deer",
			"latitude":            43.8,
			"longitude":           -110.2,
			"observed_at":         obsTime.Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/reports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		hFail.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 Service Unavailable, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /v1/reports/{id}/photo uploads photo, strips EXIF GPS, and returns suggestions", func(t *testing.T) {
		testImgWithEXIF := createTestJPEGWithEXIF(t)
		if !bytes.Contains(testImgWithEXIF, []byte("GPS:")) {
			t.Fatal("pre-check failed: test image does not contain GPS tag")
		}

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/reports/%s/photo", reportID), bytes.NewReader(testImgWithEXIF))
		req.Header.Set("Content-Type", "image/jpeg")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var res struct {
			Status string          `json:"status"`
			Data   PhotoUploadData `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.Status != "success" {
			t.Fatalf("expected status success, got %q", res.Status)
		}
		expectedKey := fmt.Sprintf("corridor-raw/photos/%s.jpg", reportID)
		if res.Data.PhotoKey != expectedKey {
			t.Fatalf("expected photo key %q, got %q", expectedKey, res.Data.PhotoKey)
		}
		if len(res.Data.SpeciesSuggestions) == 0 {
			t.Fatal("expected species suggestions chips, got empty")
		}
		topSuggestion := res.Data.SpeciesSuggestions[0]
		if topSuggestion.Species != "mule_deer" || topSuggestion.Confidence <= 0 {
			t.Fatalf("unexpected top suggestion: %+v", topSuggestion)
		}

		// Verify stored bytes in private storage had EXIF GPS stripped
		storedBytes, err := photoStore.GetPhoto(context.Background(), expectedKey)
		if err != nil {
			t.Fatalf("failed to retrieve stored photo: %v", err)
		}
		if bytes.Contains(storedBytes, []byte("GPS:")) {
			t.Fatal("privacy violation: stored photo contains GPS EXIF tags!")
		}
		if bytes.Contains(storedBytes, []byte("Exif")) {
			t.Fatal("privacy violation: stored photo contains Exif marker!")
		}

		// Verify the stored bytes are still a valid decodeable image
		_, _, err = image.Decode(bytes.NewReader(storedBytes))
		if err != nil {
			t.Fatalf("stripped image failed to decode: %v", err)
		}
	})

	t.Run("POST /v1/reports/{id}/photo supports multipart form upload", func(t *testing.T) {
		testImgWithEXIF := createTestJPEGWithEXIF(t)
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		part, err := mw.CreateFormFile("photo", "carcass.jpg")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write(testImgWithEXIF); err != nil {
			t.Fatalf("failed to write photo bytes: %v", err)
		}
		mw.Close()

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/reports/%s/photo", reportID), &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for multipart upload, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /v1/reports/{id}/photo returns 404 for unknown report ID", func(t *testing.T) {
		unknownID := uuid.New()
		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/reports/%s/photo", unknownID), bytes.NewReader(createTestJPEGWithEXIF(t)))
		req.Header.Set("Content-Type", "image/jpeg")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /v1/reports/{id}/photo returns 400 for corrupt or empty image", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/reports/%s/photo", reportID), bytes.NewReader([]byte("not an image")))
		req.Header.Set("Content-Type", "image/jpeg")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request for corrupt image, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /v1/reports/{id}/photo returns 503 when photo storage fails", func(t *testing.T) {
		failingPhotoStore := newMockPhotoStore()
		failingPhotoStore.err = errors.New("s3 bucket offline")
		hFail := NewRouter(Dependencies{
			ReportStore: reportStore,
			PhotoStore:  failingPhotoStore,
			Ready:       func(context.Context) error { return nil },
		})

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/reports/%s/photo", reportID), bytes.NewReader(createTestJPEGWithEXIF(t)))
		req.Header.Set("Content-Type", "image/jpeg")
		rec := httptest.NewRecorder()
		hFail.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 Service Unavailable, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
