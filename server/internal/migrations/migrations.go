// Package migrations runs the SQL migrations embedded in the executable.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed sql/*.sql
var files embed.FS

var ErrBaselineRequired = errors.New("existing database requires baselining: run migrate baseline first; do not replay the initial schema or delete the database volume")

// Up applies pending migrations. Existing databases without a recorded initial
// migration must be baselined separately.
func Up(ctx context.Context, db *sql.DB) error {
	fsys, err := fs.Sub(files, "sql")
	if err != nil {
		return err
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys,
		goose.WithSessionLocker(locker), goose.WithVerbose(true))
	if err != nil {
		return err
	}
	legacy, err := requiresBaseline(ctx, db)
	if err != nil {
		return err
	}
	if legacy {
		return ErrBaselineRequired
	}
	_, err = provider.Up(ctx)
	return err
}

// requiresBaseline detects an old init/database.sql database even if a prior Goose
// invocation created a tracking table without applying migration 1.
func requiresBaseline(ctx context.Context, db *sql.DB) (bool, error) {
	var tracked, hasTables bool
	err := db.QueryRowContext(ctx, `SELECT
		EXISTS (SELECT 1 FROM pg_catalog.pg_tables
			WHERE schemaname = current_schema() AND tablename = 'goose_db_version'),
		EXISTS (SELECT 1 FROM pg_catalog.pg_tables
			WHERE schemaname = current_schema() AND tablename <> 'goose_db_version')`).Scan(&tracked, &hasTables)
	if err != nil {
		return false, fmt.Errorf("inspect database: %w", err)
	}
	if !hasTables {
		return false, nil
	}
	if !tracked {
		return true, nil
	}
	var applied bool
	err = db.QueryRowContext(ctx, `SELECT COALESCE(
		(SELECT is_applied FROM goose_db_version WHERE version_id = 1 ORDER BY id DESC LIMIT 1),
		false)`).Scan(&applied)
	if err != nil {
		return false, fmt.Errorf("inspect initial migration: %w", err)
	}
	return !applied, nil
}
