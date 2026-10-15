# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

StartUp is a dementia-assistance product for SUDA (SUDA-Start-Up-Dementia-Assistance). A caregiver creates content (media, prompts, widgets) and schedules it into a daily agenda; a carepartner (the person with dementia) views that agenda on a paired device. The repo is a monorepo with three independently-run pieces and a folder of code they share:

- `server/` — Go backend (chi router), the only piece talking to Postgres/MinIO/external APIs.
- `caregiver/` — React + TypeScript + Vite web app used by caregivers to manage content.
- `carepartner/` — Expo/React Native app that displays the agenda to the carepartner. The deployment target is a React Native app on an **Android tablet**; the web build is used for local development.
- `shared/` — plain TypeScript rules used by both frontends, so the caregiver's preview matches what the tablet shows. Not an app: it has no `package.json` and nothing runs it on its own. See "Shared code" below.

## Commands

### Root (from repo root, uses `concurrently`)

```bash
npm run dev:caregiver     # backend + caregiver web app together
npm run dev:carepartner   # backend + carepartner app together
npm run dev:backend       # just the Go server
npm run dev:web           # just the caregiver app
npm run dev:mobile        # just the carepartner app
```

### Backend (`server/`)

```bash
cd server
go run cmd/main.go                # run (defaults RUN_MODE=production; see Server startup below)
go build -o main ./cmd            # build
go test ./...                     # all tests
go test ./internal/services/modules/...   # one package
go test ./internal/services/modules/ -run TestCreateModule   # one test
swag init -g ./cmd/main.go -o ./cmd/docs  # regenerate Swagger docs after changing handler annotations
```

Backend tests use `github.com/DATA-DOG/go-sqlmock` (no real DB needed) — see `server/internal/utils/test_helpers.go` for `SetupMockDB`/`SetupMockMinIO` helpers used across `*_test.go` files. Every layer (model, service, controller) typically has its own test file next to it.

### Caregiver web app (`caregiver/`)

```bash
cd caregiver
npm run dev       # vite dev server (localhost:5173)
npm run build     # tsc -b && vite build
npm run lint      # eslint .
```

### Carepartner app (`carepartner/`)

```bash
cd carepartner
npm start                 # expo start
npm run web                # expo start --web
npm run android / ios      # native builds
npm run lint                # expo lint
npm test                    # unit and component tests (Jest + React Native Testing Library)
npm run e2e                 # end-to-end flows (Maestro), against the app installed on an emulator or tablet
```

Carepartner tests live in a `__tests__/` folder inside each feature folder under `carepartner/src/`, never in `carepartner/app/` (files there become routes). End-to-end flows are in `carepartner/.maestro/`. See `carepartner/CLAUDE.md` for details.

### Shared code (`shared/`)

```bash
cd carepartner && npm test      # runs the tests in shared/ too (Jest roots include ../shared)
npm run typecheck:shared        # from the repo root: compiles shared/ with no DOM lib and no framework types (run npm install in carepartner/ first)
```

Tests sit in a `__tests__/` folder inside each feature folder of `shared/`, like the carepartner's. Only the carepartner runs them today; the caregiver is not wired up yet.

### Full stack via Docker

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

`docker-compose.yml` is the production-shaped compose (Postgres, MinIO, MinIO init, TLS via autocert, RIT-hosted domain); `docker-compose.dev.yml` overrides it for local dev (`RUN_MODE=dev`, plain HTTP on :8080, no TLS/cert volume). The `server/Dockerfile` multi-stage build compiles the **caregiver** web app and copies its `dist/` into the final Go image — the Go binary serves that built frontend as static files and serves the API from the same origin in production (see "Static frontend serving" below). The image build also copies `shared/` next to `caregiver/` (as `/frontend/shared`, so a `../shared` import from the caregiver resolves). The carepartner app is never bundled into this image; it ships separately as an Android app.

## Backend architecture (`server/internal/`)

Every domain (modules, configurations, exceptions, media/{photo,video,audio}, prompts/{message,event,caregiver_location,carepartner_location}, user/{caregiver,carepartner,device_pairing}, widgets/{news,weather}, auth) follows the same four-layer pattern, each in its own subpackage:

```
routes/<domain>/      → chi.Router registration, maps HTTP verb+path to a controller method
controllers/<domain>/ → decodes requests, calls a service through an interface, encodes JSON responses
services/<domain>/    → business logic; talks to *sql.DB (and minio.Client for media) directly with raw SQL
models/<domain>/      → plain structs representing DB rows
```

Controllers depend on a **service interface** defined in the same file as the controller (e.g. `ModuleServiceInterface` in `module_controller.go`), not a concrete type — this is what makes controllers unit-testable against mocks. When adding a new endpoint, follow this same route → controller → service → model chain and add tests at the service (sqlmock) and controller (via the interface) layers, matching the existing `*_test.go` siblings.

All wiring happens in `server/cmd/main.go`: services and controllers are constructed there (with real `*sql.DB` and `minio.Client`), then routes are registered onto a single `chi.NewRouter()`. There's no DI framework — just explicit constructor calls in main, grouped by domain (User / Auth / Agenda / Media / Widgets / Prompts). The API and `server/cmd/migrate` both use `internal/db.Open` for database settings and connection checks. See `server/CLAUDE.md` for migration commands and the current baseline requirement.

### Auth & pairing model

There are two distinct identity flows, not one:

