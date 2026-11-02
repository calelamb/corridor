package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFeedHealthCheck_SuccessHEAD(t *testing.T) {
	lastModified := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("expected HEAD request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"feed-hash-987"`)
		w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
		w.Header().Set("Content-Length", "4096")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	source := FeedSource{
		ID:             "test-feed-head",
		Name:           "Test Head Feed",
		URL:            server.URL,
		Format:         FormatJSON,
		UpdateInterval: IntervalDaily,
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %s", report.Status)
	}
	if !report.IsAvailable {
		t.Error("expected IsAvailable to be true")
	}
	if !report.IsFresh {
		t.Error("expected IsFresh to be true")
	}
	if report.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", report.StatusCode)
	}
	if report.ETag != `"feed-hash-987"` {
		t.Errorf("expected ETag \"feed-hash-987\", got %s", report.ETag)
	}
	if report.MethodUsed != http.MethodHead {
		t.Errorf("expected MethodUsed HEAD, got %s", report.MethodUsed)
	}
	if report.LastModified == nil || !report.LastModified.Equal(lastModified) {
		t.Errorf("expected LastModified %v, got %v", lastModified, report.LastModified)
	}
	if report.FreshnessAgeSec == nil || *report.FreshnessAgeSec < 3500 || *report.FreshnessAgeSec > 3700 {
		t.Errorf("unexpected FreshnessAgeSec: %v", report.FreshnessAgeSec)
	}
}

func TestFeedHealthCheck_SuccessGET_WithCSVValidation(t *testing.T) {
	csvBody := "occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude\n1,US,Alces alces,-110.5,44.5\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvBody))
	}))
	defer server.Close()

	source := FeedSource{
		ID:                  "test-csv-feed",
		Name:                "Test CSV Feed",
		URL:                 server.URL,
		Format:              FormatCSV,
		UpdateInterval:      IntervalMonthly,
		ExpectedContentType: "text/csv",
		ExpectedColumns:     []string{"occurrenceID", "scientificName"},
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %s", report.Status)
	}
	if report.SchemaDrift {
		t.Errorf("unexpected schema drift: %v", report.SchemaDriftDetails)
	}
	if report.MethodUsed != http.MethodGet {
		t.Errorf("expected GET method used, got %s", report.MethodUsed)
	}
}

