package ingest

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// UpdateInterval represents the expected cadence of data updates for a feed.
type UpdateInterval string

const (
	// IntervalDaily indicates feeds expected to update once every 24 hours.
	IntervalDaily UpdateInterval = "daily"
	// IntervalWeekly indicates feeds expected to update once per week.
	IntervalWeekly UpdateInterval = "weekly"
	// IntervalMonthly indicates feeds expected to update once per month.
	IntervalMonthly UpdateInterval = "monthly"
	// IntervalAnnual indicates feeds expected to update once per year.
	IntervalAnnual UpdateInterval = "annual"
)

// MaxAcceptableAge returns the allowable staleness threshold including grace period.
func (u UpdateInterval) MaxAcceptableAge() time.Duration {
	switch u {
	case IntervalDaily:
		return 48 * time.Hour
	case IntervalWeekly:
		return 14 * 24 * time.Hour
	case IntervalMonthly:
		return 45 * 24 * time.Hour
	case IntervalAnnual:
		return 400 * 24 * time.Hour
	default:
		return 45 * 24 * time.Hour
	}
}

// HealthStatus classifies the operational state of a public data feed.
type HealthStatus string

const (
	// StatusHealthy indicates the feed is reachable, fresh, and conforms to expected schema.
	StatusHealthy HealthStatus = "HEALTHY"
	// StatusDegraded indicates partial availability or warning conditions.
	StatusDegraded HealthStatus = "DEGRADED"
	// StatusStale indicates the feed is accessible but has not been updated within its expected window.
	StatusStale HealthStatus = "STALE"
	// StatusSchemaDrift indicates the feed structure, columns, or content-type have changed unexpectedly.
	StatusSchemaDrift HealthStatus = "SCHEMA_DRIFT"
	// StatusOutage indicates the endpoint is unreachable, timed out, or returned an HTTP 4xx/5xx error.
	StatusOutage HealthStatus = "OUTAGE"
)

// FeedFormat classifies the transport encoding of a public feed.
type FeedFormat string

const (
	// FormatCSV represents tabular comma-separated values.
	FormatCSV FeedFormat = "csv"
	// FormatJSON represents standard JSON payload.
	FormatJSON FeedFormat = "json"
	// FormatGeoJSON represents RFC 7946 GeoJSON.
	FormatGeoJSON FeedFormat = "geojson"
	// FormatArcGIS represents ESRI ArcGIS REST FeatureServer/MapServer queries.
	FormatArcGIS FeedFormat = "arcgis_rest"
	// FormatBinary represents compressed archives or binary artifacts (e.g., zip, shapefile).
	FormatBinary FeedFormat = "binary"
	// FormatHTML represents web portals or scraping targets.
	FormatHTML FeedFormat = "html"
)

// FeedSource defines a registered public data source monitored for freshness and schema drift.
type FeedSource struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	URL                 string         `json:"url"`
	Format              FeedFormat     `json:"format"`
	UpdateInterval      UpdateInterval `json:"update_interval"`
	MaxAcceptableAge    time.Duration  `json:"max_acceptable_age,omitempty"`
	ExpectedContentType string         `json:"expected_content_type,omitempty"`
	ExpectedColumns     []string       `json:"expected_columns,omitempty"`
	ExpectedJSONFields  []string       `json:"expected_json_fields,omitempty"`
	License             string         `json:"license,omitempty"`
	SourceRef           string         `json:"source_ref,omitempty"`
	Timeout             time.Duration  `json:"timeout,omitempty"`
	RequireBodyCheck    bool           `json:"require_body_check,omitempty"`
}

// FeedHealthReport captures the health, freshness, and structural integrity of a single feed source.
type FeedHealthReport struct {
	SourceID           string       `json:"source_id"`
	SourceName         string       `json:"source_name"`
	URL                string       `json:"url"`
	CheckedAt          time.Time    `json:"checked_at"`
	Status             HealthStatus `json:"status"`
	StatusCode         int          `json:"status_code"`
	ResponseTimeMs     int64        `json:"response_time_ms"`
	ContentLength      int64        `json:"content_length"`
	ContentType        string       `json:"content_type"`
	ETag               string       `json:"etag,omitempty"`
	LastModified       *time.Time   `json:"last_modified,omitempty"`
	IsAvailable        bool         `json:"is_available"`
	IsFresh            bool         `json:"is_fresh"`
	FreshnessAgeSec    *int64       `json:"freshness_age_sec,omitempty"`
	SchemaDrift        bool         `json:"schema_drift"`
	SchemaDriftDetails []string     `json:"schema_drift_details,omitempty"`
	ErrorMessage       string       `json:"error_message,omitempty"`
	MethodUsed         string       `json:"method_used"`
}

