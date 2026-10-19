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

Env vars (`VITE_API_URL`, `VITE_GOOGLE_CLIENT_ID`) come from `.env.development` under `npm run dev` only. `vite build` does not read that file, so Docker builds get both as build args from the compose files; `VITE_GOOGLE_CLIENT_ID` is fed from `GOOGLE_CLIENT_ID` in the repo-root `.env`. `VITE_CAREPARTNER_URL` (the carepartner web app, `http://localhost:8081` in development; run `npm run web` in `carepartner/`) is where the home page and `/preview` show the carepartner's screen in an iframe, in that app's preview mode (`src/screen/previewUrl.ts`). Docker builds don't set it yet, so both pages say "Carepartner screen can't be shown here." there (`src/screen/ScreenUnavailable.tsx`). Restart `npm run dev` after changing any of them.

## Architecture

### Routing and auth gating (`src/App.tsx`)

Single `<Routes>` tree. Two independent gates apply to most routes:

- `AuthContext` (`src/context/AuthContext.tsx`) tracks `isAuthenticated`/`user` (Google OAuth session) and `activeCarePartnerId` — the carepartner currently being managed, persisted to `localStorage` so it survives reloads. A caregiver can manage multiple carepartners; nothing in the UI scopes data by carepartner except this one context value.
- `RequireCarePartner` (`src/components/RequireCarepartner.tsx`) wraps almost every content route (`/library`, `/agenda`, `/agenda/:date`, `/media*`, `/prompt*`, `/widget*`) and redirects to `/setup` if `activeCarePartnerId` is falsy. It does **not** check `isAuthenticated` — a route can be "carepartner-gated" without being auth-gated, so don't assume one implies the other. `/`, `/guide`, `/login`, `/settings`, `/setup` are intentionally ungated.

### API calls: per-component, except the media page

There is no app-wide `api.ts`/service module. Each older page/component that talks to the backend independently reads `const API_URL = import.meta.env.VITE_API_URL` at the top of the file and calls `fetch`/`axios` directly inline (see `src/pages/Login.tsx`, `src/pages/AccountSetup.tsx`, `src/components/ConfigurationForm.tsx`). The newer feature folders (so far `src/media/`) keep one `api.ts` for their feature, with typed responses, an `ApiError` carrying the server's `code`, and `friendlyError` for caregiver wording; follow that for a new feature built on the endpoints in `docs/api/screen-and-media.md`. Don't add one app-wide client, and don't move the older inline calls into these files.

### Module/Configuration types mirror the backend

The backend's `Module` (`module_type: photo|video|audio|weather|news|sports|messages` + a per-type metadata table) and `Configuration` (scheduling: `occurence_type`, `occurence_definition`, `position`, `timed`/`time`) concepts are re-declared as discriminated-union TypeScript interfaces per file rather than shared types — e.g. `ModuleItem`/`MediaItem`/`WeatherItem`/`NewsItem` in `src/pages/Library.tsx`, `ConfigurationWithMetadata` in `src/components/AgendaList.tsx`. When backend field names or module types change, expect to update these interfaces in multiple places, not one central model file.

Key pages by role:
- `src/pages/Library.tsx` — central content hub: list/search/filter all modules, launch create modals, delete/rename.
- `src/components/ConfigurationForm.tsx` — create/edit a `Configuration` (recurrence: daily/weekly/monthly/yearly/single, position, optional time) for a module, including single-day exceptions off `currentDate`.
- `src/pages/CalendarView.tsx` / `src/pages/AgendaView.tsx` + `src/components/AgendaList.tsx` — calendar navigation and the resolved per-day agenda (configurations + exceptions merged).
- `src/pages/MediaFormCreate.tsx` / `MediaFormUpdate.tsx` / `MediaFormDelete.tsx`, `src/pages/PromptForm.tsx`, `src/pages/WidgetForm.tsx` — per-module-type CRUD forms.
- `src/pages/AccountSetup.tsx` / `src/pages/SelectCarepartner.tsx` — device-pairing-code flow and carepartner selection that sets `activeCarePartnerId`.

These page components are intentionally large and self-contained (state, API calls, and form UI together) — that's the established pattern here, not something to refactor apart unless asked.