- **Caregiver**: Google OAuth (`POST /api/auth/google` → `auth_service.LoginWithGoogle` verifies the ID token, finds/creates a `Caregiver` row, creates a session). Session cookie (`session`) is validated by `middleware.RequireAuth` (`server/internal/middleware/auth_middleware.go`), which puts the caregiver's user ID on the request context. Only a couple of routes (`/api/me`) are currently wrapped in this middleware — most routes are not yet auth-gated, so don't assume a route is protected just because a session exists elsewhere.
- **Carepartner device**: no login — it pairs to a caregiver's account via a short pairing code (`POST /api/users/device-pairing/code|verify|status`), then stores its own `carepartnerID` locally (`localStorage` on web, `expo-secure-store` on native — see `carepartner/src/pairing/carepartnerId.ts`). All subsequent carepartner API calls are scoped by that ID as a query param/body field, not a session.

### Data model

The schema lives in Goose migrations under `server/internal/migrations/sql/`; there is no `init/database.sql` (see `server/CLAUDE.md` for running them). The content model (ticket i332, migration 3): `Media` holds a carepartner's photos, videos and audio (kind, display name, MinIO object key, duration). `ScreenSettings` (one row per carepartner, created with the carepartner) holds the layout preset and awake hours. `ScreenWidget` places a widget in one `spot`, a 0-based index shared by every preset (widgets follow reading order), so a widget keeps its place when the preset changes. Presets are just a name and a spot count, defined in Go (`Presets` in `services/screen_settings`); the frontend and tablet decide where each spot sits on screen. Any widget type can go in any spot. A widget whose spot is past the current preset's count is kept but hidden. Each widget type's `config` JSON is validated in `services/screen_widgets/config.go`. `ScheduledItem` plays media at a set time (once, daily, weekly or monthly), with `ScheduledItemSkip` rows for single deleted days; repeats are expanded in Go (`services/scheduled_items/recurrence.go`), not stored. `GET /api/screen` assembles all of it for the tablet. `Playback` (migration 2, moved onto this model by migration 3) records each attempt to play a media item, with the scheduled item for scheduled plays. The frontend contract for these endpoints is `docs/api/screen-and-media.md`. The older `Module`/`Configuration`/`Exception`/per-type media and prompt Go code is still in the tree, but migration 3 dropped its tables; it is removed in a follow-up.

Media files themselves (photo/video/audio blobs) are stored in MinIO under bucket `media`; Postgres only holds the URI/filename/type metadata. `server/cmd/main.go` ensures the `media` bucket exists on startup.

### Static frontend serving

In production the Go binary itself serves the built caregiver SPA (`/app/frontend`, copied in by the Dockerfile from the Vite build) with a catch-all route (`r.Get("/*", ...)` → `index.html`) after all `/api/*` and `/assets/*` routes are registered — so route order in `main.go` matters; API/asset routes must be registered before the SPA catch-all.

### Server startup mode

`RUN_MODE` env var switches between `dev` (plain HTTP on `PORT`, default 8080) and anything else/unset (production: `autocert`-managed HTTPS on :443 restricted to `startup-da.gccis.rit.edu`, plus an HTTP→HTTPS-capable listener on :80). Don't assume TLS/autocert is active locally — `docker-compose.dev.yml` and direct `go run` default to dev-style plain HTTP.

## Frontend architecture

### Caregiver (`caregiver/`)

Standard Vite + React Router SPA (`src/App.tsx`). `AuthContext` (`src/context/AuthContext.tsx`) holds `isAuthenticated`/`user` plus the currently-selected `activeCarePartnerId` (persisted to `localStorage`, since a caregiver can manage multiple carepartners). Most routes are wrapped in `RequireCarePartner` (`src/components/RequireCarepartner.tsx`), which gates access on an active carepartner selection existing — not just on auth — so a logged-in caregiver with no carepartner selected still gets redirected. API base URL comes from `.env.development` (`VITE_API_URL`), matching the `API_URL`/`VITE_API_URL` build args threaded through `docker-compose*.yml` and the server Dockerfile.

### Carepartner (`carepartner/`)

Expo Router app (file-based routing under `carepartner/app/`) targeting native (Android tablet is the deployment target) and web (used for development). `app/` holds routes only; everything else lives in feature folders under `src/`. `src/api/config.ts` exports `API_URL`, read from `EXPO_PUBLIC_API_URL` and defaulting to the local dev backend. `app/pair.tsx` handles the device-pairing UI; `src/pairing/carepartnerId.ts` is the cross-platform (web localStorage vs. `expo-secure-store`) storage shim for the paired ID, used throughout instead of talking to `AuthContext`-style state. An earlier Electron/Raspberry Pi kiosk build was dropped in favor of the Android tablet target — don't reintroduce Electron-specific code.

### Shared code (`shared/`)

`shared/` holds framework-free rules that the caregiver preview and the tablet must agree on. It currently has night hours (`display/nightHours.ts`), time of day and date labels (`widgets/timeOfDay.ts`), and the always-on grid's size and `fitsGrid` (`widgets/grid.ts`).

- Plain TypeScript only: no React, React Native, Expo or DOM imports, and no dependencies. Rules take the time and data as arguments and never read the clock themselves. `npm run typecheck:shared` fails on a DOM API or a React import with types.
- Import from the barrel (`import { isNight } from "@shared"`), never a deep path, and re-export new modules from `shared/index.ts`.
- Keep code in `shared/` only if both frontends use the same rule. Anything that depends on the React Native theme or screen size stays in the app (for example `cellRect` and the grid's padding and gap).
- The carepartner reaches it through the `@shared` path in `tsconfig.json`, `watchFolders` in `metro.config.js` and the Jest `roots` and `moduleNameMapper` in `package.json`. The caregiver has no alias yet; wiring it up is deferred to a follow-up ticket (see #354).
- Write shared code to compile under the strictest settings of both apps (`import type`, no file extensions in imports, no unused locals).