// FeedHealthSummary aggregates health reports across multiple feed sources.
type FeedHealthSummary struct {
	CheckedAt        time.Time          `json:"checked_at"`
	TotalFeeds       int                `json:"total_feeds"`
	HealthyCount     int                `json:"healthy_count"`
	StaleCount       int                `json:"stale_count"`
	SchemaDriftCount int                `json:"schema_drift_count"`
	OutageCount      int                `json:"outage_count"`
	DegradedCount    int                `json:"degraded_count"`
	Reports          []FeedHealthReport `json:"reports"`
}

var (
	feedCatalogMu sync.RWMutex
	feedCatalog   = DefaultFeedCatalog()

	defaultHTTPClient = &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
)

// DefaultFeedCatalog returns the authoritative list of public feeds documented in data/SOURCES.md.
func DefaultFeedCatalog() []FeedSource {
	return []FeedSource{
		{
			ID:                  "cros-figshare",
			Name:                "California Roadkill Observation System (CROS) Figshare",
			URL:                 "https://api.figshare.com/v2/articles/29123012",
			Format:              FormatJSON,
			UpdateInterval:      IntervalMonthly,
			ExpectedContentType: "application/json",
			ExpectedJSONFields:  []string{"id", "title"},
			License:             "CC BY 4.0 / CC BY-SA 4.0 (Quarantined)",
			SourceRef:           "S23",
		},
		{
			ID:             "utah-open-data",
			Name:           "Utah Wildlife Collisions / Roadkill Reporter",
			URL:            "https://wildlifecollisions.utah.gov/",
			Format:         FormatHTML,
			UpdateInterval: IntervalDaily,
			License:        "Requires Request / Unconfirmed",
			SourceRef:      "S22",
		},
		{
			ID:                 "usgs-ungulate-migration",
			Name:               "USGS Ungulate Migration Corridor Portal - Volume 1",
			URL:                "https://data.usgs.gov/datacatalog/data/USGS:5f80c88d82cebef40f0fefc5",
			Format:             FormatJSON,
			UpdateInterval:     IntervalAnnual,
			ExpectedJSONFields: []string{"id"},
			License:            "USGS Public Domain",
			SourceRef:          "S11",
		},
		{
			ID:                  "nps-boundary-gis",
			Name:                "National Park Service Administrative Boundaries (NGDA)",
			URL:                 "https://services1.arcgis.com/fBc8EJBx5MT03Fd8/arcgis/rest/services/National_Park_Service_Administrative_Boundaries/FeatureServer/0?f=pjson",
			Format:              FormatArcGIS,
			UpdateInterval:      IntervalMonthly,
			ExpectedContentType: "application/json",
			ExpectedJSONFields:  []string{"name", "fields"},
			License:             "USGS / NPS Public Domain",
			SourceRef:           "S12",
		},
	}
}

// RegisterFeedSource adds or updates a feed source in the registered catalog.
func RegisterFeedSource(source FeedSource) {
	feedCatalogMu.Lock()
	defer feedCatalogMu.Unlock()
	for i, s := range feedCatalog {
		if s.ID == source.ID {
			feedCatalog[i] = source
			return
		}
	}
	feedCatalog = append(feedCatalog, source)
}

// SetFeedCatalog sets the catalog to a specified slice of sources.
func SetFeedCatalog(sources []FeedSource) {
	feedCatalogMu.Lock()
	defer feedCatalogMu.Unlock()
	feedCatalog = append([]FeedSource(nil), sources...)
}

// ResetFeedCatalog restores the catalog to default sources.
func ResetFeedCatalog() {
	feedCatalogMu.Lock()
	defer feedCatalogMu.Unlock()
	feedCatalog = DefaultFeedCatalog()
}

