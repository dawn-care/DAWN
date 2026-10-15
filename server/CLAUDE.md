# Database migrations

The API and migration command share `internal/db.Open` and its `DB_*`
environment variables. Neither loads `.env` automatically.

Locally, run `go run ./cmd/migrate` from `server/`.
In Docker, run `docker compose exec server ./migrate` from the repo root
against a running server container. Rebuild the image and recreate the server
container after adding migrations; SQL files are embedded in the executable.

Add the next numbered SQL file to `internal/migrations/sql/`, starting with
`-- +goose Up`. Wrap procedure bodies in Goose `StatementBegin` / `StatementEnd`
annotations. Do not edit applied migrations. Keep the initial demo data.

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
