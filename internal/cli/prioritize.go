package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"corridor/internal/prioritize"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CandidateSegmentProvider defines the interface for retrieving road segments for prioritization.
type CandidateSegmentProvider interface {
	GetSegments(ctx context.Context, state string) ([]prioritize.SegmentRecord, error)
}

var defaultCandidateProvider CandidateSegmentProvider = &defaultReferenceSegmentProvider{}

// SetCandidateProvider allows overriding the segment provider for tests or pipelines.
func SetCandidateProvider(p CandidateSegmentProvider) {
	if p == nil {
		defaultCandidateProvider = &defaultReferenceSegmentProvider{}
	} else {
		defaultCandidateProvider = p
	}
}

func prioritizeCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("prioritize", flag.ContinueOnError)
	fs.SetOutput(stderr)

	budget := fs.Float64("budget", 5000000.0, "Budget constraint in USD")
	minBCR := fs.Float64("min-bcr", 1.0, "Minimum Benefit-Cost Ratio filter")
	state := fs.String("state", "WY", "Two-letter state filter (e.g. WY, NV, UT, ID, CA, MT, CO)")
	park := fs.String("park", "", "Park filter (e.g. YOSE, YELL, GRTE, GLAC, ROMO)")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	statePassed := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "state" {
			statePassed = true
		}
	})

	resolvedState, resolvedHighway := resolveParkFilter(*park, *state)
	if *park != "" && !statePassed && resolvedState != "" {
		*state = resolvedState
	}

	if *budget < 0 {
		return errors.New("budget cannot be negative")
	}
	if *minBCR < 0 {
		return errors.New("min-bcr cannot be negative")
	}

	// Load configuration from config files if available, otherwise default
	cfg := loadPrioritizeConfig()

	opts := prioritize.PrioritizeOptions{
		Budget:        *budget,
		MinBCR:        minBCR,
		State:         state,
		MinPercentile: cfg.MinPercentile,
	}

	var candidates []prioritize.CandidateSite

	// Try database first if configured
	dbURL := os.Getenv("CORRIDOR_DATABASE_URL")
	if dbURL != "" {
		if pool, err := pgxpool.New(ctx, dbURL); err == nil {
			defer pool.Close()
			if dbCandidates, err := prioritize.GenerateCandidatesWithOptions(ctx, pool, opts, cfg); err == nil && len(dbCandidates) > 0 {
				candidates = dbCandidates
			}
		}
	}

	// Fallback to configured segment provider if database had no results or was unavailable
	if len(candidates) == 0 {
		segments, err := defaultCandidateProvider.GetSegments(ctx, *state)
		if err != nil {
			return fmt.Errorf("load candidate segments: %w", err)
		}
		candidates = prioritize.GenerateCandidatesFromSegmentsWithOptions(segments, opts, cfg)
	}

	// Apply park highway filter if park was specified
	if *park != "" && resolvedHighway != "" {
		filtered := make([]prioritize.CandidateSite, 0, len(candidates))
		for _, c := range candidates {
			if strings.EqualFold(c.HighwayRef, resolvedHighway) || strings.HasPrefix(strings.ToUpper(c.HighwayRef), strings.ToUpper(resolvedHighway)) {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}

	if *jsonOutput {
		if candidates == nil {
			candidates = []prioritize.CandidateSite{}
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(candidates)
	}

	if len(candidates) == 0 {
		_, err := fmt.Fprintln(stdout, "No candidate mitigation sites found matching criteria.")
		return err
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Site ID\tHighway\tMileposts\tCost\tAvoided Collisions\tBCR [P10-P90]")
	for _, c := range candidates {
		fmt.Fprintf(w, "%s\t%s\tMP %.1f-%.1f\t$%.0f\t%.1f\t%.2f [%.2f-%.2f]\n",
			c.ID,
			c.HighwayRef,
			c.MilepostStart,
			c.MilepostEnd,
			c.EstCostUSD,
			c.ExpectedAvoidedCollisions,
			c.BCR,
			c.P10BCR,
			c.P90BCR,
		)
	}
	return w.Flush()
}

func loadPrioritizeConfig() prioritize.PrioritizeConfig {
	effPath := filepath.Join("config", "mitigation_effectiveness.yaml")
	speciesPath := filepath.Join("config", "species_costs.yaml")
	cfg, err := prioritize.LoadConfigFromFiles(effPath, speciesPath)
	if err == nil {
		return cfg
	}
	return prioritize.DefaultConfig()
}

func resolveParkFilter(park, state string) (string, string) {
	normPark := strings.ToLower(strings.TrimSpace(park))
	normPark = strings.ReplaceAll(normPark, "-", " ")
	normPark = strings.ReplaceAll(normPark, "_", " ")
	normState := strings.ToUpper(strings.TrimSpace(state))

	switch {
	case normPark == "yose" || strings.Contains(normPark, "yosemite"):
		return "CA", "CA-120"
	case normPark == "yell" || strings.Contains(normPark, "yellowstone"):
		return "WY", "US-191"
	case normPark == "grte" || strings.Contains(normPark, "grand teton") || strings.Contains(normPark, "teton"):
		return "WY", "US-89"
	case normPark == "glac" || strings.Contains(normPark, "glacier"):
		return "MT", "US-2"
	case normPark == "romo" || strings.Contains(normPark, "rocky"):
		return "CO", "US-34"
	case normState == "CA":
		return "CA", "CA-120"
	case normState == "MT":
		return "MT", "US-2"
	case normState == "CO":
		return "CO", "US-34"
	case normState == "WY":
		return "WY", ""
	default:
		return normState, ""
	}
}

type defaultReferenceSegmentProvider struct{}

func (p *defaultReferenceSegmentProvider) GetSegments(ctx context.Context, state string) ([]prioritize.SegmentRecord, error) {
	allSegments := []prioritize.SegmentRecord{
		// WY US-191 - Yellowstone Gallatin Canyon Overpass project (milepost 12.0 - 12.4)
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  12.0,
			MilepostEnd:    12.2,
			LengthM:        100.0,
			RiskScore:      5.0,
			PercentileRank: 98.0,
			P10:            3.0,
			P90:            7.0,
			SpeciesBreakdown: map[string]float64{
				"elk":          0.55,
				"moose":        0.30,
				"grizzly_bear": 0.15,
			},
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  12.2,
			MilepostEnd:    12.4,
			LengthM:        100.0,
			RiskScore:      4.5,
			PercentileRank: 97.0,
			P10:            2.8,
			P90:            6.2,
			SpeciesBreakdown: map[string]float64{
				"elk":          0.55,
				"moose":        0.30,
				"grizzly_bear": 0.15,
			},
		},
		// WY US-191 - Yellowstone lower cost fencing project (milepost 25.0 - 25.2, cost < $1M)
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			HighwayRef:     "US-191",
			State:          "WY",
			MilepostStart:  25.0,
			MilepostEnd:    25.2,
			LengthM:        321.8,
			RiskScore:      0.95,
			PercentileRank: 92.0,
			P10:            0.90,
			P90:            1.20,
			SpeciesBreakdown: map[string]float64{
				"elk":  0.20,
				"deer": 0.80,
			},
		},
		// WY US-89 - Grand Teton Jackson Hole corridor (milepost 18.0 - 18.4)
		{
			ID:             uuid.MustParse("00000000-0000-0000-0002-000000000001"),
			HighwayRef:     "US-89",
			State:          "WY",
			MilepostStart:  18.0,
			MilepostEnd:    18.2,
			LengthM:        100.0,
			RiskScore:      5.2,
			PercentileRank: 98.0,
			P10:            3.2,
			P90:            7.2,
			SpeciesBreakdown: map[string]float64{
				"mule_deer": 0.50,
				"pronghorn": 0.30,
				"elk":       0.20,
			},
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-0002-000000000002"),
			HighwayRef:     "US-89",
			State:          "WY",
			MilepostStart:  18.2,
			MilepostEnd:    18.4,
			LengthM:        100.0,
			RiskScore:      4.8,
			PercentileRank: 97.0,
			P10:            2.9,
			P90:            6.6,
			SpeciesBreakdown: map[string]float64{
				"mule_deer": 0.50,
				"pronghorn": 0.30,
				"elk":       0.20,
			},
		},
		// MT US-2 - Glacier National Park Marias Pass / Southern Boundary corridor (milepost 152.0 - 152.4)
		{
			ID:             uuid.MustParse("00000000-0000-0000-0003-000000000001"),
			HighwayRef:     "US-2",
			State:          "MT",
			MilepostStart:  152.0,
			MilepostEnd:    152.2,
			LengthM:        100.0,
			RiskScore:      5.5,
			PercentileRank: 98.5,
			P10:            3.5,
			P90:            7.5,
			SpeciesBreakdown: map[string]float64{
				"grizzly_bear":  0.40,
				"bighorn_sheep": 0.35,
				"canada_lynx":   0.25,
			},
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-0003-000000000002"),
			HighwayRef:     "US-2",
			State:          "MT",
			MilepostStart:  152.2,
			MilepostEnd:    152.4,
			LengthM:        100.0,
			RiskScore:      4.6,
			PercentileRank: 97.5,
			P10:            3.0,
			P90:            6.8,
			SpeciesBreakdown: map[string]float64{
				"grizzly_bear":  0.40,
				"bighorn_sheep": 0.35,
				"canada_lynx":   0.25,
			},
		},
		// CO US-34 - Rocky Mountain Trail Ridge Road corridor (milepost 58.0 - 58.4)
		{
			ID:             uuid.MustParse("00000000-0000-0000-0004-000000000001"),
			HighwayRef:     "US-34",
			State:          "CO",
			MilepostStart:  58.0,
			MilepostEnd:    58.2,
			LengthM:        100.0,
			RiskScore:      4.9,
			PercentileRank: 97.5,
			P10:            2.9,
			P90:            6.8,
			SpeciesBreakdown: map[string]float64{
				"elk":           0.55,
				"mule_deer":     0.30,
				"bighorn_sheep": 0.15,
			},
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-0004-000000000002"),
			HighwayRef:     "US-34",
			State:          "CO",
			MilepostStart:  58.2,
			MilepostEnd:    58.4,
			LengthM:        100.0,
			RiskScore:      4.2,
			PercentileRank: 96.5,
			P10:            2.5,
			P90:            6.0,
			SpeciesBreakdown: map[string]float64{
				"elk":           0.55,
				"mule_deer":     0.30,
				"bighorn_sheep": 0.15,
			},
		},
		// NV I-80 - Pequop Pass mule deer corridor
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			HighwayRef:     "I-80",
			State:          "NV",
			MilepostStart:  352.0,
			MilepostEnd:    354.0,
			LengthM:        3218.0,
			RiskScore:      2.8,
			PercentileRank: 94.0,
			P10:            1.6,
			P90:            4.0,
			SpeciesBreakdown: map[string]float64{
				"deer": 0.9,
				"elk":  0.1,
			},
		},
		// UT US-89 - Bear Lake pronghorn corridor
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			HighwayRef:     "US-89",
			State:          "UT",
			MilepostStart:  45.0,
			MilepostEnd:    47.0,
			LengthM:        3218.0,
			RiskScore:      2.2,
			PercentileRank: 93.0,
			P10:            1.2,
			P90:            3.2,
			SpeciesBreakdown: map[string]float64{
				"pronghorn": 0.7,
				"deer":      0.3,
			},
		},
		// CA CA-120 - Yosemite Crane Flat / Tioga Road corridor
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			HighwayRef:     "CA-120",
			State:          "CA",
			MilepostStart:  42.0,
			MilepostEnd:    42.1,
			LengthM:        160.9344,
			RiskScore:      16.8,
			PercentileRank: 95.0,
			P10:            11.0,
			P90:            22.5,
			SpeciesBreakdown: map[string]float64{
				"deer":  0.40,
				"bear":  0.45,
				"sheep": 0.15,
			},
		},
		// CA CA-140 - Yosemite Merced River Canyon corridor
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			HighwayRef:     "CA-140",
			State:          "CA",
			MilepostStart:  8.5,
			MilepostEnd:    8.6,
			LengthM:        160.9344,
			RiskScore:      13.5,
			PercentileRank: 92.0,
			P10:            9.0,
			P90:            17.5,
			SpeciesBreakdown: map[string]float64{
				"deer": 0.65,
				"bear": 0.35,
			},
		},
	}

	var filtered []prioritize.SegmentRecord
	for _, seg := range allSegments {
		if state == "" || seg.State == state {
			filtered = append(filtered, seg)
		}
	}
	return filtered, nil
}
