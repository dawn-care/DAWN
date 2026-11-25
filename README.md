# DAWN

## Background
For a patient with Alzheimer's or similar type of dementia who feels unsettled by their surroundings, who wants to know where they are and not feel abandoned when they wake up, DAWN (Daily Awareness and Well-being Navigator) is a dementia assistance tool acting as a surrogate for a loved one, reducing the physiological anxiety of the person with dementia and improving the customer’s quality of life. Unlike face-to-face interactions in real-time, our project is more scalable and available, allowing customers to feel connected despite the caregiver's busy schedules.

DAWN enables caregivers to create, manage, and deliver personalized content for carepartners. The system supports media (photo, video, audio) and screen widgets (messages, reminders, weather, news, sports), all displayed through a simple and accessible interface.

---

## Key Features
- Upload and manage media (photos, videos, audio)
- Create and edit message widgets with optional photos and reminder widgets
- Add dynamic widgets (weather, news, sports)
- Assign content to specific carepartners
- Interactive library for browsing and selecting content
- Agenda integration for structured content delivery
- Secure authentication using Google OAuth

---

## System Architecture

### Frontend
- React
- TypeScript
- Tailwind CSS

### Backend
- Go (Golang)
- RESTful API using Go standard `net/http`
- Handles:
  - Authentication (Google OAuth)
  - Screen layouts and widget configuration
  - Media upload/storage metadata
  - Widget integrations
  - Media scheduling and skipped occurrences

### Storage Layer
- PostgreSQL: Stores users, media metadata, screen settings/widgets, scheduled items, and playback records
- MinIO: Stores raw uploaded files (media) before deployment migration to AWS S3

### External Integrations
- Google OAuth authentication
- Weather API
- News API
- Sports API

## Data Flow

1. The caregiver signs in and pairs the carepartner's device using a pairing code.
2. The caregiver uploads media through `/api/media`; files are stored in MinIO and metadata in PostgreSQL.
3. The caregiver saves photos, messages, reminders, and live widgets through `/api/screen-layout`.
4. The caregiver schedules media through `/api/scheduled-items`, including repeats and skipped occurrences.
5. The tablet reads `/api/screen` for its layout, live widget data, and scheduled content.
6. The tablet reports playback through the playback API and realtime connection.

---

## Data Model

- **Media:** Photos, videos, and audio owned by a carepartner, with MinIO object keys and file metadata.
- **ScreenSettings:** The carepartner's awake hours.
- **ScreenWidget:** Widgets placed on the screen grid, with configuration for photos, messages, reminders, weather, news, or sports.
- **ScheduledItem:** Media scheduled once or on a daily, weekly, or monthly repeat.
- **ScheduledItemSkip:** Individual occurrences removed from a repeating schedule.
- **Playback:** Records of media playback attempts and their status.

The old `Module`, `Configuration`, `Exception`, per-type media/widget, and prompt APIs have been removed. Migration 3 removes their tables, the `insert_configurations` and `delete_configurations` procedures, and `auto_start`. Historical migrations and baseline checks remain to support upgrades from older databases.

See [the screen and media API guide](docs/api/screen-and-media.md) for endpoint contracts and examples.

---

## Setup Instructions

### Prerequisites
- Docker (with Docker Compose)
- Node.js and npm
- Go 1.25+ (only for running the backend outside Docker)
- PostgreSQL 15 (only for running the backend outside Docker)

---

### 1. Create your `.env`

Docker Compose reads a `.env` file in the repo root. It is gitignored, so each developer creates their own:

```bash
cp .env.example .env
```

Then fill in:

