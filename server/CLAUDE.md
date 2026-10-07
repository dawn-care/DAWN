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

Existing databases created by `init/database.sql` still require baselining;
that step is not implemented yet. Do not bypass the guard or delete volumes.

Safety tests: `go test -race ./internal/db ./internal/migrations`.
