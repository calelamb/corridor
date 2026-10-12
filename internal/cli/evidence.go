package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"corridor/internal/explore"
)

func evidenceCmd(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("evidence", flag.ContinueOnError)
	fs.SetOutput(stderr)

	category := fs.String("category", "all", "Category filter: all, mitigation, bias, ecology, economics")
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	catalog := explore.GetEvidenceCatalog()

	var filtered []explore.AcademicPaper
	catFilter := strings.TrimSpace(strings.ToLower(*category))

	for _, paper := range catalog {
		if catFilter == "" || catFilter == "all" || strings.ToLower(paper.Category) == catFilter {
			filtered = append(filtered, paper)
		}
	}

	if *jsonOutput {
		if filtered == nil {
			filtered = []explore.AcademicPaper{}
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}

	if len(filtered) == 0 {
		_, err := fmt.Fprintf(stdout, "No academic evidence found matching category %q.\n", *category)
		return err
	}

	for i, p := range filtered {
		fmt.Fprintf(stdout, "[%d] %s (%d)\n", i+1, p.Title, p.Year)
		fmt.Fprintf(stdout, "    Authors:     %s\n", p.Authors)
		fmt.Fprintf(stdout, "    Journal:     %s\n", p.Journal)
		fmt.Fprintf(stdout, "    DOI:         %s\n", p.DOI)
		fmt.Fprintf(stdout, "    Category:    %s\n", p.Category)
		fmt.Fprintf(stdout, "    Key Findings:\n      %s\n", p.KeyFindings)
		fmt.Fprintf(stdout, "    Corridor Application:\n      %s\n", p.AppliedInCorridor)
		if i < len(filtered)-1 {
			fmt.Fprintln(stdout)
		}
	}

	return nil
}
