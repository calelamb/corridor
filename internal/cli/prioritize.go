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
	state := fs.String("state", "WY", "Two-letter state filter (e.g. WY, NV, UT, ID)")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
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

type defaultReferenceSegmentProvider struct{}

func (p *defaultReferenceSegmentProvider) GetSegments(ctx context.Context, state string) ([]prioritize.SegmentRecord, error) {
	allSegments := []prioritize.SegmentRecord{
		// WY US-191 - Major Overpass project (milepost 12.0 - 12.4)
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
				"elk":  0.7,
				"deer": 0.3,
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
				"elk":  0.7,
				"deer": 0.3,
			},
		},
		// WY US-191 - Lower cost fencing project (milepost 25.0 - 25.2, cost < $1M)
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
	}

	var filtered []prioritize.SegmentRecord
	for _, seg := range allSegments {
		if state == "" || seg.State == state {
			filtered = append(filtered, seg)
		}
	}
	return filtered, nil
}
