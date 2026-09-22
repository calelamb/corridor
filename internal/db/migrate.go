package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies versioned schema changes transactionally with administrator credentials.
func Migrate(ctx context.Context, dsn string) error {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open migration connection: invalid configuration")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	files, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("migration files: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, conn, files)
	if err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply schema migrations: %w", err)
	}
	return nil
}