| Variable | What it is |
|---|---|
| `OPENWEATHER_API_KEY` | Weather widget key from https://openweathermap.org/api |
| `NEWS_API_KEY` | News widget key from https://newsapi.org |
| `GOOGLE_CLIENT_ID` | Google sign-in client ID. Already filled in by `.env.example`; leave it as is. |
| `DB_PASSWORD`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` | The database and MinIO passwords. `.env.example` has development values; leave them as is locally. |

Compose refuses to start if `GOOGLE_CLIENT_ID`, `DB_PASSWORD`, `MINIO_ROOT_USER` or `MINIO_ROOT_PASSWORD` is missing. If your `.env` is older than these, copy the three lines from `.env.example`.

---

### 2. Run in development mode

There are two ways to run the app locally. Use Option A to just get the app running, and Option B when you are actively changing code and want hot reload.

#### Option A: everything in Docker

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Both `-f` flags are needed: `docker-compose.yml` on its own is the production setup (HTTPS on the RIT domain).

This starts Postgres, MinIO, and the Go server, which also serves the built caregiver web app:

| What | URL |
|---|---|
| Caregiver web app + API | http://localhost:8080 |
| API docs (Swagger) | http://localhost:8080/swagger/ |
| MinIO console (`MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` from `.env`) | http://localhost:9001 |

Code changes need a rebuild (`up --build` again). On startup, the API applies pending Goose migrations and seeds a fresh database. Existing databases created by `init/database.sql` need the one-time baseline step below before the API can start.

#### Option B: backend and frontend on your machine (hot reload)

The backend needs Postgres and MinIO to be reachable on `localhost`.

**MinIO** — start just that service from Compose (the dev file is what publishes MinIO on `localhost:9000`):

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d minio minio-init
```

**Postgres** — the Compose database is not exposed to your machine, so use a local Postgres on port 5432. One-time setup, run as a Postgres superuser:

```bash
psql -c "CREATE USER myuser WITH PASSWORD 'mypassword';"
psql -c "CREATE DATABASE mydb OWNER myuser;"
```

The API creates the schema and demo data through Goose on its first startup.

**Backend** — the Go server does not read `.env` itself, so load it into your shell first. `RUN_MODE=dev` is required; without it the server starts in production mode and tries to bind ports 80/443.

```bash
set -a; source .env; set +a
cd server
RUN_MODE=dev go run cmd/main.go
```

The API is now at http://localhost:8080/api.

**Caregiver web app** — in a second terminal:

```bash
cd caregiver
npm install
npm run dev
```

Open http://localhost:5173. Its settings (`VITE_API_URL`, `VITE_GOOGLE_CLIENT_ID`) come from `caregiver/.env.development`, which is committed.

#### Database migrations

The API automatically applies pending Goose migrations after connecting to the
database and before accepting requests, locally and in Docker. An up-to-date
database is unchanged; a migration failure stops startup. Fresh databases get
their schema and demo data from Goose rather than Docker's init script.

The manual command remains available: `go run ./cmd/migrate` from `server/`,
using the same `DB_*` environment variables as the API. In Docker, with the
database running and the server image built:

```bash
docker compose run --rm --no-deps server ./migrate
```

SQL files are embedded in the executable: rebuild it or the Docker image after
adding migrations, then restart/recreate the server. Startup also applies
destructive SQL, so review pending migrations before deployment.
Add the next numbered file under `server/internal/migrations/sql/`
(for example `00006_add_album_covers.sql`) with `-- +goose Up` and the forward
SQL. Use Goose's `StatementBegin` / `StatementEnd` around procedure bodies.
Do not edit already-applied migrations. The initial migration retains demo data;
no rollback command is provided.

**Existing databases:** a database created by the old `init/database.sql`, before
migrations existed, must be registered once, without replaying the schema or demo
inserts:

```bash
# Local (from server/)
go run ./cmd/migrate baseline
# Docker (from the repo root, with the database running and server image built)
docker compose run --rm --no-deps server ./migrate baseline
```

Then restart the API to apply pending migrations. Baselining is never automatic;
the one-off Docker command works even when the API cannot start yet.
Baselining checks the original tables' columns, defaults, identity modes,
constraints and stored procedures. It uses temporary reference tables that are
removed automatically, and records version 1 only if validation succeeds.
Application rows and sequence counters are untouched. Repeating a successful
baseline is safe. A failed check may leave an empty Goose tracking table, but
never marks migration 1 applied. Resolve reported schema differences before
retrying; do not delete your database volume.

#### Carepartner app

The carepartner app is never part of the Docker stack. With the backend running (either option), start it separately:

```bash
cd carepartner
npm install
npm start        # or: npm run web
```

The web version opens at http://localhost:8081 and talks to the backend at `http://localhost:8080/api` (set in `carepartner/app/config.ts`).

---

### Changing the database or MinIO password

