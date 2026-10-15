# Database migrations

The API and migration command share `internal/db.Open` and its `DB_*`
environment variables. Neither loads `.env` automatically.

The API runs `migrations.Up` after connecting to the database and before serving
requests, locally and in Docker. Only pending migrations run; a failure stops
startup. Fresh databases receive the initial schema and demo data through Goose.
Rebuild the executable or Docker image after adding migrations because SQL files
are embedded. Review destructive SQL before deploying: startup applies it too.

Manual commands remain available: `go run ./cmd/migrate` from `server/`, or
`docker compose run --rm --no-deps server ./migrate` from the repo root with the
database running and the server image built.

Add the next numbered SQL file to `internal/migrations/sql/`, starting with
`-- +goose Up`. Wrap procedure bodies in Goose `StatementBegin` / `StatementEnd`
annotations. Do not edit applied migrations. Keep the initial demo data.

For databases created by `init/database.sql`, first run the same command with
`baseline` appended. This validates migration 1 using temporary reference tables
and records it without replaying seeds. Then restart the API to apply pending
migrations. The Docker command above works even when the API cannot start because
baselining is required. Do not bypass failed schema checks or delete volumes.
A new, empty database gets its whole schema and demo data from the migrations;
Docker Compose no longer loads an `init/` script. Run the migrate command once
after `docker compose ... up` on a fresh volume.

Databases created by the old `init/database.sql` (before migrations existed)
need one extra step: first run the same command with `baseline` appended. This
validates migration 1 using temporary reference tables and records it without
replaying seeds. Then run the normal command for the later migrations. Do not
bypass failed schema checks or delete volumes.

Safety tests: `go test -race ./internal/db ./internal/migrations`.

Integration tests also run when `MIGRATION_TEST_DSN` is set to a disposable
PostgreSQL database (lib/pq keyword connection string). They create and remove
isolated test schemas. Never point this setting at an application database.
