# shared

Rules used by both the caregiver app and the carepartner tablet, so the caregiver's preview always matches what the tablet shows.

## What's in it

| File | What it decides |
|---|---|
| `display/awakeHours.ts` | When the carepartner is awake. `isAwake`, `isStartAllowed` (an item must start between wake up and bed, on a 15 minute step, but may run past bed), `snapToStep`, `startTimeOptions` for the time picker, `itemsOutsideHours` for the warning before new hours are saved, `isValidAwakeHours`, and `parseClock` / `formatClock` for the server's time text. Default 8 am to 10 pm. |
| `display/nightHours.ts` | The tablet's night hours, and `nightHoursFromAwake` to turn the screen's awake hours into night hours. |
| `playback/currentlyPlaying.ts` | What the screen shows at a given time: her usual screen, the night screen, or a video, audio (over her usual screen or one photo) or timed photo that is playing. An item that starts outside the awake hours never plays. `overlappingItems` finds scheduled items that would play at the same time, so the caregiver can refuse to schedule them (pass the id of an item being edited so it doesn't overlap itself); the tablet doesn't pick a winner between overlaps, it finishes the item it is on. |
| `widgets/gridPlacement.ts` | Where widgets can go on the 4 x 3 grid. Each widget type has a fixed size (`WIDGET_SIZES` in `widgetSizes.ts`) and is placed as a type plus the `col` and `row` of its top left cell, counted from 0 at the top left. `openCells` lists the cells to highlight for the widget being placed, `canPlace` and `hasRoom` check one cell or the whole grid, and `isValidPlacement` checks the full list before it is saved. |
| `widgets/screenWidgets.ts` | Widget types and what to call them (`widgetName`). `splitKnownWidgets` sets aside widgets of a type this app doesn't know, so they are not placed or checked and can be sent back unchanged when saving. |
| `widgets/weather.ts` | The weather tile's word ("Chilly", "Sunny"...). Takes the weather widget's data as it is. The words and temperature lines are provisional. |
| `widgets/timeOfDay.ts` | Morning, afternoon or evening, the day name and the date label. |
| `widgets/dateLabels.ts` | Wording for the server's dates and times: `parseDate` ("YYYY-MM-DD" as a local date), `timeLabel` ("18:30" as "6:30 pm") and `ordinal` ("3rd"). |
| `widgets/grid.ts` | The screen's 4 x 3 grid: its size, `GridArea`, `fitsGrid`, `areasOverlap`, and `cornersToArea` / `areaToCorners` for the screen layout API's inclusive corners. |
| `widgets/widgetSizes.ts` | Each widget type's one fixed size (`WIDGET_SIZES`, `widgetSize`, `hasWidgetSize`). The tablet skips a widget that isn't its type's size. |
| `widgets/textLimits.ts` | The longest message and reminder text the tablet shows whole (`TEXT_LIMITS`), for the caregiver's editor to enforce. |

## What belongs here

- Plain TypeScript: functions, constants and types.
- Rules take the time and data as arguments. Nothing here reads the clock, the network or storage.

## What doesn't

- No imports of React, React Native, Expo or any DOM API. Anything that depends on the React Native theme, fonts or screen size stays in the app.
- No dependencies. `shared/` has no `package.json` or `node_modules`.

`npm run typecheck:shared` (from the repo root, using the carepartner's TypeScript, so run `npm install` in `carepartner/` first) compiles this folder with no DOM lib and no framework types, so such an import fails. It uses `tsconfig.lib.json`, which leaves out the tests. `tsconfig.json` is for the editor: it also covers the tests, with Jest's types taken from `carepartner/node_modules`.

## Using it

Import from the barrel, never from a deep path:

```ts
import { currentlyPlaying, isAwake, weatherWord } from "@shared";
```

- **Carepartner:** the `@shared` path is set in `tsconfig.json`, Metro watches this folder and looks for packages in the app's `node_modules` (`metro.config.js`), and Jest maps it and looks there too (`roots`, `moduleNameMapper` and `modulePaths` in `package.json`). Files here sit outside the app, so without that lookup they could not find Babel's helpers.
- **Caregiver:** the `@shared` alias is in `vite.config.ts` (which also lets the dev server read this folder) and `paths` and `include` in `tsconfig.app.json`. The caregiver has no test runner, so these tests run from the carepartner.
- **Docker:** `server/Dockerfile` and `caregiver/Dockerfile` copy this folder next to `caregiver/`, so the caregiver build resolves `../shared`. Both build from the repo root (`docker build -f caregiver/Dockerfile .`).

Add a module under a folder named for its feature, put its tests in a `__tests__/` folder next to it, and re-export it from `index.ts`. Write code that compiles under the strictest settings of both apps (`import type`, no extensions in imports, no unused locals).

## Testing

```bash
cd carepartner && npm test           # runs the tests in shared/ too
npm run typecheck:shared             # from the repo root
```