`DB_PASSWORD`, `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` come from `.env`. In production, use
long random values (`openssl rand -hex 24` makes one) and keep them only in the server's `.env`.
Compose won't start without them, so add them to the server's `.env` before deploying a version
that needs them. Setting them to the old development values (from `.env.example`) keeps
everything working as it was; then change them as below.

On a server that already has data:

1. **Database.** Postgres only reads `DB_PASSWORD` when the database is first created, so change
   the existing user's password to match (no password is needed from inside the container; avoid
   `'` in the new password):

   ```bash
   docker compose exec db psql -U myuser -d mydb -c "ALTER USER myuser WITH PASSWORD 'the-new-password';"
   ```

2. **`.env`.** Set `DB_PASSWORD` to that same password, and set new `MINIO_ROOT_USER` and
   `MINIO_ROOT_PASSWORD` values. MinIO reads its credentials on every start, and uploaded media
   stays where it is.
3. **Restart** with the new values: `docker compose up -d --build`.
4. **Check** that you can sign in to the caregiver site and that photos and videos still load.

To undo, put the old values back in `.env` (and run step 1 again with the old database password).

### Google sign-in

The client ID appears in two places and both must hold the same value:

- `GOOGLE_CLIENT_ID` in the root `.env` — used by the backend, and by the frontend when built in Docker
- `VITE_GOOGLE_CLIENT_ID` in `caregiver/.env.development` — used by `npm run dev`

If the Google button fails with "The given origin is not allowed for the given client ID", the URL you are on is not listed under **Authorized JavaScript origins** for the OAuth client in Google Cloud. `http://localhost:5173` and `http://localhost:8080` are the two used in development.

---

### Tests

```bash
cd server && go test ./...          # backend
cd carepartner && npm test          # carepartner app and shared/
npm run typecheck:shared            # from the repo root
```

Backend tests need nothing running. A few extra tests run only when you provide what they need, and skip otherwise: `MIGRATION_TEST_DSN` (a disposable PostgreSQL database), `NEWS_API_KEY`, `OPENWEATHER_API_KEY`, and `ffprobe` installed. The caregiver app has no tests; check it with `npm run lint` and `npm run build`.

CI runs all of these on every pull request (see `.github/workflows/ci.yml`).

---

## Branching and pull requests

```
your branch ──PR──▶ canary ──PR──▶ main
```

- **`canary`** is the default branch, where all work is merged. Start new branches from it.
- **`main`** holds released code, and only `canary` is merged into it. Once deploys are set up (#219), a merge into `main` will deploy.
- Neither branch accepts direct pushes or force pushes, and neither can be deleted. Every change goes through a pull request.

### Into `canary`

- Open a pull request from your branch into `canary`.
- It needs **1 approval** from someone other than you. Pushing new commits dismisses earlier approvals, and the latest push must be approved by someone other than the person who pushed it.
- CI must pass: `changes`, `server`, `caregiver` and `carepartner`. Jobs for parts of the repo your pull request doesn't touch are skipped, and a skipped job counts as passing.
- Your branch must be **up to date** with `canary`. If `canary` has moved on since you branched, use **Update branch** on the pull request (or merge or rebase `canary` yourself) and let CI run again. This way the exact result of the merge has been tested.
- Any merge method is allowed.

### Releasing: `canary` into `main`

- Open a pull request from `canary` into `main`.
- It needs **2 approvals**. New commits on `canary` while it is open dismiss them, so the approvals always cover what will be released.
- CI must pass, and so must `source-branch`, which fails a pull request into `main` from any branch other than `canary`.
- Merge with a **merge commit** (squash and rebase are turned off). Squashing would give `main` a commit `canary` never has, and later release pull requests would show old changes again.
- The pull request doesn't need to be up to date with `main`. Each release adds a merge commit to `main` that `canary` doesn't have, and that is expected.

### Hotfixes

There is no shortcut into `main`. A fix goes into `canary` like any other change, then out in a release.

---

## License

DAWN is licensed under the [Apache License, Version 2.0](LICENSE). Copyright 2026 The DAWN Authors; see [NOTICE](NOTICE).

Unless you state otherwise, any contribution you submit to this repository is licensed under the same terms (section 5 of the license).
