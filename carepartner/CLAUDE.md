# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

The carepartner-facing app for StartUp (SUDA-Start-Up-Dementia-Assistance): an Expo Router / React Native app that pairs to a caregiver's account and plays back the carepartner's daily agenda. In practice it is primarily run as a **web build**, either in a browser or packaged in Electron as a Raspberry Pi kiosk `.deb`. See the repo-root `CLAUDE.md` for the full monorepo picture (backend, caregiver app, data model). This file covers only what's specific to `carepartner/`.

## Commands

```bash
npm start                    # expo start
npm run web                  # expo start --web (localhost:8081), needs the Go backend on :8080
npm run lint                 # expo lint
npm run electron:dev         # expo web + electron window pointed at http://localhost:8081
npm run electron:export      # expo export --platform web → dist/
npm run electron:build:rpi   # export + electron-builder arm64 .deb → electron-dist/
npm run electron:build:rpi32 # same, armv7l
```

This package has no test suite. `npm run reset-project` is left over from the Expo template, and the `scripts/` directory it points to doesn't exist.

The production `.deb` is built via Docker from the repo root: `docker compose --profile build-rpi up --build carepartner-build`. This uses `Dockerfile.build` and copies the `.deb` into `carepartner/electron-dist/`.

## Architecture

### API URL is a hardcoded file, overwritten at build time

`app/config.ts` exports a single `API_URL` constant (`http://localhost:8080/api`). There is no env-var mechanism. `Dockerfile.build` **overwrites `app/config.ts`** with the `API_URL` build arg before exporting. If you add config values to that file, they'll be dropped in the Docker build unless you also update the `echo` line in `Dockerfile.build`.

Every screen imports `API_URL` and calls `fetch` inline. There is no shared API client.

### Everything lives in `app/`, so every file is a route

Expo Router treats each file in `app/` as a route, including the non-screen modules `config.ts`, `carepartnerId.ts` and `ScheduleContext.tsx`. The media screens (`photo.tsx`, `audio.tsx`, `greeting.tsx`, `text.tsx`, `widget.tsx`) are also routes, but they are really child components: `agendaView.tsx` imports and renders them directly with a `configuration` prop. They aren't navigated to.

### Flow and global scheduling (`app/ScheduleContext.tsx`)

`_layout.tsx` wraps the whole `Stack` in `ScheduleProvider`. This is the app's central state and scheduler:
- It fetches `/users/carepartner/:id` (for `auto_start`) and `/configurations/day?carepartner_id&day&month&day_of_week&date` (today's resolved agenda). It refetches on mount and every 5 minutes.
- It navigates to `/agendaView` on its own, from two 15s-interval checks:
  - `auto_start` time → `trigger=auto_start`, but only if there are unplayed non-timed items.
  - Each `timed` item's `time` → `trigger=scheduled`.
- It tracks `playedIds` (reset at midnight) so items don't replay.

`/` renders `screensaver.tsx`, the idle home screen. From there the carepartner can manually enter the agenda (`trigger=manual`), sync, or go to `/pair`.

`agendaView.tsx` builds its play queue **once** per mount:
- `scheduled` trigger → timed items due now or earlier that haven't played, sorted by time.
- Any other trigger → all non-timed items.

The queue is deliberately not rebuilt on later refreshes, to avoid replay loops. Audio and video advance through the `onPlaybackComplete` callback. Everything else advances after `duration` seconds (default 10). When the queue ends or errors, the app returns to `/`.

### Types mirror the backend by hand

`ConfigurationWithMetadata` and the per-type metadata types are defined in `agendaView.tsx`. They mirror the Go `ConfigurationWithMetadata` struct returned by `GetConfigurationsByDayAndCarepartner`, and other files import them from there. Casing is inconsistent: weather and news metadata use upper-case keys (`ID`, `ZIPCODE`). Match whatever the backend actually serializes. The `module_type` → screen mapping lives in `AgendaView`'s render; `messages` renders both the photo and text viewers.

### Web-only media playback

`audio.tsx` and `greeting.tsx` (video) render raw HTML `<audio>`/`<video>` elements, not React Native components, so playback only works on web/Electron. Media is streamed from the backend (`/photo|audio|video/file?id=`). Relative `image_uri` values are prefixed with `API_URL`.

### Pairing and identity (`app/pair.tsx`, `app/carepartnerId.ts`)

There is no login. On the pair screen, the device:
1. Validates its stored ID, or creates a new carepartner (`POST /users/carepartner/create-carepartner` with a placeholder name).
2. Requests a pairing code.
3. Polls `/users/device-pairing/status` when the user taps Verify.

The ID is stored under the key `carepartnerID`, via `carepartnerId.ts` (`localStorage` on web, `expo-secure-store` on native). `terms.tsx` is still a stub.

### Electron wrapper (`electron.js`)

When packaged, it serves `dist/` through a custom privileged `app://` protocol with SPA fallback to `index.html`, so Expo Router's History API routing works without an HTTP server. When unpackaged, it loads `http://localhost:8081`. New asset types may need an entry in its `getMime` table.
