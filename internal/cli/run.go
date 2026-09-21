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

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 2 && args[0] == "ingest" && args[1] == "all" {
		return ingestAll(ctx, stdout)
	}
	if len(args) != 1 {
		return errors.New("usage: corridorctl help|init-dev|migrate|setup|health|ingest all")
	}
	switch args[0] {
	case "help":
		_, err := fmt.Fprintln(stdout, "corridorctl help|init-dev|migrate|setup|health|ingest all")
		return err
	case "init-dev":
		return initDev(".env", rand.Reader, stdout)
	case "migrate":
		dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
		if dsn == "" {
			return errors.New("CORRIDOR_MIGRATION_DATABASE_URL required")
		}
		if err := db.Migrate(ctx, dsn); err != nil {
			return errors.New("migration failed; verify database availability and required extensions")
		}
		return nil
	case "setup":
		return setup(ctx)
	case "health":
		return health(ctx)
	default:
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
