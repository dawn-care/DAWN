# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

The carepartner-facing app for StartUp (SUDA-Start-Up-Dementia-Assistance): an Expo Router / React Native app that pairs to a caregiver's account and plays back the carepartner's daily agenda. It is deployed as a **React Native app on an Android tablet**; the web build (`npm run web`) is mainly a development convenience. (An earlier Electron/Raspberry Pi kiosk build was removed — don't reintroduce it.) See the repo-root `CLAUDE.md` for the full monorepo picture (backend, caregiver app, data model). This file covers only what's specific to `carepartner/`.

This app is being **replaced outright**. "What is being built" describes the target, "Redesign" records the agreed decisions and governs new work, and "Current architecture" describes the code that still exists until each piece is replaced; update it as pieces land.

## What is being built

D.A.W.N. is becoming three apps on one backend: this tablet in the resident's room, a family app that supplies content, and a staff phone app that decides when content plays. In the code, the resident is the `carepartner` and family is the `caregiver`; keep those names.

The tablet's job is to show a calm, always-on screen and to play media when it is scheduled or when staff start it. The resident never has to touch it.

**Always-on screen**
- The day of the week and the time of day (morning, afternoon, evening) in very large type, at all times.
- A large photo of a family member.
- What's next (lunch, bingo, a visit) as a picture plus a clock face.
- Small extras the family chooses, such as the weather as one icon and one word, or a sports result. Staff can switch any of them off for a resident.
- Later: today's nurse's photo and first name, and an invitation shortly before an activity.
- At night the screen turns dim and warm, with one reassuring line.

**Playback**
- Family schedule media to play at set times, such as a morning message at 8 AM or a favourite song. It plays full screen with no tap needed.
- Staff start a clip from their phone and it begins on the tablet within 3 seconds; they can stop it from the phone too.
- When playback ends, the screen returns to the always-on view by itself.
- Later: captions and the sender's photo and name on every video, and tapping the family photo to replay the latest message as an optional shortcut.

**Device**
- The tablet pairs by a code shown on its screen and will eventually run locked in kiosk mode.
- It reports what was shown and every play (start, finish, stop, and whether it was scheduled, staff-started or tapped).

**Design rules** (these apply to every tablet screen)
- Nothing requires a tap. Tapping is at most an optional shortcut.
- Nothing moves or flashes on its own apart from playing content.
- Orientation text (day, time of day) is 76px or larger; all other text is 26px or larger; text contrast is 7:1 or better; touch targets are about 20 mm or larger.
- Each screen uses a handful of words at most.
- No medical information in anything the tablet reports, and structured fields only, never free text.

## Commands

```bash
npm start                    # expo start
npm run web                  # expo start --web (localhost:8081), needs the Go backend on :8080
npm run android              # expo run:android (build + install on a connected tablet/emulator)
npm run lint                 # expo lint
npm test                     # jest (jest-expo preset); `npx jest src/pairing` for one folder, `-t "name"` for one test
npm run e2e                  # maestro test .maestro, against the app installed on an emulator or tablet
```

- The Android build needs JDK 17 or 21 (set `JAVA_HOME`); the Gradle version in use does not run on JDK 25.
- For Maestro, build with the emulator closed (`npm run build:emulator`, x86_64 only), then start the emulator and run `npm run e2e:install` and `npm run e2e`. Building and emulating at the same time needs a lot of memory. `android/` is generated and gitignored.
- Framework-free rules that the caregiver app also needs live in the repo-root `shared/` folder and are imported as `@shared` (see "Shared code" below). `npm test` runs their tests too, because Jest's `roots` include `../shared`.
- Test files live in a `__tests__/` folder inside each feature folder (`src/pairing/x.ts` → `src/pairing/__tests__/x.test.ts`), never in `app/` (where they would become routes). A test of a route file goes in the matching feature's `__tests__/`. React Native Testing Library is v14, where `render` is async (`await render(...)`).
- `npm run lint` reports existing warnings in the old `app/screensaver.tsx`; it is due to be replaced (#316), so don't fix them in passing.
- `npm run reset-project` is left over from the Expo template, and the `scripts/` directory it points to doesn't exist.

## Redesign

### Assumptions

- The tablet shares its backend with the caregiver app, so endpoints the caregiver app uses must keep working. Where a caregiver setting has no UI yet (widget and photo choices), the server supplies a default.
- Target device: an 11-inch Samsung Galaxy Tab A class tablet (Tab A9+ / A11+), 1920×1200, **landscape**, tested on an Android 15 emulator. The exact model isn't confirmed, so compute layout from the reported screen size and never hardcode dimensions. The visual design is drawn at 1280×800, the same 16:10 shape.
- Tablets are assumed to have live internet. Don't build dedicated offline behaviour; keep the last good data and send play reports when the connection returns.
- Floor staff carry phones, so the tablet has no staff controls. Play and stop commands reach it from the server.

### Playback

- `Module` is the playable unit. There are two media kinds, **video** and **audio**, plus a category label (family message, song, news, weather, and later ones). News and weather reports are videos, not generated cards. The tablet needs only those two players; the category doesn't affect how something plays.
- Caregiver scheduling stays on `Configuration`, **timed items only**. The untimed routine and `auto_start` go away.
- A scheduled item plays when it is due and hasn't played today, within a grace window (about 10 minutes), not on an exact-minute match.
- Every play is a `Playback` row on the server with a source (`scheduled`, `staff`, `tap`) and a status (`requested` → `acknowledged` → `started` → `finished` / `stopped` / `failed` / `expired`). The tablet reports each step, including plays it starts itself. Played state lives there, not in memory.
- Staff-triggered play and stop arrive over a WebSocket, and the tablet acknowledges them. Commands expire after about 30 seconds so a clip never starts late. On reconnect the tablet asks for any live command. The socket also carries a "refresh now" nudge.
- A staff play pre-empts a scheduled one; a scheduled item never interrupts a staff clip.
- The tablet caches **all** of the resident's playable content ahead of time, so a staff-started clip can begin within 3 seconds. The cache is synced from a server manifest, and the tablet streams as a fallback if a file isn't downloaded yet.
- If the tablet can't play something, it reports the `Playback` as `failed`.

### Always-on screen and widgets

- The tablet polls one "screen" endpoint about every 10–15 minutes. Widget changes don't need to be immediate.
- The server resolves every widget and returns layout plus display data together. Widgets do not fetch for themselves, so renderers are pure display.
- The screen is a coarse grid (about 4 columns × 3 rows). Everything on it is a widget, including day and time, the photo and "what's next". Each widget type declares the sizes it supports; day and time is locked.
- The caregiver picks widgets and the **server arranges them**. Positions are stored and sent explicitly (column, row, width, height) so free placement can be added later.
- The tablet has one renderer per *look*, not per widget: orientation (day and time), photo, next event, person, and a generic tile (icon, word, short label) shared by weather, sports and future data widgets. An unknown renderer is skipped.
- Updates apply in place with no animation. At night the grid is replaced by the night screen. A widget whose data has gone stale for hours hides instead of showing old information.
- The design rules above apply to every widget: no taps required, no motion, and the text-size floors.

### State machine, not routes

Only two routes exist: `index` and `pair`. The index route shows the always-on screen or the full-screen player depending on one playback state machine. Playback is started by the schedule or the socket, never by navigation, so there is no player route, no transition and no back stack. An unpaired tablet opens `pair` by itself.

### Target structure

```
carepartner/
  app/                    routes only, each a few lines
    _layout.tsx           providers (screen data, playback)
    index.tsx             always-on screen, or the player when something is playing
    pair.tsx
  src/
    api/                  server address, shared request helper
    theme/                palette, font, sizes
    screen/               polling hook, screen response types
    widgets/
      types.ts            shared widget interface
      Grid.tsx
      WidgetFrame.tsx
      registry.ts
      renderers/          Orientation.tsx, Photo.tsx, Tile.tsx, ...
    playback/             state machine, players, cache, socket
    schedule/             what is due, grace window
    pairing/              stored id, pairing requests
    clock/                controllable clock for tests
    <feature>/__tests__/  unit and component tests for that feature
  .maestro/               end-to-end flows
```

- Group by feature. There is no shared `src/components` folder; add one only when a second feature needs the same visual.
- Inside a feature, keep logic files (plain functions and hooks, no layout) separate from visual files (take data in and draw it; no fetching, clock reads or timers).
- Check the shape of server responses once where they enter the app, and drop malformed items there.
- Read the time through `src/clock` so tests can control it.
- Rules that the caregiver app must agree on, such as night hours, time of day and the grid's size, belong in `../shared`, not in a feature folder here. They take the time as an argument.
- `android/` is generated and never edited by hand. Native code, such as kiosk mode, goes in an Expo module with a config plugin.

### Shared code

The repo-root `shared/` folder holds plain TypeScript rules used by this app and the caregiver app: `NightHours`, `DEFAULT_NIGHT_HOURS` and `isNight` (`display/nightHours.ts`), `timeOfDay`, `TimeOfDay`, `dayName` and `dateLabel` (`widgets/timeOfDay.ts`), and `GRID` (column and row counts only), `GridArea` and `fitsGrid` (`widgets/grid.ts`). It also has rules this app does not use yet: awake hours (`display/awakeHours.ts`), what is playing at a given time (`playback/currentlyPlaying.ts`), layouts and their spots (`widgets/layoutPresets.ts`), which widgets a layout hides (`widgets/screenWidgets.ts`) and the weather word (`widgets/weather.ts`). They are for the tablet to adopt when it moves from the 4 x 3 grid and night hours to layout spots and awake hours. See `shared/README.md`.

- Import from `@shared` only (`import { isNight } from "@shared"`). The path is set in `tsconfig.json`, Metro watches the folder through `metro.config.js`, and Jest maps it through `roots` and `moduleNameMapper` in `package.json`. Metro's `nodeModulesPaths` and Jest's `modulePaths` point at this folder's `node_modules` so files in `shared/` can find Babel's helpers. Names are unchanged from when this code lived in `src/`.
- Nothing in `shared/` may import React, React Native, Expo or a DOM API. Anything that needs the theme, fonts or screen size stays here: `src/widgets/layout.ts` keeps `cellRect` and builds the tablet's `GRID` from the shared one by adding padding and gap. There are two objects called `GRID`: the shared one holds only `columns` and `rows`, and the tablet's adds `gap` and `padding`. `fitsGrid` takes any `{ columns, rows }`, so either works.
- Tests for shared code live in `shared/<feature>/__tests__/` and run with `npm test` in this folder. `npm run typecheck:shared` at the repo root checks the framework-free rule; it uses this folder's TypeScript, so run `npm install` here first.

### Styling

- Use a shared theme (palette, Atkinson Hyperlegible font, sizes, corner radii) so the tablet matches the other apps.
- The grid and tile are our own components: positions are data-driven and the look is fixed by the visual design.
- Gluestack (the free, MIT-licensed library) is the intended component library for the staff and family apps. Its current version installs only into Expo and React Native projects; confirm it renders on Expo web before adopting it here. It does not install into the Vite caregiver web app (its setup is built around Metro and React Native Web), so `caregiver/` uses Radix UI instead.

### Testing

- **Maestro** for end-to-end flows (pairing, always-on screen by time of day, a scheduled play and return, staff play and stop). A flow can call the server directly to stand in for the staff app.
- **Jest with React Native Testing Library** for logic and components. Tests live in each feature's `__tests__/` folder, and new code comes with tests.

## Current architecture

### What has moved to `src/` so far

- `src/api/config.ts` exports `API_URL`: `EXPO_PUBLIC_API_URL` if set (in `.env.local`; see `.env.example`), otherwise the local backend for the platform, `http://localhost:8080/api` on web and `http://10.0.2.2:8080/api` on Android (the emulator's address for the computer it runs on).
- `app.config.ts` extends `app.json` with build settings that depend on the environment. It allows plain `http://` traffic in Android builds only when the server address is `http://` or unset, so production builds against `https://` keep it blocked. Rebuild after changing `EXPO_PUBLIC_API_URL`.
- `src/pairing/carepartnerId.ts` is the stored-ID shim.
- `src/theme/theme.ts` holds the palette, type sizes and spacing from the visual design. Its test enforces the contrast and text-size floors in the design rules.
- `src/theme/fonts.ts` loads Atkinson Hyperlegible Next (regular, bold, extra bold) and exports `fonts`, the family names to use in styles. On Android a custom font's weight comes from its family name, so set `fontFamily: fonts.bold` and never `fontWeight`. `app/_layout.tsx` keeps the splash screen up until the fonts load, and carries on with the system font if they fail.
- `src/clock/clock.ts` is the only source of the current time (`now()`). `EXPO_PUBLIC_FAKE_TIME` starts the clock at a given time for manual checks, and tests call `setClock`. Never call `new Date()` for the current time in app code.
- `src/display/` keeps the screen on while the app runs (`expo-keep-awake`) and dims it at night. Night hours are part of the carepartner's schedule, set by the caregiver and stored on the server; the server doesn't send them yet, so `shared/display/nightHours.ts` (imported from `@shared`) supplies a default (20:00 to 7:00). `useTabletDisplay(nightHours)`, called in `app/_layout.tsx`, sets the app window's brightness low at night and returns it to the system setting by day, on Android only. It sets the window brightness, not the system setting, so the `expo-brightness` config plugin (which only adds the `WRITE_SETTINGS` permission) is deliberately left out of `app.json`.
- The app is locked to landscape in `app.json`. On the target tablet and the emulator it lays out at 1280 x 800 dp, the same size as the visual design.
- `src/widgets/` is the always-on screen's grid. `Grid` places each widget from its column, row, width and height (`layout.ts`: the shared 4 x 3 cell counts from `@shared` plus the tablet's 48 dp padding and 32 dp gaps, sized from the screen; `fitsGrid` and `GridArea` come from `@shared`) and skips widgets with an unknown renderer or that don't fit. `registry.ts` maps renderer names to components and says which sit in a tile frame (`WidgetFrame`). Renderers so far: `orientation` (day and time of day from the app clock, using `timeOfDay`, `dayName` and `dateLabel` from `@shared`), `photo` (family photo with a name caption, placeholder if missing) and `tile` (icon key, word, optional label; icon keys are ours, mapped in `renderers/icons.ts`). Types are in `types.ts`.
- Text that must fit a fixed width uses `fitFontSize` (`src/widgets/fitText.ts`), which works on web too; React Native's `adjustsFontSizeToFit` is ignored by the web build. Never shrink below the legibility floors in `theme.ts`.
- `app/preview.tsx` is a temporary route showing the grid on `sampleScreen` (`/preview` on web, `carepartner://preview` on Android). Delete it when the always-on screen becomes the home screen.

Every screen imports `API_URL` and calls `fetch` inline. There is no shared API client yet.

### Routes in `app/`

Expo Router treats each file in `app/` as a route. `index.tsx` and `pair.tsx` are the redesign's two routes. `screensaver.tsx` (the old home screen, still shown as the always-on screen until #316) and `terms.tsx` (a stub) are left over.

### Playback state machine (`src/playback/machine.ts`, `PlaybackProvider.tsx`)

One state machine decides what the screen shows. Its state is three facts: `pairing` (`checking`, `paired` or `unpaired`), `night`, and `playing` (the current play: a `playId`, the `MediaItem` and its `source`, or null). `screenFor` turns that into one screen: `starting` (nothing), `unpaired`, `playing`, `night` or `alwaysOn`. Playing wins over night, so a clip that runs past the start of night finishes first.

- Playback starts only through `usePlayback().play(item, source)`, never by navigating, so there is no player route and no back stack. `canPlay` holds the rules: nothing plays until paired; a `staff` play pre-empts anything (the newest wins); a `scheduled` or `tap` play starts only when nothing is playing, so it never interrupts a staff clip.
- When the player reports finished or failed, `ended(playId)` returns the screen to idle with no animation. A report whose `playId` isn't the current play is ignored, so a replaced player can't end the new one.
- `PlaybackProvider` (in `_layout.tsx`) reads pairing at startup and checks night once a minute through `src/clock`. A stored carepartner ID counts as paired for now, though the pair screen stores one before pairing completes (#343). After pairing succeeds, `pair.tsx` calls `checkPairing()` and returns to `/`.
- `app/index.tsx` only switches on the screen: an unpaired tablet is redirected to `/pair`, `playing` renders `MediaPlayer` keyed by `playId`, `night` renders `src/display/NightScreen.tsx` (a plain dark placeholder until #314), and `alwaysOn` renders `screensaver.tsx`.

### Schedule (`src/schedule/`)

`ScheduleProvider` (inside `PlaybackProvider`) asks the machine to play each due entry, one at a time, with source `scheduled`:
- Once paired, it fetches `/users/carepartner/:id` (for `auto_start`) and `/configurations/day?carepartner_id&day&month&day_of_week&date`, and again every 5 minutes and on Sync. `readEntries` drops malformed entries.
- Every 15 seconds it marks timed entries due at their minute (`HH:MM` match) and starts the untimed routine at the `auto_start` minute. A timed entry that comes due during another play waits, and plays next.
- Whenever `canPlay(state, "scheduled")`, it plays `nextEntry`: due timed entries first (earliest first), then, while the routine is running, the routine in `position` order. Entries `isPlayable` rejects (weather, news, messages, events) are skipped. An entry counts as played when it starts, so a failed one doesn't loop.
- On a new day (checked through `src/clock`) it forgets what played and fetches the new schedule.

The pure decisions are in `schedule.ts`. Played state lives in memory for now; it moves to the server's `Playback` records (#299), and the exact-minute match becomes a grace window (#298).

### Types mirror the backend by hand

`ScheduleEntry` (`src/schedule/schedule.ts`) is the part of the Go `ConfigurationWithMetadata` struct returned by `GetConfigurationsByDayAndCarepartner` that the tablet uses, and `MediaItem` (`src/playback/media.ts`) is its media part. Media metadata is checked in `readMetadata`, not typed, because the server's per-type metadata shapes differ (weather and news use upper-case keys such as `ID`).

### Media playback (`src/playback/`)

`MediaPlayer` takes a media item (`MediaItem`: `module_type`, optional `duration`, and `metadata` typed `unknown` because it is unchecked server JSON) plus `onFinished` / `onFailed`, and renders `VideoPlayer` (`expo-video`), `AudioPlayer` (`expo-audio`) or `PhotoPlayer` for it, or nothing for anything else. Server data enters playback in `src/playback/media.ts` (logic only, so the schedule can use it without loading the players): `readMetadata` checks an item's metadata and `isPlayable` says whether it will play. `MediaPlayer` builds the file URL (`/video|audio|photo/file?id=`), an audio item's title and cover, and a photo's duration. All three players start on their own, show no controls and work on Android and web. Each instance plays one source; give it a new `key` to play something else. `expo-audio` reports no load errors, so `AudioPlayer` treats a source that hasn't loaded within `AUDIO_LOAD_TIMEOUT_MS` as failed.

Every player keeps the contract written on `PlayerProps` in `src/playback/types.ts`: it starts on its own, calls exactly one of `onFinished` or `onFailed` once, and reports nothing after it is taken off screen. A shared suite in `players.test.tsx` checks it against every kind; a new player gets a driver there.

`PhotoPlayer` shows one photo whole (`contain`) for the required `durationSeconds`, counted from when the image loads, then reports finished. It reports failed on an image error or if the image hasn't loaded within `PHOTO_LOAD_TIMEOUT_MS`. It shows no caption.

`AudioPlayer` takes an optional `coverUri` (a song's album cover, or later a sender's photo) and falls back to a music note when it is missing or fails to load. `MediaPlayer` reads it from an optional `cover_uri` field on audio metadata, a full URL or a path relative to `API_URL` (resolved by `src/api/resolveApiUri.ts`). The server doesn't send `cover_uri` yet; once it extracts embedded album art and returns that field, covers appear with no tablet change.

The server doesn't yet guarantee a duration above 0 (#344), so photos show for `DEFAULT_DURATION_SECONDS` (30, in `media.ts`) when it is missing or not positive; remove that once #344 lands. The `expo-audio` plugin in `app.json` has its microphone permissions switched off; keep them off, since the tablet never records.

Media is streamed from the backend (`/photo|audio|video/file?id=`). Relative `image_uri` values are prefixed with `API_URL`. Plain `http://` is allowed in Android builds only for an `http://` server address (see `app.config.ts`).

### Pairing and identity (`app/pair.tsx`, `src/pairing/carepartnerId.ts`)

There is no login. On the pair screen, the device:
1. Validates its stored ID, or creates a new carepartner (`POST /users/carepartner/create-carepartner` with a placeholder name).
2. Requests a pairing code.
3. Polls `/users/device-pairing/status` when the user taps Verify.

The ID is stored under the key `carepartnerID`, via `carepartnerId.ts` (`localStorage` on web, `expo-secure-store` on native). `terms.tsx` is still a stub.

A tablet without a stored ID is sent to the pair screen by `index.tsx`; one with an ID counts as paired, even if pairing never completed (#343). Verifying succeeds whenever the status request does, even if the server reports `paired: false`. The home screen's Pair Device button also still opens it. If the stored ID is lost, the tablet creates a new carepartner row instead of reattaching to the existing one.

Nothing authenticates the device: the carepartner ID is its only credential, and it is a guessable integer. The intended fix is a device record with a secret token issued at pairing, which the tablet presents on every request and on its socket. That should be in place before tablets are deployed in residents' rooms.
