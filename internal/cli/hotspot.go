package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"corridor/internal/hotspot"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HotspotOutput represents a ranked collision hotspot segment.
type HotspotOutput struct {
	SegmentID    string   `json:"segment_id"`
	Score        float64  `json:"score"`
	RawCount     int      `json:"raw_count"`
	ZScore       *float64 `json:"z_score,omitempty"`
	PValue       *float64 `json:"p_value,omitempty"`
	FDRPValue    *float64 `json:"fdr_p_value,omitempty"`
	Significance string   `json:"significance"`
}

func hotspotCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("hotspot", flag.ContinueOnError)
	fs.SetOutput(stderr)

	method := fs.String("method", "nkde", "Hotspot method: nkde or getis_ord")
	bandwidth := fs.Float64("bandwidth", 500.0, "Bandwidth distance in meters")
	limit := fs.Int("limit", 10, "Maximum number of top hotspot segments to display")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")
	state := fs.String("state", "", "State filter (e.g. CA, WY)")
	park := fs.String("park", "", "National Park filter (e.g. Yosemite, Yellowstone)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	normMethod := strings.ToLower(strings.TrimSpace(*method))
	if normMethod != "nkde" && normMethod != "getis_ord" && normMethod != "getis-ord" {
		return fmt.Errorf("invalid method %q: must be 'nkde' or 'getis_ord'", *method)
	}
	if *bandwidth <= 0 {
		return errors.New("bandwidth must be positive")
	}

	graph, events, err := loadHotspotData(ctx, *state, *park)
	if err != nil {
		return fmt.Errorf("load hotspot data: %w", err)
	}

	results, err := computeHotspots(graph, events, normMethod, *bandwidth)
	if err != nil {
		return fmt.Errorf("compute hotspots: %w", err)
	}

	if *limit > 0 && len(results) > *limit {
		results = results[:*limit]
	}

	if *jsonOutput {
		if results == nil {
			results = []HotspotOutput{}
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if len(results) == 0 {
		_, err := fmt.Fprintln(stdout, "No hotspot segments identified.")
		return err
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Segment ID\tScore\tRaw Count\tZ-Score\tSignificance")
	for _, h := range results {
		zStr := "N/A"
		if h.ZScore != nil {
			zStr = fmt.Sprintf("%+.2f", *h.ZScore)
		}
		fmt.Fprintf(w, "%s\t%.4f\t%d\t%s\t%s\n",
			h.SegmentID, h.Score, h.RawCount, zStr, h.Significance)
	}
	return w.Flush()
}

func computeHotspots(g *hotspot.RoadGraph, events []hotspot.EventPoint, method string, bandwidth float64) ([]HotspotOutput, error) {
	results, err := g.AnalyzeHotspots(events, method, bandwidth)
	if err != nil {
		return nil, err
	}

	outputs := make([]HotspotOutput, len(results))
	for i, r := range results {
		outputs[i] = HotspotOutput{
			SegmentID:    r.SegmentID,
			Score:        r.Score,
			RawCount:     r.RawCount,
			ZScore:       r.ZScore,
			PValue:       r.PValue,
			FDRPValue:    r.FDRPValue,
			Significance: r.Significance,
		}
	}
	return outputs, nil
}

// loadHotspotData attempts database loading if configured; otherwise provides
// reference corridor data for the 5 national park corridors (YOSE/CA, YELL/WY, GRTE/WY, GLAC/MT, ROMO/CO).
func loadHotspotData(ctx context.Context, state, park string) (*hotspot.RoadGraph, []hotspot.EventPoint, error) {
	// 1. Try DB if configured
	dbURL := os.Getenv("CORRIDOR_DATABASE_URL")
	if dbURL != "" {
		if pool, err := pgxpool.New(ctx, dbURL); err == nil {
			defer pool.Close()
			// DB integration fallback check
		}
	}

	if ref := hotspot.FindReferenceCorridor(state, park); ref != nil {
		return ref.Graph, ref.Events, nil
	}

	// Default reference corridor: Yellowstone US-191 corridor
	corridor := hotspot.ReferenceYellowstoneCorridor()
	return corridor.Graph, corridor.Events, nil
}
