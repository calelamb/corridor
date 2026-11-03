package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"corridor/internal/config"
	"corridor/internal/db"
	"corridor/internal/ingest"
	"corridor/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ingestCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: corridorctl ingest <all|irma|mdt|caltrans|wydot|udot|idfg> [flags]")
	}

	switch args[0] {
	case "all":
		if len(args) != 1 {
			return errors.New("usage: corridorctl ingest all")
		}
		return ingestAll(ctx, stdout)
	case "irma":
		return ingestIRMACmd(ctx, args[1:], stdout, stderr)
	case "mdt":
		return ingestMDTCmd(ctx, args[1:], stdout, stderr)
	case "caltrans":
		return ingestCaltransCmd(ctx, args[1:], stdout, stderr)
	case "wydot":
		return ingestWYDOTCmd(ctx, args[1:], stdout, stderr)
	case "udot":
		return ingestUDOTCmd(ctx, args[1:], stdout, stderr)
	case "idfg":
		return ingestIDFGCmd(ctx, args[1:], stdout, stderr)
	case "help", "--help", "-h":
		const ingestHelp = `corridorctl ingest - Run data ingestion pipelines and parsers

Usage:
  corridorctl ingest <subcommand> [flags]

Subcommands:
  all                             Execute full pilot ingestion pipeline
  irma --file <path> [--park <code>] [--json]
                                  Parse NPS IRMA wildlife observation and mortality records
  mdt --file <path> [--json]      Parse MDT Montana statewide carcass records
  caltrans (--file <path> | --url <url>) [--json]
                                  Parse Caltrans AADT and highway postmile records
  wydot --file <path> [--json]    Parse a WYDOT Roadkill Reports feature JSON page
  udot --file <path> [--json]     Parse a UDOT Crash Locations feature JSON page
  idfg --file <path> [--json]     Parse an IDFG Roadkill Observations feature JSON page
`
		_, err := fmt.Fprint(stdout, ingestHelp)
		return err
	default:
		return fmt.Errorf("unsupported ingest command %q; usage: corridorctl ingest <all|irma|mdt|caltrans|wydot|udot|idfg> [flags]", args[0])
	}
}

func ingestIRMACmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("ingest irma", flag.ContinueOnError)
	fs.SetOutput(stderr)

	filePath := fs.String("file", "", "Path to IRMA CSV or data file (required)")
	park := fs.String("park", "", "National Park filter code (e.g. YOSE, YELL, GRTE, GLAC, ROMO)")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *filePath == "" {
		return errors.New("usage: corridorctl ingest irma --file <path> [--park <code>]")
	}

	// #nosec G304 -- operator-specified file path for CLI ingestion
	file, err := os.Open(*filePath)
	if err != nil {
		return fmt.Errorf("open irma file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()

	var filter *ingest.IRMAFilter
	if *park != "" {
		filter = &ingest.IRMAFilter{
			ParkCodes: []string{strings.ToUpper(strings.TrimSpace(*park))},
		}
	}

	dataset, err := ingest.ParseIRMA(file, filter)
	if err != nil {
		return fmt.Errorf("parse irma dataset: %w", err)
	}

	speciesCounts := make(map[string]int64)
	for _, r := range dataset.Records {
		sp := r.NormalizedSpecies
		if sp == "" {
			sp = r.CommonName
		}
		if sp != "" {
			speciesCounts[sp]++
		}
	}

	out := struct {
		Source          string           `json:"source"`
		Park            string           `json:"park,omitempty"`
		RecordsRead     int64            `json:"records_read"`
		Accepted        int64            `json:"accepted"`
		Rejected        int64            `json:"rejected"`
		Duplicates      int64            `json:"duplicates"`
		SensitiveCount  int64            `json:"sensitive_count"`
		RoadkillCount   int64            `json:"roadkill_count"`
		SpeciesCounts   map[string]int64 `json:"species_counts,omitempty"`
		QuarantineDelay string           `json:"quarantine_delay"`
	}{
		Source:          "NPS IRMA Wildlife Mortality and Morbidity Observations",
		Park:            *park,
		RecordsRead:     dataset.QA.Read,
		Accepted:        dataset.QA.Accepted,
		Rejected:        dataset.QA.Rejected,
		Duplicates:      dataset.QA.Duplicates,
		SensitiveCount:  dataset.QA.SensitiveCount,
		RoadkillCount:   dataset.QA.RoadkillCount,
		SpeciesCounts:   speciesCounts,
		QuarantineDelay: "30-day anti-poaching quarantine applied to sensitive species",
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	parkStr := ""
	if *park != "" {
		parkStr = fmt.Sprintf(" for park %s", *park)
	}
	_, err = fmt.Fprintf(stdout, "Ingested %d IRMA records%s (accepted: %d, rejected: %d, sensitive: %d, roadkill: %d).\n",
		dataset.QA.Read, parkStr, dataset.QA.Accepted, dataset.QA.Rejected, dataset.QA.SensitiveCount, dataset.QA.RoadkillCount)
	return err
}

func ingestMDTCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("ingest mdt", flag.ContinueOnError)
	fs.SetOutput(stderr)

	filePath := fs.String("file", "", "Path to MDT carcass CSV file (required)")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *filePath == "" {
		return errors.New("usage: corridorctl ingest mdt --file <path>")
	}

	// #nosec G304 -- operator-specified file path for CLI ingestion
	file, err := os.Open(*filePath)
	if err != nil {
		return fmt.Errorf("open mdt file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()

	records, err := ingest.ParseMDTCarcassCSV(file)
	if err != nil {
		return fmt.Errorf("parse mdt dataset: %w", err)
	}

	speciesCounts := make(map[string]int64)
	var totalAnimals int64
	routes := make(map[string]bool)
	for _, r := range records {
		sp := r.NormalizedSpecies
		if sp == "" {
			sp = r.AnimalType
		}
		if sp != "" {
			speciesCounts[sp] += int64(r.Quantity)
		}
		qty := int64(r.Quantity)
		if qty <= 0 {
			qty = 1
		}
		totalAnimals += qty
		if r.Route != "" {
			routes[r.Route] = true
		}
	}

	out := struct {
		Source        string           `json:"source"`
		RecordsRead   int              `json:"records_read"`
		Accepted      int              `json:"accepted"`
		TotalAnimals  int64            `json:"total_animals"`
		UniqueRoutes  int              `json:"unique_routes"`
		SpeciesCounts map[string]int64 `json:"species_counts,omitempty"`
		SpatialRef    string           `json:"spatial_reference"`
	}{
		Source:        "Montana Department of Transportation Statewide Carcass Collection",
		RecordsRead:   len(records),
		Accepted:      len(records),
		TotalAnimals:  totalAnimals,
		UniqueRoutes:  len(routes),
		SpeciesCounts: speciesCounts,
		SpatialRef:    "Montana State Plane NAD83 (EPSG:32100) transformed to WGS84 (EPSG:4326)",
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	_, err = fmt.Fprintf(stdout, "Ingested %d MDT carcass records (%d animals across %d routes).\n",
		len(records), totalAnimals, len(routes))
	return err
}

func ingestWYDOTCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	_ = ctx
	records, err := readFeatureFile(args, stderr, "wydot", ingest.ParseWYDOTRoadkill)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	public := 0
	withheld := 0
	now := time.Now().UTC()
	for _, rec := range records {
		if _, ok := rec.PublicRecord(now); ok {
			public++
		} else if rec.IsSensitive {
			withheld++
		}
	}
	return writeFeatureSummary(stdout, argsJSON(args), featureSummary{
		Source:            ingest.WYDOTRoadkillArtifact().Name,
		License:           ingest.WYDOTRoadkillArtifact().License,
		RecordsRead:       len(records),
		PublicRecords:     public,
		WithheldSensitive: withheld,
	})
}

func ingestUDOTCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	_ = ctx
	records, err := readFeatureFile(args, stderr, "udot", ingest.ParseUDOTCrashes)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wild := 0
	for _, rec := range records {
		if rec.WildAnimal() {
			wild++
		}
	}
	return writeFeatureSummary(stdout, argsJSON(args), featureSummary{
		Source:        ingest.UDOTCrashLocationsArtifact().Name,
		License:       ingest.UDOTCrashLocationsArtifact().License,
		RecordsRead:   len(records),
		PublicRecords: len(records),
		WildAnimal:    wild,
	})
}

func ingestIDFGCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	_ = ctx
	records, err := readFeatureFile(args, stderr, "idfg", ingest.ParseIDFGRoadkill)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	public := 0
	withheld := 0
	now := time.Now().UTC()
	for _, rec := range records {
		if _, ok := rec.PublicRecord(now); ok {
			public++
		} else if rec.IsSensitive {
			withheld++
		}
	}
	return writeFeatureSummary(stdout, argsJSON(args), featureSummary{
		Source:            ingest.IDFGRoadkillArtifact().Name,
		License:           ingest.IDFGRoadkillArtifact().License,
		RecordsRead:       len(records),
		PublicRecords:     public,
		WithheldSensitive: withheld,
	})
}

type featureSummary struct {
	Source            string `json:"source"`
	License           string `json:"license"`
	RecordsRead       int    `json:"records_read"`
	PublicRecords     int    `json:"public_records"`
	WithheldSensitive int    `json:"withheld_sensitive,omitempty"`
	WildAnimal        int    `json:"wild_animal,omitempty"`
}

func argsJSON(args []string) bool {
	for _, arg := range args {
		if arg == "--json" || arg == "-json" {
			return true
		}
	}
	return false
}

func writeFeatureSummary(stdout io.Writer, jsonOutput bool, summary featureSummary) error {
	if jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}
	_, err := fmt.Fprintf(stdout, "Parsed %d %s records (%d public, %d sensitive withheld).\n",
		summary.RecordsRead, summary.Source, summary.PublicRecords, summary.WithheldSensitive)
	return err
}