func TestFeedHealthCheck_SuccessGET_WithJSONValidation(t *testing.T) {
	jsonData := `{"id": 29123012, "title": "CA State Roadkill Observation Data", "files": [{"name": "records.csv"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonData))
	}))
	defer server.Close()

	source := FeedSource{
		ID:                  "cros-figshare-test",
		Name:                "CROS Figshare Test",
		URL:                 server.URL,
		Format:              FormatJSON,
		UpdateInterval:      IntervalMonthly,
		ExpectedContentType: "application/json",
		ExpectedJSONFields:  []string{"id", "title"},
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %s", report.Status)
	}
	if report.SchemaDrift {
		t.Errorf("unexpected schema drift: %v", report.SchemaDriftDetails)
	}
}

func TestFeedHealthCheck_StaleFeed(t *testing.T) {
	// Daily feed modified 5 days ago (max acceptable: 48h)
	staleDate := time.Now().Add(-5 * 24 * time.Hour).UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Last-Modified", staleDate.Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	source := FeedSource{
		ID:             "stale-feed",
		Name:           "Stale Daily Feed",
		URL:            server.URL,
		Format:         FormatJSON,
		UpdateInterval: IntervalDaily,
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusStale {
		t.Errorf("expected StatusStale, got %s", report.Status)
	}
	if report.IsFresh {
		t.Error("expected IsFresh to be false")
	}
	if report.FreshnessAgeSec == nil || *report.FreshnessAgeSec < 4*24*3600 {
		t.Errorf("unexpected FreshnessAgeSec: %v", report.FreshnessAgeSec)
	}
	if !strings.Contains(report.ErrorMessage, "stale") {
		t.Errorf("expected stale error message, got: %s", report.ErrorMessage)
	}
}

func TestFeedHealthCheck_Errors_404_500(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Error", http.StatusInternalServerError},
		{"502 Bad Gateway", http.StatusBadGateway},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			source := FeedSource{
				ID:             "error-feed",
				Name:           "Error Feed",
				URL:            server.URL,
				Format:         FormatJSON,
				UpdateInterval: IntervalDaily,
			}

			report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if report.Status != StatusOutage {
				t.Errorf("expected StatusOutage, got %s", report.Status)
			}
			if report.IsAvailable {
				t.Error("expected IsAvailable to be false")
			}
			if report.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, report.StatusCode)
			}
			if !strings.Contains(report.ErrorMessage, fmt.Sprintf("%d", tt.statusCode)) {
				t.Errorf("expected error message to contain status %d, got: %s", tt.statusCode, report.ErrorMessage)
			}
		})
	}
}

func TestFeedHealthCheck_TimeoutHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	source := FeedSource{
		ID:             "timeout-feed",
		Name:           "Timeout Feed",
		URL:            server.URL,
		Format:         FormatJSON,
		UpdateInterval: IntervalDaily,
		Timeout:        40 * time.Millisecond,
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusOutage {
		t.Errorf("expected StatusOutage, got %s", report.Status)
	}
	if report.IsAvailable {
		t.Error("expected IsAvailable to be false")
	}
	if !strings.Contains(strings.ToLower(report.ErrorMessage), "time") &&
		!strings.Contains(strings.ToLower(report.ErrorMessage), "context deadline exceeded") {
		t.Errorf("expected timeout message, got: %s", report.ErrorMessage)
	}
}

func TestFeedHealthCheck_SchemaDrift_CSV_MissingColumn(t *testing.T) {
	// Header missing "scientificName"
	csvHeaderOnly := "occurrenceID,countryCode,latitude,longitude\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvHeaderOnly))
	}))
	defer server.Close()

	source := FeedSource{
		ID:                  "drift-csv",
		Name:                "Drift CSV",
		URL:                 server.URL,
		Format:              FormatCSV,
		UpdateInterval:      IntervalMonthly,
		ExpectedContentType: "text/csv",
		ExpectedColumns:     []string{"occurrenceID", "scientificName"},
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusSchemaDrift {
		t.Errorf("expected StatusSchemaDrift, got %s", report.Status)
	}
	if !report.SchemaDrift {
		t.Error("expected SchemaDrift to be true")
	}
	if len(report.SchemaDriftDetails) == 0 || !strings.Contains(report.SchemaDriftDetails[0], "scientificName") {
		t.Errorf("expected drift details to mention scientificName, got: %v", report.SchemaDriftDetails)
	}
}

func TestFeedHealthCheck_SchemaDrift_JSON_MissingField(t *testing.T) {
	invalidJSON := `{"data": 123, "count": 456}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(invalidJSON))
	}))
	defer server.Close()

	source := FeedSource{
		ID:                  "drift-json",
		Name:                "Drift JSON",
		URL:                 server.URL,
		Format:              FormatJSON,
		UpdateInterval:      IntervalAnnual,
		ExpectedContentType: "application/json",
		ExpectedJSONFields:  []string{"id", "title"},
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusSchemaDrift {
		t.Errorf("expected StatusSchemaDrift, got %s", report.Status)
	}
	if !report.SchemaDrift {
		t.Error("expected SchemaDrift to be true")
	}
	if len(report.SchemaDriftDetails) == 0 {
		t.Errorf("expected drift details, got empty: %v", report.SchemaDriftDetails)
	}
}

func TestFeedHealthCheck_SchemaDrift_ContentTypeMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>Cloudflare Portal Error</html>"))
	}))
	defer server.Close()

	source := FeedSource{
		ID:                  "ct-mismatch",
		Name:                "Content Type Mismatch",
		URL:                 server.URL,
		Format:              FormatJSON,
		UpdateInterval:      IntervalDaily,
		ExpectedContentType: "application/json",
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusSchemaDrift {
		t.Errorf("expected StatusSchemaDrift, got %s", report.Status)
	}
	if !report.SchemaDrift {
		t.Error("expected SchemaDrift to be true")
	}
	if len(report.SchemaDriftDetails) == 0 || !strings.Contains(report.SchemaDriftDetails[0], "content-type mismatch") {
		t.Errorf("expected content-type mismatch detail, got: %v", report.SchemaDriftDetails)
	}
}

