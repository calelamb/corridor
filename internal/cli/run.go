package cli

import (
	"context"
	"corridor/internal/config"
	"corridor/internal/db"
	"corridor/internal/storage"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
)

const helpUsage = `corridorctl - Corridor Wildlife Collision & Connectivity Platform CLI

Usage:
  corridorctl <command> [flags]

Commands:
  help|--help                     Show this help message
  init-dev                        Initialize local .env development secrets
  migrate                         Execute database schema migrations
  setup                           Provision database roles and private storage buckets
  health                          Verify API service connectivity and readiness
  ingest <subcommand> [flags]     Execute ingestion pipeline or source-specific parser
  prioritize [flags]              Compute benefit-cost ratio mitigation site prioritization
  hotspot [flags]                 Identify collision hotspots using NKDE or Getis-Ord Gi*
  evidence [flags]                Query peer-reviewed scientific evidence and research catalog
  seed-segments                   Populate benchmark road segments and risk scores
  outreach [flags]                Generate federal outreach email drafts, briefs, and CRM ledger

Subcommands for ingest:
  all                             Execute full ingestion pipeline
  irma --file <path> [--park <code>] [--json]
                                  Parse NPS IRMA wildlife observation and mortality records
  mdt --file <path> [--json]      Parse MDT Montana statewide carcass records
  caltrans (--file <path> | --url <url>) [--json]
                                  Parse Caltrans AADT and highway postmile records

Flags for prioritize:
  --budget <float>                Budget limit in USD (default: 5000000)
  --min-bcr <float>               Minimum Benefit-Cost Ratio filter (default: 1.0)
  --state <string>                Two-letter state filter (default: "WY")
  --park <string>                 Park filter (e.g. YOSE, YELL, GRTE, GLAC, ROMO)
  --json                          Output results as JSON array

Flags for hotspot:
  --method <string>               Algorithm: "nkde" or "getis_ord" (default: "nkde")
  --bandwidth <float>             Bandwidth in meters (default: 500.0)
  --limit <int>                   Maximum top segments to display (default: 10)
  --state <string>                Two-letter state filter (e.g. CA, WY, MT, CO)
  --park <string>                 National Park filter (e.g. YOSE, YELL, GRTE, GLAC, ROMO)
  --json                          Output results as JSON array

Flags for evidence:
  --category <string>             Category filter: "all", "mitigation", "bias", "ecology", "economics" (default: "all")
  --json                          Output results as JSON array

Flags for outreach:
  --park <name>                   Park filter ("Yellowstone", "Grand Teton", ..., or "all", default: "all")
  --persona <role>                Target role filter ("biologist", "ranger", "superintendent", "all", default: "all")
  --format <format>               Format: "all", "eml", "brief", "csv" (default: "all")
  --out <dir>                     Output directory (default: "docs/outreach/generated")`

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stdout, helpUsage)
		return errors.New("usage: corridorctl help|init-dev|migrate|setup|health|ingest <all|irma|mdt|caltrans>|prioritize|hotspot|evidence|seed-segments|outreach")
	}

	switch args[0] {
	case "help", "--help", "-h":
		_, err := fmt.Fprintln(stdout, helpUsage)
		return err
	case "init-dev":
		if len(args) != 1 {
			return errors.New("usage: corridorctl init-dev")
		}
		return initDev(".env", rand.Reader, stdout)
	case "migrate":
		if len(args) != 1 {
			return errors.New("usage: corridorctl migrate")
		}
		dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
		if dsn == "" {
			return errors.New("CORRIDOR_MIGRATION_DATABASE_URL required")
		}
		if err := db.Migrate(ctx, dsn); err != nil {
			return errors.New("migration failed; verify database availability and required extensions")
		}
		return nil
	case "setup":
		if len(args) != 1 {
			return errors.New("usage: corridorctl setup")
		}
		return setup(ctx)
	case "health":
		if len(args) != 1 {
			return errors.New("usage: corridorctl health")
		}
		return health(ctx)
	case "ingest":
		return ingestCmd(ctx, args[1:], stdout, stderr)
	case "prioritize":
		return prioritizeCmd(ctx, args[1:], stdout, stderr)
	case "hotspot":
		return hotspotCmd(ctx, args[1:], stdout, stderr)
	case "evidence":
		return evidenceCmd(ctx, args[1:], stdout, stderr)
	case "seed-segments":
		return seedSegmentsCmd(ctx, stdout)
	case "outreach":
		return outreachCmd(ctx, args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintln(stdout, helpUsage)
		return errors.New("unsupported command; use corridorctl help")
	}
}

func setup(ctx context.Context) error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	if err := db.Provision(ctx, os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL"), os.Getenv("CORRIDOR_DB_PASSWORD")); err != nil {
		return errors.New("database provisioning failed")
	}
	if err := storage.Setup(ctx, cfg, os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD")); err != nil {
		return errors.New("private storage provisioning failed")
	}
	return nil
}