func readFeatureFile[T any](args []string, stderr io.Writer, name string, parse func(io.Reader) ([]T, error)) ([]T, error) {
	fs := flag.NewFlagSet("ingest "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	filePath := fs.String("file", "", "Path to an ArcGIS feature JSON page")
	_ = fs.Bool("json", false, "Output a count summary as JSON")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *filePath == "" {
		return nil, fmt.Errorf("usage: corridorctl ingest %s --file <path>", name)
	}
	// #nosec G304 -- operator-specified file path for CLI ingestion
	file, err := os.Open(*filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s file: %w", name, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	records, err := parse(file)
	if err != nil {
		return nil, fmt.Errorf("parse %s dataset: %w", name, err)
	}
	return records, nil
}

func ingestCaltransCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("ingest caltrans", flag.ContinueOnError)
	fs.SetOutput(stderr)

	filePath := fs.String("file", "", "Path to Caltrans CSV or GeoJSON file")
	urlPath := fs.String("url", "", "HTTP URL to Caltrans FeatureServer or CSV export")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *filePath == "" && *urlPath == "" {
		return errors.New("usage: corridorctl ingest caltrans --file <path> | --url <url>")
	}

	var reader io.Reader
	var sourceDesc string

	if *filePath != "" {
		// #nosec G304 -- operator-specified file path for CLI ingestion
		f, err := os.Open(*filePath)
		if err != nil {
			return fmt.Errorf("open caltrans file: %w", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				slog.Warn("resource cleanup failed", "error", err)
			}
		}()
		reader = f
		sourceDesc = *filePath
	} else {
		// Fetch from HTTP URL
		if !strings.HasPrefix(*urlPath, "http://") && !strings.HasPrefix(*urlPath, "https://") {
			return errors.New("invalid url: must start with http:// or https://")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, *urlPath, nil)
		if err != nil {
			return fmt.Errorf("build caltrans request: %w", err)
		}
		req.Header.Set("User-Agent", "corridorctl/1.0 (wildlife-connectivity-platform)")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("fetch caltrans url: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				slog.Warn("resource cleanup failed", "error", err)
			}
		}()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("fetch caltrans url: HTTP status %d", resp.StatusCode)
		}
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, io.LimitReader(resp.Body, 64<<20)); err != nil {
			return fmt.Errorf("read caltrans response: %w", err)
		}
		reader = buf
		sourceDesc = *urlPath
	}

	records, err := ingest.ParseCaltransAADT(reader)
	if err != nil {
		return fmt.Errorf("parse caltrans dataset: %w", err)
	}

	routes := make(map[string]bool)
	var totalAADT float64
	for _, r := range records {
		if r.Route != "" {
			routes[r.Route] = true
		}
		totalAADT += r.AverageBidirectionalAADT()
	}

	avgAADT := 0.0
	if len(records) > 0 {
		avgAADT = totalAADT / float64(len(records))
	}

	out := struct {
		Source       string  `json:"source"`
		Input        string  `json:"input"`
		RecordsRead  int     `json:"records_read"`
		Accepted     int     `json:"accepted"`
		UniqueRoutes int     `json:"unique_routes"`
		AverageAADT  float64 `json:"average_aadt"`
		Status       string  `json:"status"`
	}{
		Source:       "Caltrans Annual Average Daily Traffic (AADT)",
		Input:        sourceDesc,
		RecordsRead:  len(records),
		Accepted:     len(records),
		UniqueRoutes: len(routes),
		AverageAADT:  math.Round(avgAADT*10) / 10,
		Status:       "accepted",
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	_, err = fmt.Fprintf(stdout, "Ingested %d Caltrans AADT records across %d highway routes (average AADT: %.0f).\n",
		len(records), len(routes), avgAADT)
	return err
}