func TestFeedHealthCheck_MethodFallback_405(t *testing.T) {
	var headCount, getCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			atomic.AddInt32(&headCount, 1)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method == http.MethodGet {
			atomic.AddInt32(&getCount, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	source := FeedSource{
		ID:             "fallback-405",
		Name:           "Fallback 405 Feed",
		URL:            server.URL,
		Format:         FormatJSON,
		UpdateInterval: IntervalDaily,
	}

	report, err := CheckFeedHealthWithClient(context.Background(), server.Client(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy after fallback, got %s", report.Status)
	}
	if report.MethodUsed != http.MethodGet {
		t.Errorf("expected MethodUsed GET after fallback, got %s", report.MethodUsed)
	}
	if atomic.LoadInt32(&headCount) != 1 || atomic.LoadInt32(&getCount) != 1 {
		t.Errorf("expected 1 HEAD and 1 GET, got %d HEAD and %d GET", headCount, getCount)
	}
}

func TestFeedCheckAllFeeds_Concurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cros":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "title": "CROS"}`))
		case "/utah":
			w.WriteHeader(http.StatusOK)
		case "/usgs":
			w.WriteHeader(http.StatusNotFound)
		case "/nps":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name": "NPS Boundaries", "fields": []}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	sources := []FeedSource{
		{
			ID:                  "cros",
			Name:                "CROS Figshare",
			URL:                 server.URL + "/cros",
			Format:              FormatJSON,
			UpdateInterval:      IntervalMonthly,
			ExpectedContentType: "application/json",
			ExpectedJSONFields:  []string{"id", "title"},
		},
		{
			ID:             "utah",
			Name:           "Utah Roadkill",
			URL:            server.URL + "/utah",
			Format:         FormatHTML,
			UpdateInterval: IntervalDaily,
		},
		{
			ID:             "usgs",
			Name:           "USGS Ungulate",
			URL:            server.URL + "/usgs",
			Format:         FormatJSON,
			UpdateInterval: IntervalAnnual,
		},
		{
			ID:                  "nps",
			Name:                "NPS Boundaries",
			URL:                 server.URL + "/nps",
			Format:              FormatArcGIS,
			UpdateInterval:      IntervalMonthly,
			ExpectedContentType: "application/json",
			ExpectedJSONFields:  []string{"name", "fields"},
		},
	}

	reports, err := CheckSourcesWithClient(context.Background(), server.Client(), sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 4 {
		t.Fatalf("expected 4 reports, got %d", len(reports))
	}

	if reports[0].Status != StatusHealthy {
		t.Errorf("expected CROS healthy, got %s", reports[0].Status)
	}
	if reports[1].Status != StatusHealthy {
		t.Errorf("expected Utah healthy, got %s", reports[1].Status)
	}
	if reports[2].Status != StatusOutage {
		t.Errorf("expected USGS outage, got %s", reports[2].Status)
	}
	if reports[3].Status != StatusHealthy {
		t.Errorf("expected NPS healthy, got %s", reports[3].Status)
	}

	summary := SummarizeReports(reports)
	if summary.TotalFeeds != 4 || summary.HealthyCount != 3 || summary.OutageCount != 1 {
		t.Errorf("unexpected summary metrics: %+v", summary)
	}
}

func TestFeedSummaryFormatting_PlainAndJSON(t *testing.T) {
	now := time.Now().UTC()
	reports := []FeedHealthReport{
		{
			SourceID:       "cros-figshare",
			SourceName:     "CROS Figshare",
			URL:            "https://example.com/cros",
			CheckedAt:      now,
			Status:         StatusHealthy,
			StatusCode:     http.StatusOK,
			ResponseTimeMs: 42,
			ContentLength:  2048,
			ContentType:    "application/json",
			ETag:           `"feed-123"`,
			IsAvailable:    true,
			IsFresh:        true,
			MethodUsed:     "GET",
		},
		{
			SourceID:       "stale-feed",
			SourceName:     "Utah Stale",
			URL:            "https://example.com/utah",
			CheckedAt:      now,
			Status:         StatusStale,
			StatusCode:     http.StatusOK,
			ResponseTimeMs: 120,
			IsAvailable:    true,
			IsFresh:        false,
			ErrorMessage:   "feed data is stale: last modified 96h ago (max allowed: 48h)",
			MethodUsed:     "HEAD",
		},
		{
			SourceID:           "drift-feed",
			SourceName:         "USGS Drift",
			URL:                "https://example.com/usgs",
			CheckedAt:          now,
			Status:             StatusSchemaDrift,
			StatusCode:         http.StatusOK,
			ResponseTimeMs:     65,
			IsAvailable:        true,
			IsFresh:            true,
			SchemaDrift:        true,
			SchemaDriftDetails: []string{"missing expected CSV columns: scientificName"},
			MethodUsed:         "GET",
		},
		{
			SourceID:       "down-feed",
			SourceName:     "NPS Down",
			URL:            "https://example.com/nps",
			CheckedAt:      now,
			Status:         StatusOutage,
			StatusCode:     http.StatusServiceUnavailable,
			ResponseTimeMs: 250,
			IsAvailable:    false,
			ErrorMessage:   "HTTP 503: Service Unavailable",
			MethodUsed:     "GET",
		},
	}

	// Test Plain Text Formatting
	textOutput := FormatReportsText(reports)
	if !strings.Contains(textOutput, "CORRIDOR DATA INGESTION FEED HEALTH REPORT") {
		t.Error("text output missing header")
	}
	if !strings.Contains(textOutput, "[HEALTHY]") || !strings.Contains(textOutput, "[STALE]") ||
		!strings.Contains(textOutput, "[SCHEMA_DRIFT]") || !strings.Contains(textOutput, "[OUTAGE]") {
		t.Error("text output missing status tags")
	}
	if !strings.Contains(textOutput, "scientificName") {
		t.Error("text output missing schema drift detail")
	}
	if !strings.Contains(textOutput, "Total Feeds: 4 | Healthy: 1 | Stale: 1 | Schema Drift: 1 | Outages: 1") {
		t.Errorf("text output missing correct metrics summary: %s", textOutput)
	}

	// Test JSON Serialization Round-trip
	jsonOutput, err := FormatReportsJSON(reports)
	if err != nil {
		t.Fatalf("failed to format JSON: %v", err)
	}

	var parsedSummary FeedHealthSummary
	if err := json.Unmarshal([]byte(jsonOutput), &parsedSummary); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsedSummary.TotalFeeds != 4 {
		t.Errorf("expected 4 feeds in parsed JSON, got %d", parsedSummary.TotalFeeds)
	}
	if parsedSummary.HealthyCount != 1 || parsedSummary.StaleCount != 1 ||
		parsedSummary.SchemaDriftCount != 1 || parsedSummary.OutageCount != 1 {
		t.Errorf("unexpected counts in parsed JSON: %+v", parsedSummary)
	}
	if len(parsedSummary.Reports) != 4 {
		t.Errorf("expected 4 reports in parsed JSON, got %d", len(parsedSummary.Reports))
	}

	// Test single report ToJSON
	repJSON, err := reports[0].ToJSON()
	if err != nil {
		t.Fatalf("report ToJSON failed: %v", err)
	}
	if !strings.Contains(repJSON, `"source_id": "cros-figshare"`) {
		t.Errorf("report JSON missing source_id: %s", repJSON)
	}
}

func TestFeedCatalogRegistry(t *testing.T) {
	ResetFeedCatalog()
	initial := RegisteredFeedSources()
	if len(initial) != 4 {
		t.Fatalf("expected 4 default feeds, got %d", len(initial))
	}

	customSource := FeedSource{
		ID:             "custom-id",
		Name:           "Custom Source",
		URL:            "https://example.com/feed",
		Format:         FormatJSON,
		UpdateInterval: IntervalDaily,
	}

	RegisterFeedSource(customSource)
	updated := RegisteredFeedSources()
	if len(updated) != 5 {
		t.Fatalf("expected 5 feeds after registration, got %d", len(updated))
	}

	// Update existing
	customSource.Name = "Updated Custom Source"
	RegisterFeedSource(customSource)
	afterUpdate := RegisteredFeedSources()
	if len(afterUpdate) != 5 {
		t.Fatalf("expected 5 feeds after update, got %d", len(afterUpdate))
	}

	// Reset
	ResetFeedCatalog()
	afterReset := RegisteredFeedSources()
	if len(afterReset) != 4 {
		t.Fatalf("expected 4 feeds after reset, got %d", len(afterReset))
	}
}

func TestFeedValidation_InvalidSource(t *testing.T) {
	_, err := CheckFeedHealth(context.Background(), FeedSource{})
	if err == nil {
		t.Fatal("expected error on empty feed source")
	}
}
