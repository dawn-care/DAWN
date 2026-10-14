# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

The caregiver-facing web app for StartUp (SUDA-Start-Up-Dementia-Assistance). React + TypeScript + Vite SPA where a caregiver manages media/prompts/widgets and schedules them into a carepartner's daily agenda. See the repo-root `CLAUDE.md` for the full monorepo picture (backend, carepartner app, Docker/data model); this file covers only what's specific to working in `caregiver/`.

## Commands

```bash
npm run dev       # vite dev server (localhost:5173), needs the Go backend running separately
npm run build     # tsc -b && vite build
npm run lint      # eslint .
npm run preview   # preview a production build locally
```

There is no test suite in this package (no `test`/`vitest` script, no `*.test.tsx` files) — the backend (`server/`) is where automated tests live.

Env vars (`VITE_API_URL`, `VITE_GOOGLE_CLIENT_ID`) come from `.env.development` under `npm run dev` only. `vite build` does not read that file, so Docker builds get both as build args from the compose files; `VITE_GOOGLE_CLIENT_ID` is fed from `GOOGLE_CLIENT_ID` in the repo-root `.env`.

## Architecture

### Routing and auth gating (`src/App.tsx`)

Single `<Routes>` tree. Two independent gates apply to most routes:

- `AuthContext` (`src/context/AuthContext.tsx`) tracks `isAuthenticated`/`user` (Google OAuth session) and `activeCarePartnerId` — the carepartner currently being managed, persisted to `localStorage` so it survives reloads. A caregiver can manage multiple carepartners; nothing in the UI scopes data by carepartner except this one context value.
- `RequireCarePartner` (`src/components/RequireCarepartner.tsx`) wraps almost every content route (`/library`, `/agenda`, `/agenda/:date`, `/media*`, `/prompt*`, `/widget*`) and redirects to `/setup` if `activeCarePartnerId` is falsy. It does **not** check `isAuthenticated` — a route can be "carepartner-gated" without being auth-gated, so don't assume one implies the other. `/`, `/guide`, `/login`, `/settings`, `/setup` are intentionally ungated.

### No API client layer — calls are per-component

There is no shared `api.ts`/service module. Each page/component that talks to the backend independently reads `const API_URL = import.meta.env.VITE_API_URL` at the top of the file and calls `fetch`/`axios` directly inline (see `src/pages/Login.tsx`, `src/pages/AccountSetup.tsx`, `src/components/ConfigurationForm.tsx`). When adding a new API call, follow this existing per-file convention rather than introducing a new shared client — that would be inconsistent with everything else in the codebase.

### Module/Configuration types mirror the backend

The backend's `Module` (`module_type: photo|video|audio|weather|news|sports|messages` + a per-type metadata table) and `Configuration` (scheduling: `occurence_type`, `occurence_definition`, `position`, `timed`/`time`) concepts are re-declared as discriminated-union TypeScript interfaces per file rather than shared types — e.g. `ModuleItem`/`MediaItem`/`WeatherItem`/`NewsItem` in `src/pages/Library.tsx`, `ConfigurationWithMetadata` in `src/components/AgendaList.tsx`. When backend field names or module types change, expect to update these interfaces in multiple places, not one central model file.

Key pages by role:
- `src/pages/Library.tsx` — central content hub: list/search/filter all modules, launch create modals, delete/rename.
- `src/components/ConfigurationForm.tsx` — create/edit a `Configuration` (recurrence: daily/weekly/monthly/yearly/single, position, optional time) for a module, including single-day exceptions off `currentDate`.
- `src/pages/CalendarView.tsx` / `src/pages/AgendaView.tsx` + `src/components/AgendaList.tsx` — calendar navigation and the resolved per-day agenda (configurations + exceptions merged).
- `src/pages/MediaFormCreate.tsx` / `MediaFormUpdate.tsx` / `MediaFormDelete.tsx`, `src/pages/PromptForm.tsx`, `src/pages/WidgetForm.tsx` — per-module-type CRUD forms.
- `src/pages/AccountSetup.tsx` / `src/pages/SelectCarepartner.tsx` — device-pairing-code flow and carepartner selection that sets `activeCarePartnerId`.

These page components are intentionally large and self-contained (state, API calls, and form UI together) — that's the established pattern here, not something to refactor apart unless asked.

### Shared rules (`../shared`)

Rules the tablet must agree with (awake hours, what is playing, layouts and their spots, widget names, the weather word) live in the repo-root `shared/` folder, not here. Import them with `import { ... } from "@shared"`; the alias is in `vite.config.ts` and `tsconfig.app.json`, and the Docker build copies `shared/` beside this folder. Their tests run from `carepartner/` (`npm test`), since this package has no test runner. See `shared/README.md`.

### Styling

Tailwind CSS (`tailwind.config.js` scans `index.html` and `src/**/*.{js,ts,jsx,tsx}`), no component library on top of it.

### Docker

`Dockerfile` in this directory builds a standalone nginx-served static image of just this app. It is **not** what production uses — per the root `CLAUDE.md`, `server/Dockerfile` compiles this app and serves it from the Go binary itself alongside the API. Check which Dockerfile an infra change actually affects before editing.
