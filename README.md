# StartUp

## Background
For a patient with Alzheimer's or similar type of dementia who feels unsettled by their surroundings, who wants to know where they are and not feel abandoned when they wake up, StartUp is a dementia assistance tool acting as a surrogate for a loved one, reducing the physiological anxiety of the person with dementia and improving the customer’s quality of life. Unlike face-to-face interactions in real-time, our project is more scalable and available, allowing customers to feel connected despite the caregiver's busy schedules.

StartUp enables caregivers to create, manage, and deliver personalized content for carepartners. The system supports multiple module types including media (photo, video, audio), prompts (messages), and widgets (weather, news, sports), all displayed through a simple and accessible interface.

---

## Key Features
- Upload and manage media (photos, videos, audio)
- Create and edit prompts (messages with optional images)
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
  - Module management
  - Media upload/storage metadata
  - Widget integrations
  - Configuration scheduling logic

### Storage Layer
- PostgreSQL: Stores all structured metadata (modules, users, configurations, prompts, widgets)
- MinIO: Stores raw uploaded files (media) before deployment migration to AWS S3

### External Integrations
- Google OAuth authentication
- Weather API
- News API
- Sports API

## Data Flow

1. Caregiver logs in and pairs with carepartner via Google OAuth
2. Caregiver uploads media → BLOB stored in MinIO
3. Metadata (URI, filename, type) created → stored in PostgreSQL
4. Caregiver creates widgets and/or prompts
5. Modules are created linking metadata to carepartners
6. Frontend fetches modules and resolves metadata via REST endpoints
7. Carepartner views scheduled content based on Configuration rules
8. External APIs enrich widget-based modules (weather/news/sports)

---

## Data Model

The system uses a modular architecture centered around the `Module` table.

### Module 
- id 
- module_type (photo, video, audio, weather, news, sports, messages) 
- caregiver_id 
- carepartner_id 
- metadata_id 

### Each module links to a metadata record: 

### Media (photo/video/audio) 
- id 
- uri 
- displayed_name 
- filename 
- extension 

### Prompts (locations, events)
- id 
- message 
- image_uri 

### Widgets (sports, news, weather)
- id
- zipcode

---

## Configuration System

The Configuration system controls **when and how modules are displayed** to carepartners.

### Purpose
- Defines ordering of modules in a daily schedule
- Supports timed or interval-based content delivery
- Enables flexible reordering without modifying modules themselves
- Supports exceptions (special cases like skipping or replacing content)

### How It Works

Each configuration row links:
- a carepartner
- a module
- a position in the schedule
- timing rules (optional)

### Key Features
- Ordered scheduling via `position`
- Time-based or interval-based display (`timed`, `time`, `duration`)
- Batch updates via PostgreSQL stored procedure
- Exception handling for overrides

### Stored Procedures

#### insert_configurations
- Inserts multiple configuration rows atomically
- Maintains ordering and shifts existing positions
- Ensures no gaps in schedule ordering

#### delete_configurations
- Deletes configurations by ID list
- Re-indexes remaining positions to maintain order integrity

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

Compose refuses to start if `GOOGLE_CLIENT_ID` is missing.

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
| MinIO console (`localuser` / `localpass123`) | http://localhost:9001 |

Code changes need a rebuild (`up --build` again). On startup, the API applies pending Goose migrations and seeds a fresh database. Existing databases created by `init/database.sql` need the one-time baseline step below before the API can start.

#### Option B: backend and frontend on your machine (hot reload)

The backend needs Postgres and MinIO to be reachable on `localhost`.

**MinIO** — start just that service from Compose:

```bash
docker compose up -d minio minio-init
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
(for example `00002_add_album_covers.sql`) with `-- +goose Up` and the forward
SQL. Use Goose's `StatementBegin` / `StatementEnd` around procedure bodies.
Do not edit already-applied migrations. The initial migration retains demo data;
no rollback command is provided.

**Existing databases:** register a database created by `init/database.sql` once,
without replaying the schema or demo inserts:

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

### Google sign-in

The client ID appears in two places and both must hold the same value:

- `GOOGLE_CLIENT_ID` in the root `.env` — used by the backend, and by the frontend when built in Docker
- `VITE_GOOGLE_CLIENT_ID` in `caregiver/.env.development` — used by `npm run dev`

If the Google button fails with "The given origin is not allowed for the given client ID", the URL you are on is not listed under **Authorized JavaScript origins** for the OAuth client in Google Cloud. `http://localhost:5173` and `http://localhost:8080` are the two used in development.

---

### Tests

```bash
cd server
go test ./...
```

Some backend tests need MinIO running on `localhost:9000` and real weather/news API keys.
