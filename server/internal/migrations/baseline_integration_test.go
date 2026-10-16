package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestBaselineIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	initial, err := files.ReadFile("sql/00001_initial_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	for i, test := range []struct {
		name, alter string
	}{
		{"matching", ""},
		{"empty", ""},
		{"missing_table", `DROP TABLE "NewsWidget"`},
		{"missing_column", `ALTER TABLE "Caregiver" DROP COLUMN role`},
		{"wrong_type", `ALTER TABLE "Caregiver" ALTER COLUMN role TYPE varchar(20)`},
		{"missing_constraint", `ALTER TABLE "Caregiver" DROP CONSTRAINT "Caregiver_google_id_key"`},
		{"wrong_default", `ALTER TABLE "Carepartner" ALTER COLUMN auto_start SET DEFAULT '09:00:00'`},
		{"missing_procedure", `DROP PROCEDURE delete_configurations(integer, integer[])`},
		{"changed_procedure", `CREATE OR REPLACE PROCEDURE delete_configurations(p_carepartner_id INT, id_list INT[]) LANGUAGE plpgsql AS $$ BEGIN NULL; END; $$`},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			schema := fmt.Sprintf("baseline_test_%d_%d", time.Now().UnixNano(), i)
			if _, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
				t.Fatal(err)
			}
			defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
			db, err := sql.Open("postgres", dsn+" search_path="+schema)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if test.name != "empty" {
				if _, err := db.Exec(strings.Split(string(initial), "-- +goose Down")[0]); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO "Caregiver" (name) VALUES ('Preserve this user')`); err != nil {
					t.Fatal(err)
				}
			}
			if test.alter != "" {
				if _, err := db.Exec(test.alter); err != nil {
					t.Fatal(err)
				}
			}
			before := snapshot(t, db, schema)
			err = Baseline(ctx, db)
			if (err == nil) != (test.name == "matching") {
				t.Fatalf("unexpected baseline result: %v", err)
			}
			var recorded int
			if err := db.QueryRow(`SELECT count(*) FROM goose_db_version WHERE version_id = 1 AND is_applied`).Scan(&recorded); err != nil {
				t.Fatal(err)
			}
			if test.name == "matching" {
				if recorded != 1 {
					t.Fatalf("expected one baseline record, got %d", recorded)
				}
				if err := Baseline(ctx, db); err != nil {
					t.Fatalf("repeated baseline: %v", err)
				}
			} else if recorded != 0 {
				t.Fatal("failed validation recorded migration 1")
			}
			if after := snapshot(t, db, schema); !reflect.DeepEqual(before, after) {
				t.Fatal("baseline changed application rows or sequence values")
			}
			if test.name == "matching" {
				provider, err := goose.NewProvider(goose.DialectPostgres, db, fstest.MapFS{
					"00001_initial_schema.sql": {Data: initial},
					"00002_test.sql":           {Data: []byte("-- +goose Up\nALTER TABLE \"Caregiver\" ADD COLUMN test_note text;")},
				})
				if err != nil {
					t.Fatal(err)
				}
				results, err := provider.Up(ctx)
				if err != nil || len(results) != 1 || results[0].Source.Version != 2 {
					t.Fatalf("upgrade after baseline: results=%v error=%v", results, err)
				}
				if err := Baseline(ctx, db); err != nil {
					t.Fatalf("baseline after later migration: %v", err)
				}
			}
		})
	}
}

// Include sequence counters to catch accidental replay of seed INSERTs.
func snapshot(t *testing.T, db *sql.DB, schema string) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT relname FROM pg_class WHERE relnamespace = $1::regnamespace
		AND relkind IN ('r', 'S') AND relname NOT LIKE 'goose_db_version%'`, schema)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	result := make(map[string]string)
	for _, name := range names {
		var data string
		query := `SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text), '[]'::jsonb)::text FROM (SELECT * FROM ` + pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier(name) + " OFFSET 0) t"
		if err := db.QueryRow(query).Scan(&data); err != nil {
			t.Fatal(err)
		}
		result[name] = data
	}
	return result
}
