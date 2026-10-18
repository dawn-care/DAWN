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
- Nothing requires a tap. Tapping is at most an optional shortcut. The pair screen is the exception: the person setting up the tablet is in the room and does the tapping, so it may have buttons.
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
- The screen is a grid of 4 columns × 3 rows. Everything on it is a widget, including day and time, the photo and "what's next". Each widget type has **one fixed size** (`WIDGET_SIZES` in `@shared`); a different size is a different type. Day and time is locked.
- The caregiver places each widget anywhere on the grid (since #364), and the server stores and sends its position explicitly (column, row, width, height). The server doesn't check the fixed sizes yet, so the tablet skips a widget that isn't its type's size.
- The tablet has one renderer per *look*, not per widget: orientation (day and time), photo, next event, person, a generic tile (icon, word, short label) shared by weather and future 1×1 data widgets, and a wide tile (icon, label, word, detail) for 2×1 ones such as sports. An unknown renderer is skipped.
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

The repo-root `shared/` folder holds plain TypeScript rules used by this app and the caregiver app: `NightHours`, `DEFAULT_NIGHT_HOURS` and `isNight` (`display/nightHours.ts`), `timeOfDay`, `TimeOfDay`, `dayName` and `dateLabel` (`widgets/timeOfDay.ts`), and `GRID` (column and row counts only), `GridArea`, `fitsGrid`, `areasOverlap` and the corner conversions (`widgets/grid.ts`), `WIDGET_SIZES`, `widgetSize` and `hasWidgetSize` (`widgets/widgetSizes.ts`), and `TEXT_LIMITS` (`widgets/textLimits.ts`), the longest message and reminder text the tablet shows whole. `nightHoursFromAwake` turns the screen's awake hours into night hours. It also has rules this app does not use yet: awake hours (`display/awakeHours.ts`), what is playing at a given time (`playback/currentlyPlaying.ts`), grid placement (`widgets/gridPlacement.ts`: the cells a widget can go in, using `WIDGET_SIZES`), widget names (`widgets/screenWidgets.ts`) and the weather word (`widgets/weather.ts`). See `shared/README.md`.

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
- `src/storage/storage.ts` keeps small values on the device across restarts (`localStorage` on web, `expo-secure-store` on native). `src/pairing/carepartnerId.ts` stores the paired ID with it.
- `src/theme/theme.ts` holds the palette, type sizes and spacing from the visual design. Its test enforces the contrast and text-size floors in the design rules.
- `src/theme/fonts.ts` loads Atkinson Hyperlegible Next (regular, bold, extra bold) and exports `fonts`, the family names to use in styles. On Android a custom font's weight comes from its family name, so set `fontFamily: fonts.bold` and never `fontWeight`. `app/_layout.tsx` keeps the splash screen up until the fonts load, and carries on with the system font if they fail.
- `src/clock/clock.ts` is the only source of the current time (`now()`), and has `dayOf` for the local "YYYY-MM-DD" and `everyMinute`, which calls back as each minute of the app clock starts. Checks of the time (night, the schedule, `useNow`) use it so they happen on the minute, not up to a minute late. `EXPO_PUBLIC_FAKE_TIME` starts the clock at a given time for manual checks, and tests call `setClock`. Never call `new Date()` for the current time in app code.
- `app/_layout.tsx` keeps the screen on while the app runs (`expo-keep-awake`) and hides Android's status and navigation bars (`src/display/useHiddenSystemBars.ts`, through `expo-status-bar` and `expo-navigation-bar`), again whenever the app returns from the background. Staff can swipe a bar back in for a moment; it hides again by itself. Adding `expo-navigation-bar` means the app has to be rebuilt (`npm run android`), not just reloaded. `src/display/TabletDisplay.tsx` (inside `ScreenProvider`, draws nothing) decides when it is night: from the carepartner's bed time to their wake time (the screen's `awakeHours`, through `nightHoursFromAwake`), or the default 20:00 to 7:00 (`DEFAULT_NIGHT_HOURS`) until the first screen arrives. As each minute starts, and whenever the hours change, it tells the machine through `usePlayback().setNight`. It also calls `useTabletDisplay(nightHours)`, which sets the app window's brightness low at night and returns it to the system setting by day, on Android only. It sets the window brightness, not the system setting, so the `expo-brightness` config plugin (which only adds the `WRITE_SETTINGS` permission) is deliberately left out of `app.json`.
- The app is locked to landscape in `app.json`. On the target tablet and the emulator it lays out at 1280 x 800 dp, the same size as the visual design.
- `src/widgets/` is the always-on screen's grid. `Grid` places each widget from its column, row, width and height (`layout.ts`: the shared 4 x 3 cell counts from `@shared` plus the tablet's 48 dp padding and 32 dp gaps, sized from the screen; `fitsGrid` and `GridArea` come from `@shared`) and skips widgets with an unknown renderer or that don't fit. `registry.ts` maps renderer names to components and says which sit in a tile frame (`WidgetFrame`). Renderers: `orientation` (day and time of day from the app clock, using `timeOfDay`, `dayName` and `dateLabel` from `@shared`), `photo` (family photo with a name caption, placeholder if missing), `tile` (icon key, word, optional label; icon keys are ours, mapped in `renderers/icons.ts`), `message` (a caregiver's message with an optional role and photo), `reminder` (the next two of the day's reminders by the app clock, the soonest marked; long text shrinks to the 26px floor over two lines) `headline` (one news headline) and `wideTile` (a 2×1 data widget: icon, a short label, a word, and an optional detail line). Types are in `types.ts`.
- Text that must fit a fixed width uses `fitFontSize` (`src/widgets/fitText.ts`), and wrapped text that must fit a box uses `fitFontSizeToBox` with `linesAt`; both work on web too, where React Native's `adjustsFontSizeToFit` is ignored. Never shrink below the legibility floors in `theme.ts`.

`src/api/request.ts` has `getJson` and `postJson`: the server address, a timeout (`REQUEST_TIMEOUT_MS`, 10 s by default), and an `HttpError` (with the `status`) for a failed status, so a caller can tell the server saying no from the server not answering. They return the JSON unchecked, for the caller to check where it enters the app.

### Routes in `app/`

Expo Router treats each file in `app/` as a route. `index.tsx` and `pair.tsx` are the redesign's two routes. `terms.tsx` (a stub) is left over.

### Playback state machine (`src/playback/machine.ts`, `PlaybackProvider.tsx`)

One state machine decides what the screen shows. Its state is three facts: `pairing` (`checking`, `paired` or `unpaired`, plus the paired `carepartnerId`), `night`, and `playing` (the current play: a `playId`, the `MediaItem`, its `source` and, for staff plays, the server's `playbackId`; or null). It also records `lastEnded`, the last play to leave the screen and why (`finished`, `failed`, `stopped`, or `replaced` by a staff play), which is what play reports watch. `screenFor` turns that into one screen: `starting` (nothing), `unpaired`, `playing`, `night` or `alwaysOn`. Playing wins over night, so a clip that runs past the start of night finishes first.

- Playback starts only through `usePlayback().play(item, source)`, never by navigating, so there is no player route and no back stack. `canPlay` holds the rules: nothing plays until paired; a `staff` play pre-empts anything (the newest wins); a `scheduled` or `tap` play starts only when nothing is playing, so it never interrupts a staff clip.
- When the player reports finished or failed, `ended(playId, reason)` returns the screen to idle with no animation. A report whose `playId` isn't the current play is ignored, so a replaced player can't end the new one. `stop(playbackId)` stops only the play with that record, so a late stop can't end a newer clip.
- `PlaybackProvider` (in `_layout.tsx`) reads pairing at startup (`readPairedId`, see "Pairing and identity"), and tries again every `PAIRING_RETRY_MS` while storage can't be read; `TabletDisplay` tells it when night starts and ends. The machine keeps the paired carepartner ID, so the screen and the socket take it from `usePlayback().state.carepartnerId` instead of reading storage again. After pairing succeeds, `pair.tsx` calls `checkPairing()` and returns to `/`.
- `app/index.tsx` only switches on the screen: an unpaired tablet is redirected to `/pair`, `playing` renders `MediaPlayer` keyed by `playId`, `night` renders `src/display/NightScreen.tsx` (dark and warm, with the day and "You're safe. Sleep well."), and `alwaysOn` renders `src/screen/AlwaysOnScreen.tsx`.

### Staff commands (`src/playback/useStaffCommands.ts`, `socket.ts`, `socketMessages.ts`, `staffItem.ts`)

Staff play and stop clips from their phones; the server pushes each command to the tablet over a WebSocket (`/realtime/tablet?carepartner_id=`, see the server's `internal/realtime/messages.go`). `StaffCommands`, rendered in `_layout.tsx` inside both providers, runs `useStaffCommands` and draws nothing.

- `socket.ts` connects once paired, reconnects after a drop (1 s, doubling to at most 30 s, back to 1 s once connected), and holds messages sent while disconnected until it reconnects. The server pings every 25 s to keep the connection open; the WebSocket answers by itself. `socketMessages.ts` checks every server message and drops malformed ones.
- `play`: the tablet reports `acknowledged` and looks the media up (`staffItem.ts`, `GET /media/{id}`, given up after 5 s) at the same time. It plays the item with source `staff`, which pre-empts anything, only once the server's answer to the acknowledgement is `acknowledged`. The server judges expiry, since the tablet's clock (or `EXPO_PUBLIC_FAKE_TIME`) can be wrong; any other answer means the command must not start. A newer command cancels an older one still waiting and reports it `stopped`, so a slow lookup can't let an older command take over. The server pushes the latest live command on every reconnect, which is how the tablet "asks for" it; a `playback_id` already handled isn't started twice, but is acknowledged again if the server's answer never came. It reports `failed` if the media can't be looked up or played.
- `stop` stops the play with that record, or cancels one still waiting to start. `refresh` fetches the screen again.
- Reports follow the machine: `started` when a staff play is on screen, then `finished`, `failed` or `stopped` (also when another staff play replaces it) from `lastEnded`. Scheduled plays aren't reported yet (#299).

### Screen (`src/screen/`)

`ScreenProvider` (in `_layout.tsx`, inside `PlaybackProvider`) keeps the server's `GET /screen?carepartner_id=` response (see `docs/api/screen-and-media.md`), which holds the awake hours, the widgets with their places on the grid, and the schedule. `useScreen()` gives `screen` (null until the first good response) and `refresh()`.
- Once paired, it fetches the screen, then checks every minute and fetches again when 5 minutes have passed since the last try or the day has changed. The socket's `refresh` message calls `refresh()`.
- Only the latest fetch counts: a response that arrives after a newer fetch started, or after the carepartner changed, is dropped.
- `readScreen` (`screen.ts`) checks the response once: a malformed widget or a schedule entry the tablet can't play is dropped, and the rest kept. A failed or malformed response keeps the last good screen.
- `hideStaleWidgets` removes weather, sports and news widgets once the last good fetch is `STALE_AFTER_MS` (3 hours) old, so old data is never shown as current.
- `AlwaysOnScreen` draws the widgets on the grid, through `placementsFor` (`placements.ts`, logic only). It checks each type's `config` and `data` and maps it to a renderer: `photo` → photo, `weather` (temperature and a plain condition word) → tile, `sports` (the team, "Lost 3–4" or "No games yet", and "vs Tampa Bay Rays") → wideTile, `message` → message, `reminder` → reminder, `news` → headline, and `day_time` → orientation. A widget is left out, so its cells stay empty, if its type is unknown, it isn't its type's size, it runs off the grid or overlaps a widget before it, its outside service is down (`error`), or its data is missing. Only a count is logged.
- When there is nothing to draw (before the first screen arrives, with no widgets, or only ones it left out), `AlwaysOnScreen` shows the day, time of day and date instead, so the screen is never blank.
- Day and time on the grid is the `day_time` type, which the server doesn't have yet (#366), so a screen with widgets shows no day and time until it does.

### Schedule (`src/schedule/`)

`SchedulePlayer` (in `_layout.tsx`, draws nothing) plays the screen's `schedule`, today's and tomorrow's plays, through the machine with source `scheduled`:
- As each minute starts, and whenever the screen or the machine changes, it plays `duePlay` if `canPlay(state, "scheduled")`: the earliest play whose time came within the last `GRACE_MS` (10 minutes) and that hasn't played. Later than that it is skipped, so a play missed while the tablet was off or busy with a staff clip doesn't start long after its time. Plays flagged `outsideAwakeHours` are skipped.
- A play counts as played once the machine puts it on screen, so a failed one doesn't loop, and one the machine refuses (a staff clip got there first) can still play later. Played plays (`playKey`: scheduled item and date) are stored on the device by `playedStore.ts`, so a restart within the grace window doesn't repeat one. Nothing plays until they've loaded. This moves to the server's `Playback` records once scheduled plays are reported (#299).
- Times are compared in the tablet's local time, which is assumed to be the carepartner's timezone.

The pure decisions are in `schedule.ts`.

### Developer toolbar (`src/dev/`)

In development builds only (`__DEV__`; `app/_layout.tsx` leaves it out of release builds), `DevToolbar` draws two buttons in the top right, above every screen:
- ☰ opens a panel with the app clock (marked when `EXPO_PUBLIC_FAKE_TIME` is set), the screen and what's playing, the awake hours and whether it's night, the carepartner and pairing, whether the server sees the socket, the last good screen fetch, today's and tomorrow's scheduled plays with their status (played, due, upcoming, missed, asleep), and the carepartner's media. Tapping a media item starts a staff play through `POST /playbacks/play`, the same path the staff app will use (#302).
- ⏭, shown while something plays, stops it as a staff stop would.

Its server calls and the play statuses are in `devInfo.ts`. It uses its own plain styling, not the resident's theme, and none of the design rules apply to it.

### Types mirror the backend by hand

`Screen`, `ScreenWidget` and `ScheduledPlay` (`src/screen/screen.ts`) are the parts of the Go `Screen` struct (`server/internal/services/screen/types.go`) that the tablet uses, renamed to camelCase. `MediaItem` (`src/playback/media.ts`) is what playback needs from a media object; `readMedia` checks one from `/media/{id}` or a schedule entry. A widget's `config` and `data` stay `unknown` until renderers check them for their type (#316).

### Media playback (`src/playback/`)

`MediaPlayer` takes a media item (`MediaItem`: `kind`, the file's `url`, and optionally a `title` and `coverUrl` for audio and `durationSeconds` for a photo) plus `onFinished` / `onFailed`, and renders `VideoPlayer` (`expo-video`), `AudioPlayer` (`expo-audio`) or `PhotoPlayer` for it. Server media enters playback through `readMedia` in `src/playback/media.ts` (logic only, so the schedule can use it without loading the players), which accepts only the three kinds the tablet plays. All three players start on their own, show no controls and work on Android and web. Each instance plays one source; give it a new `key` to play something else. `expo-audio` reports no load errors, so `AudioPlayer` treats a source that hasn't loaded within `AUDIO_LOAD_TIMEOUT_MS` as failed. Android pauses a video when the app goes to the background and doesn't resume it, so `VideoPlayer` plays again when the app returns to the foreground.

Every player keeps the contract written on `PlayerProps` in `src/playback/types.ts`: it starts on its own, calls exactly one of `onFinished` or `onFailed` once, and reports nothing after it is taken off screen. Players report through `useReportOnce`, which lets only the first report through and always calls the latest callbacks, so a parent passing new functions on every render doesn't restart a player's timers. A shared suite in `players.test.tsx` checks it against every kind; a new player gets a driver there.

`PhotoPlayer` shows one photo whole (`contain`) for the required `durationSeconds`, counted from when the image loads, then reports finished. It reports failed on an image error or if the image hasn't loaded within `PHOTO_LOAD_TIMEOUT_MS`. It shows no caption.

`AudioPlayer` takes an optional `coverUri` and falls back to a music note when it is missing or fails to load. A scheduled `audio_with_background` item's background photo is its cover; staff plays of audio have none.

A scheduled photo shows for its display time. A staff-started photo has none, so it shows for `DEFAULT_DURATION_SECONDS` (30, in `media.ts`), as does one whose time is missing or not positive. The `expo-audio` plugin in `app.json` has its microphone permissions switched off; keep them off, since the tablet never records.

Media is streamed from the `url` the server gives (`/media/{id}/file`), resolved against `API_URL` by `src/api/resolveApiUri.ts`. Plain `http://` is allowed in Android builds only for an `http://` server address (see `app.config.ts`).

### Pairing and identity (`app/pair.tsx`, `src/pairing/`)

There is no login. `app/pair.tsx` runs `usePairing` (`src/pairing/usePairing.ts`, logic) and draws `PairScreen` (visual): "Set up this tablet", a large code, "Enter this code in the family app", a seconds countdown and a "New code" button. Pairing needs no tap:
1. It finds this tablet's carepartner: the stored ID if `findCarepartner` (`pairingApi.ts`) says the server still has it, or a new one (`POST /users/carepartner/create-carepartner` with no name: the server uses a placeholder until the caregiver names the resident when entering the code) if there is none or the server answers 404. While the server can't be reached it shows "Can't reach the server. Trying again…" and retries every `PAIRING_POLL_MS` (5 s), and never creates one, so a tablet that is briefly offline keeps its resident.
2. It requests a pairing code (10 minutes) and shows it.
3. Every 5 s it checks `/users/device-pairing/status`. Once that reports `paired: true` it records pairing, calls `checkPairing()` and returns to `/`.
4. When the code expires, or "New code" is tapped, it checks the status once more and then gets a new code, so a code used just before it expired still completes, and no unused code is made after a claim (the status reads only the newest code, #371). A check that fails while the code still works leaves it on screen.

The QR code #315 asks for is deferred until the staff or family app can read one.

The ID is stored under the key `carepartnerID`, via `carepartnerId.ts` (`localStorage` on web, `expo-secure-store` on native), and that pairing completed under a separate key (`pairingComplete`), since the pair screen stores the ID before a caregiver claims it. Storing a new ID clears it. `terms.tsx` is still a stub.

`readPairedId` (`pairingState.ts`) is what counts as paired: a stored ID with pairing recorded. A tablet that paired before pairing was recorded has an ID but no record, so the server is asked once and the answer recorded; if it can't be reached, the tablet keeps counting as paired and asks again next launch. An unpaired tablet is sent to the pair screen by `index.tsx`; nothing else opens it. If the stored ID is lost, the tablet creates a new carepartner row instead of reattaching to the existing one.

Nothing authenticates the device: the carepartner ID is its only credential, and it is a guessable integer. The intended fix is a device record with a secret token issued at pairing, which the tablet presents on every request and on its socket. That should be in place before tablets are deployed in residents' rooms.