func ingestAll(ctx context.Context, out io.Writer) error {
	return ingestInputs(ctx, out, "data/raw/sweep-2026-09-21/global-roadkill-v5.csv", ingest.PilotArtifact(), ingestMovementInputs)
}
func ingestInputs(ctx context.Context, out io.Writer, path string, artifact ingest.Artifact, movement func(context.Context, *pgxpool.Pool, *storage.Client) error) error {
	// #nosec G304 -- fixed operator path in production; dependency injection for isolated fixtures.
	file, err := os.Open(path)
	if err != nil {
		return errors.New("pilot artifact missing; see data/SOURCES.md acquisition instructions")
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	if err = ingest.Verify(file, artifact.SHA256); err != nil {
		return err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind source: %w", err)
	}
	dataset, err := ingest.ParseRoadkill(file)
	if err != nil {
		return err
	}
	if len(dataset.Records) == 0 {
		return errors.New("pilot has no accepted records")
	}
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	cfg.S3AccessKey = os.Getenv("MINIO_ROOT_USER")
	cfg.S3SecretKey = os.Getenv("MINIO_ROOT_PASSWORD")
	objects, err := storage.New(cfg)
	if err != nil {
		return errors.New("ingestion requires administrator storage credentials")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind archive: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("inspect artifact: %w", err)
	}
	if err = objects.Archive(ctx, artifact.ObjectKey, file, stat.Size()); err != nil {
		return err
	}
	dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
	if dsn == "" {
		return errors.New("ingestion requires CORRIDOR_MIGRATION_DATABASE_URL")
	}
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err = ingest.Import(ctx, pool, dataset, artifact); err != nil {
		return err
	}
	if err = movement(ctx, pool, objects); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Source  string    `json:"source"`
		QA      ingest.QA `json:"qa"`
		Skipped []string  `json:"skipped"`
	}{artifact.Name, dataset.QA, []string{"Montana: rights unresolved", "California derivative: quarantined", "BC WARS: redistribution restricted", "telemetry: CRS/time validation pending", "other research candidates: no adapter or artifact-level clearance"}})
}

func ingestMigration(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client) error {
	const path = "data/raw/exploration/pequop.shp"
	file, err := os.Open(path)
	if err != nil {
		return errors.New("migration shape missing; see data/SOURCES.md")
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	a := ingest.MigrationArtifact()
	if err = ingest.Verify(file, a.SHA256); err != nil {
		return err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if err = objects.Archive(ctx, a.ObjectKey, file, stat.Size()); err != nil {
		return err
	}
	_, err = ingest.ImportMigration(ctx, pool, path)
	return err
}