// RegisteredFeedSources returns a shallow copy of the registered feed sources.
func RegisteredFeedSources() []FeedSource {
	feedCatalogMu.RLock()
	defer feedCatalogMu.RUnlock()
	copied := make([]FeedSource, len(feedCatalog))
	copy(copied, feedCatalog)
	return copied
}

// CheckFeedHealth inspects a single public data feed source using the default HTTP client.
func CheckFeedHealth(ctx context.Context, source FeedSource) (*FeedHealthReport, error) {
	return CheckFeedHealthWithClient(ctx, defaultHTTPClient, source)
}

// CheckFeedHealthWithClient inspects a single public data feed using the provided HTTP client.
func CheckFeedHealthWithClient(ctx context.Context, client *http.Client, source FeedSource) (*FeedHealthReport, error) {
	if source.ID == "" || source.URL == "" {
		return nil, errors.New("feed source requires valid ID and URL")
	}

	timeout := source.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	needsBody := source.RequireBodyCheck || len(source.ExpectedColumns) > 0 || len(source.ExpectedJSONFields) > 0
	method := http.MethodHead
	if needsBody {
		method = http.MethodGet
	}

	start := time.Now()
	report := &FeedHealthReport{
		SourceID:   source.ID,
		SourceName: source.Name,
		URL:        source.URL,
		CheckedAt:  start.UTC(),
		MethodUsed: method,
		IsFresh:    true,
	}

	resp, err := executeHTTPRequest(reqCtx, client, method, source.URL)
	if err != nil {
		report.ResponseTimeMs = time.Since(start).Milliseconds()
		report.Status = StatusOutage
		report.IsAvailable = false
		report.ErrorMessage = err.Error()
		return report, nil
	}
	defer resp.Body.Close()

	// If HEAD received 405 Method Not Allowed or 501 Not Implemented, retry with GET
	if method == http.MethodHead && (resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented) {
		method = http.MethodGet
		report.MethodUsed = method

		resp2, err2 := executeHTTPRequest(reqCtx, client, method, source.URL)
		if err2 != nil {
			report.ResponseTimeMs = time.Since(start).Milliseconds()
			report.Status = StatusOutage
			report.IsAvailable = false
			report.ErrorMessage = err2.Error()
			return report, nil
		}
		defer resp2.Body.Close()
		resp = resp2
	}

	report.ResponseTimeMs = time.Since(start).Milliseconds()
	report.StatusCode = resp.StatusCode
	report.ContentType = resp.Header.Get("Content-Type")
	report.ETag = resp.Header.Get("ETag")
	report.ContentLength = resp.ContentLength

	// Parse Last-Modified header
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if parsed, parseErr := http.ParseTime(lm); parseErr == nil {
			utc := parsed.UTC()
			report.LastModified = &utc
		}
	}

	// Endpoint outage check
	if resp.StatusCode >= 400 {
		report.Status = StatusOutage
		report.IsAvailable = false
		report.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return report, nil
	}

	report.IsAvailable = true
	report.Status = StatusHealthy

	// Evaluate Freshness
	if report.LastModified != nil {
		age := time.Since(*report.LastModified)
		if age < 0 {
			age = 0
		}
		ageSec := int64(age.Seconds())
		report.FreshnessAgeSec = &ageSec

		maxAge := source.MaxAcceptableAge
		if maxAge <= 0 {
			maxAge = source.UpdateInterval.MaxAcceptableAge()
		}

		if age > maxAge {
			report.IsFresh = false
			report.Status = StatusStale
			report.ErrorMessage = fmt.Sprintf("feed data is stale: last modified %s ago (max allowed: %s)", age.Round(time.Hour), maxAge)
		}
	}

	// Evaluate Content-Type Drift
	if source.ExpectedContentType != "" {
		actualCT := strings.ToLower(resp.Header.Get("Content-Type"))
		expectedCT := strings.ToLower(source.ExpectedContentType)
		if !strings.Contains(actualCT, expectedCT) {
			report.SchemaDrift = true
			report.SchemaDriftDetails = append(report.SchemaDriftDetails,
				fmt.Sprintf("content-type mismatch: expected %q, got %q", source.ExpectedContentType, resp.Header.Get("Content-Type")))
			report.Status = StatusSchemaDrift
		}
	}

	// Evaluate Body Schema Drift (CSV headers or JSON keys)
	if needsBody {
		// Limit reading to 1MB to prevent OOM on large archives
		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			report.Status = StatusDegraded
			report.ErrorMessage = fmt.Sprintf("read response body failed: %v", readErr)
			return report, nil
		}

		// CSV Schema Validation
		if len(source.ExpectedColumns) > 0 {
			csvReader := csv.NewReader(bytes.NewReader(bodyBytes))
			headers, csvErr := csvReader.Read()
			if csvErr != nil {
				report.SchemaDrift = true
				report.SchemaDriftDetails = append(report.SchemaDriftDetails, fmt.Sprintf("malformed CSV header: %v", csvErr))
				report.Status = StatusSchemaDrift
			} else {
				headerSet := make(map[string]bool, len(headers))
				for _, h := range headers {
					headerSet[strings.TrimSpace(h)] = true
				}
				var missing []string
				for _, expected := range source.ExpectedColumns {
					if !headerSet[expected] {
						missing = append(missing, expected)
					}
				}
				if len(missing) > 0 {
					report.SchemaDrift = true
					report.SchemaDriftDetails = append(report.SchemaDriftDetails,
						fmt.Sprintf("missing expected CSV columns: %s", strings.Join(missing, ", ")))
					report.Status = StatusSchemaDrift
				}
			}
		}

		// JSON / GeoJSON / ArcGIS Schema Validation
		if len(source.ExpectedJSONFields) > 0 {
			var jsonMap map[string]any
			if jsonErr := json.Unmarshal(bodyBytes, &jsonMap); jsonErr != nil {
				report.SchemaDrift = true
				report.SchemaDriftDetails = append(report.SchemaDriftDetails, fmt.Sprintf("invalid JSON payload: %v", jsonErr))
				report.Status = StatusSchemaDrift
			} else {
				var missing []string
				for _, field := range source.ExpectedJSONFields {
					if _, exists := jsonMap[field]; !exists {
						missing = append(missing, field)
					}
				}
				if len(missing) > 0 {
					report.SchemaDrift = true
					report.SchemaDriftDetails = append(report.SchemaDriftDetails,
						fmt.Sprintf("missing expected JSON fields: %s", strings.Join(missing, ", ")))
					report.Status = StatusSchemaDrift
				}
			}
		}
	}

	return report, nil
}

