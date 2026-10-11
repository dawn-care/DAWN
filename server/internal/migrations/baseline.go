package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Baseline records migration 1 only after validating the existing schema.
// Goose runs validation and version recording in one transaction under its lock.
func Baseline(ctx context.Context, db *sql.DB) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, nil,
		goose.WithSessionLocker(locker),
		goose.WithGoMigrations(goose.NewGoMigration(1, &goose.GoFunc{RunTx: validateBaseline}, nil)))
	if err != nil {
		return err
	}
	_, err = provider.ApplyVersion(ctx, 1, true)
	if errors.Is(err, goose.ErrAlreadyApplied) {
		return nil
	}
	return err
}

func validateBaseline(ctx context.Context, tx *sql.Tx) error {
	initial, err := files.ReadFile("sql/00001_initial_schema.sql")
	if err != nil {
		return err
	}
	var schema string
	if err := tx.QueryRowContext(ctx, `SELECT current_schema()`).Scan(&schema); err != nil {
		return err
	}
	// This extracts only the fixed initial migration's CREATE TABLE statements,
	// never its seed inserts. Temporary reference tables disappear on commit/rollback.
	tables := regexp.MustCompile(`(?ms)^CREATE TABLE "([^"]+)" \(\n.*?^\);`).FindAllStringSubmatch(string(initial), -1)
	procedures := regexp.MustCompile(`(?s)CREATE OR REPLACE PROCEDURE (\w+)\((.*?)\)\s*LANGUAGE plpgsql\s*AS \$\$(.*?)\$\$;`).FindAllStringSubmatch(string(initial), -1)
	if len(tables) != 16 || len(procedures) != 2 {
		return errors.New("cannot validate the initial schema: unexpected SQL format")
	}
	for _, table := range tables {
		// Prevent schema changes while checking, without blocking normal row writes.
		name := pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier(table[1])
		if _, err := tx.ExecContext(ctx, "LOCK TABLE "+name+" IN ACCESS SHARE MODE"); err != nil {
			return fmt.Errorf("baseline requires table %s: %w", table[1], err)
		}
		ddl := strings.Replace(table[0], "CREATE TABLE", "CREATE TEMP TABLE", 1)
		ddl = strings.TrimSuffix(ddl, ");") + ") ON COMMIT DROP;"
		if _, err := tx.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("build reference for %s: %w", table[1], err)
		}
	}
	var mismatch string
	err = tx.QueryRowContext(ctx, baselineComparison, schema).Scan(&mismatch)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		return fmt.Errorf("baseline refused: table %s differs from migration 1", mismatch)
	}
	for _, procedure := range procedures {
		var types, names []string
		for _, arg := range strings.Split(procedure[2], ",") {
			fields := strings.Fields(arg)
			names = append(names, fields[0])
			types = append(types, fields[1])
		}
		signature := pq.QuoteIdentifier(schema) + "." + procedure[1] + "(" + strings.Join(types, ",") + ")"
		var matches bool
		err := tx.QueryRowContext(ctx, `SELECT prosrc = $2 AND proargnames = $3
			AND prokind = 'p' AND NOT prosecdef AND proconfig IS NULL
			AND prolang = (SELECT oid FROM pg_language WHERE lanname = 'plpgsql')
			FROM pg_proc WHERE oid = to_regprocedure($1)`, signature, procedure[3], pq.Array(names)).Scan(&matches)
		if err != nil || !matches {
			return fmt.Errorf("baseline refused: procedure %s is missing or differs from migration 1 (lookup error: %v)", procedure[1], err)
		}
	}
	return nil
}

// Compare columns, defaults, identity modes and constraints against temporary
// reference tables. Normalize only schema prefixes in default expressions.
const baselineComparison = `
WITH definitions AS (
  SELECT c.relnamespace, c.relname, c.relkind, c.relrowsecurity,
    (SELECT jsonb_agg(jsonb_build_array(a.attname, a.atttypid, a.atttypmod,
        a.attnotnull, a.attidentity, a.attgenerated,
        replace(pg_get_expr(d.adbin, d.adrelid),
          quote_ident((SELECT nspname FROM pg_namespace WHERE oid = c.relnamespace)) || '.', '')) ORDER BY a.attnum)
     FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
     WHERE a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped) AS columns,
    (SELECT jsonb_agg(def ORDER BY def) FROM (
       SELECT jsonb_build_array(k.contype, k.conkey, k.confkey, r.relname, r.relnamespace = c.relnamespace,
         k.confupdtype, k.confdeltype, k.confmatchtype, k.condeferrable,
         k.condeferred, k.convalidated, pg_get_expr(k.conbin, k.conrelid)) AS def
       FROM pg_constraint k LEFT JOIN pg_class r ON r.oid = k.confrelid
       WHERE k.conrelid = c.oid) constraints) AS constraints
  FROM pg_class c
  WHERE c.relnamespace IN (pg_my_temp_schema(), $1::regnamespace)
    AND c.relkind IN ('r', 'p')
)
SELECT reference.relname FROM definitions reference
LEFT JOIN definitions actual ON actual.relname = reference.relname
  AND actual.relnamespace = $1::regnamespace
WHERE reference.relnamespace = pg_my_temp_schema()
  AND ROW(reference.relkind, reference.relrowsecurity, reference.columns, reference.constraints)
      IS DISTINCT FROM ROW(actual.relkind, actual.relrowsecurity, actual.columns, actual.constraints)
LIMIT 1`