### Screen editor (`src/screen/`, shown by `src/pages/Home.tsx`)

The home page shows the carepartner's screen as the carepartner web app in an iframe (`CarepartnerScreen`). Edit screen swaps it for a draft of the saved layout held by `useScreenEditor`: the add bar (`AddBar`) picks a widget type, `openCells` lights where it fits, the pencil, bin or a click on a tile selects a widget (`ScreenGrid`), and the settings panel beside the screen (`WidgetPanel`) changes the selected widget. A widget needs its settings before it can be kept (`isComplete` in `widgetSettings.ts`), and a newly added one must be saved with the panel's Save before another can be added. Save changes runs `saveProblem`, sends the whole list with `saveWidgets`, then loads the layout again; the draft is thrown away by Cancel or a failed load, never by a failed save. Reminders aren't edited here (existing ones are kept and sent back unchanged), and the day and time is never added, edited or removed. Wording for toasts and hints lives in `Home.tsx` and `AddBar.tsx`.

### Shared rules (`../shared`)

Rules the tablet must agree with (awake hours, what is playing, where widgets can be placed on the grid, widget names, the weather word) live in the repo-root `shared/` folder, not here. Import them with `import { ... } from "@shared"`; the alias is in `vite.config.ts` and `tsconfig.app.json`, and the Docker build copies `shared/` beside this folder. Their tests run from `carepartner/` (`npm test`), since this package has no test runner. See `shared/README.md`.

### Styling

Tailwind CSS (`tailwind.config.js` scans `index.html` and `src/**/*.{js,ts,jsx,tsx}`). The palette (`dawn-*` colors, including `dawn-accent`, `dawn-success`, `dawn-control`), the Atkinson Hyperlegible font (`font-dawn`, also set on `body`) and the popover/dialog/toast shadows are tokens in that config. Prefer them over raw hex values. If a new token doesn't show up in `npm run dev`, restart the dev server.

Shared UI pieces live in `src/components/ui/` (app-specific components such as `NavBar` stay directly in `src/components/`). They are Tailwind-styled React components, with [Radix UI](https://www.radix-ui.com/) (`@radix-ui/react-popover`, `-radio-group`, `-dialog`) supplying focus, keyboard and ARIA behavior and `lucide-react` supplying icons. Gluestack is not used here: it only installs into Expo/React Native projects, not this Vite app.

- `Badge`, `Dialog`, `IconButton`, `InfoPopover`, `PhotoPicker`, `SaveIndicator`, `Toast`
- `form/`: `RadioGroup`/`RadioOption`, `SegmentedControl`, `SelectField` (native `<select>`), `TextField`, `ToggleButton`
- `styles.ts` holds the shared focus ring (`FOCUS_RING`) and field classes; use `FOCUS_RING` on any new interactive control.

Toasts: `ToastProvider` is mounted once in `main.tsx`, so any page can call `useToast()`. Each layout draws the toasts with a `ToastViewport` (top center of the page content): `AppShell` for pages with the nav bar, `ToastLayout` for the rest (login, settings, setup, preview). A new layout route needs one of the two. `show(text, { actionLabel, onAction, onDismiss })` returns an id; `dismiss(id)` closes that toast wherever it is in the stack (no id closes the top one). `onDismiss` runs when a toast closes without its action being used (timeout, its X, or `dismiss`), which lets a caller hold something back until the chance to Undo has passed; the media page uses it to delete a removed file only then. Toasts stack: a new one waits behind the one on top and its 5-second timer starts only when it reaches the top. The timer pauses while the pointer is over the toast or focus is inside it. Closing a toast with its X or action button returns focus to what had it when the toast was shown. Show a toast after a `Dialog` closes, not while it is open. Use `aria-pressed` for on/off toggles and a Radix radio group (`RadioGroup`, `SegmentedControl`, `PhotoPicker`) when exactly one option must be chosen.

### Docker

`Dockerfile` in this directory builds a standalone nginx-served static image of just this app. It is **not** what production uses — per the root `CLAUDE.md`, `server/Dockerfile` compiles this app and serves it from the Go binary itself alongside the API. Check which Dockerfile an infra change actually affects before editing.