func executeHTTPRequest(ctx context.Context, client *http.Client, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Corridor-FeedMonitor/1.0 (+https://github.com/corridor)")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		return nil, fmt.Errorf("network error: %w", err)
	}
	return resp, nil
}

// CheckAllFeeds queries the registered source catalog and returns health reports for all feeds.
func CheckAllFeeds(ctx context.Context) ([]FeedHealthReport, error) {
	return CheckAllFeedsWithClient(ctx, defaultHTTPClient)
}

// CheckAllFeedsWithClient queries the registered source catalog using the specified HTTP client.
func CheckAllFeedsWithClient(ctx context.Context, client *http.Client) ([]FeedHealthReport, error) {
	sources := RegisteredFeedSources()
	return CheckSourcesWithClient(ctx, client, sources)
}

// CheckSourcesWithClient performs health checks across a given slice of feed sources concurrently.
func CheckSourcesWithClient(ctx context.Context, client *http.Client, sources []FeedSource) ([]FeedHealthReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	reports := make([]FeedHealthReport, len(sources))
	var wg sync.WaitGroup

	// Concurrency limiter to prevent connection spikes
	concurrency := 5
	if len(sources) < concurrency {
		concurrency = len(sources)
	}
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)

	for i, src := range sources {
		wg.Add(1)
		go func(idx int, s FeedSource) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				reports[idx] = FeedHealthReport{
					SourceID:     s.ID,
					SourceName:   s.Name,
					URL:          s.URL,
					CheckedAt:    time.Now().UTC(),
					Status:       StatusOutage,
					IsAvailable:  false,
					ErrorMessage: ctx.Err().Error(),
				}
				return
			}

			rep, _ := CheckFeedHealthWithClient(ctx, client, s)
			if rep != nil {
				reports[idx] = *rep
			}
		}(i, src)
	}

	wg.Wait()
	return reports, nil
}

