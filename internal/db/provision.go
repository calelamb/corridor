package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Provision creates the public application login separately from schema migration.
func Provision(ctx context.Context, adminDSN, password string) error {
	if adminDSN == "" || len(password) < 32 {
		return errors.New("administrator connection and strong application password required")
	}
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return errors.New("administrator connection unavailable")
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin provisioning: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827364)"); err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT FROM pg_roles WHERE rolname='corridor_app')").Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := tx.Exec(ctx, "CREATE ROLE corridor_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION"); err != nil {
			return err
		}
	}
	// PostgreSQL utility statements cannot bind password parameters. quote_literal is server-side.
	var literal string
	if err := tx.QueryRow(ctx, "SELECT quote_literal($1::text)", password).Scan(&literal); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "ALTER ROLE corridor_app PASSWORD "+literal); err != nil {
		return errors.New("set application password failed")
	}
	if _, err := tx.Exec(ctx, "GRANT corridor_public TO corridor_app"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