// SummarizeReports compiles an aggregated summary from a list of feed health reports.
func SummarizeReports(reports []FeedHealthReport) FeedHealthSummary {
	summary := FeedHealthSummary{
		CheckedAt:  time.Now().UTC(),
		TotalFeeds: len(reports),
		Reports:    reports,
	}

	for _, r := range reports {
		switch r.Status {
		case StatusHealthy:
			summary.HealthyCount++
		case StatusStale:
			summary.StaleCount++
		case StatusSchemaDrift:
			summary.SchemaDriftCount++
		case StatusOutage:
			summary.OutageCount++
		case StatusDegraded:
			summary.DegradedCount++
		}
	}

	return summary
}

// ToJSON formats a FeedHealthReport as an indented JSON string.
func (r *FeedHealthReport) ToJSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report json: %w", err)
	}
	return string(b), nil
}

// ToJSON formats a FeedHealthSummary as an indented JSON string.
func (s *FeedHealthSummary) ToJSON() (string, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal summary json: %w", err)
	}
	return string(b), nil
}

// FormatReportsJSON converts an array of reports into an indented JSON summary document.
func FormatReportsJSON(reports []FeedHealthReport) (string, error) {
	summary := SummarizeReports(reports)
	return summary.ToJSON()
}

// FormatSummaryText produces an operational dashboard summary in plain text.
func FormatSummaryText(summary FeedHealthSummary) string {
	var sb strings.Builder

	sb.WriteString("================================================================================\n")
	sb.WriteString("CORRIDOR DATA INGESTION FEED HEALTH REPORT\n")
	sb.WriteString(fmt.Sprintf("Checked At: %s\n", summary.CheckedAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("Total Feeds: %d | Healthy: %d | Stale: %d | Schema Drift: %d | Outages: %d\n",
		summary.TotalFeeds, summary.HealthyCount, summary.StaleCount, summary.SchemaDriftCount, summary.OutageCount))
	sb.WriteString("================================================================================\n\n")

	for _, r := range summary.Reports {
		sb.WriteString(r.FormatText())
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatReportsText formats a slice of feed reports as a human-readable operational summary.
func FormatReportsText(reports []FeedHealthReport) string {
	summary := SummarizeReports(reports)
	return FormatSummaryText(summary)
}

// FormatText formats a single report into human-readable multi-line plain text.
func (r *FeedHealthReport) FormatText() string {
	var sb strings.Builder

	statusBadge := fmt.Sprintf("[%s]", r.Status)
	sb.WriteString(fmt.Sprintf("%-14s %s (%s)\n", statusBadge, r.SourceName, r.SourceID))
	sb.WriteString(fmt.Sprintf("  URL:            %s\n", r.URL))

	if r.StatusCode > 0 {
		sb.WriteString(fmt.Sprintf("  Status Code:    %d %s (%s %dms)\n",
			r.StatusCode, http.StatusText(r.StatusCode), r.MethodUsed, r.ResponseTimeMs))
	} else {
		sb.WriteString(fmt.Sprintf("  Response Time:  %dms (%s)\n", r.ResponseTimeMs, r.MethodUsed))
	}

	if r.ContentType != "" {
		sb.WriteString(fmt.Sprintf("  Content-Type:   %s", r.ContentType))
		if r.ContentLength >= 0 {
			sb.WriteString(fmt.Sprintf(" (Size: %d bytes)\n", r.ContentLength))
		} else {
			sb.WriteString("\n")
		}
	}

	if r.ETag != "" {
		sb.WriteString(fmt.Sprintf("  ETag:           %s\n", r.ETag))
	}

	if r.LastModified != nil {
		age := time.Since(*r.LastModified)
		sb.WriteString(fmt.Sprintf("  Last-Modified:  %s (%s ago)\n",
			r.LastModified.Format("2006-01-02 15:04:05 UTC"), age.Round(time.Minute)))
	}

	if len(r.SchemaDriftDetails) > 0 {
		sb.WriteString("  Schema Drift:\n")
		for _, detail := range r.SchemaDriftDetails {
			sb.WriteString(fmt.Sprintf("    - %s\n", detail))
		}
	}

	if r.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("  Issue:          %s\n", r.ErrorMessage))
	}

	return sb.String()
}
