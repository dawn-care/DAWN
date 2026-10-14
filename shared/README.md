# shared

Rules used by both the caregiver app and the carepartner tablet, so the caregiver's preview always matches what the tablet shows.

## What's in it

| File | What it decides |
|---|---|
| `display/awakeHours.ts` | When the carepartner is awake. `isAwake`, `isStartAllowed` (an item must start between wake up and bed, on a 15 minute step, but may run past bed), `snapToStep`, `startTimeOptions` for the time picker, `itemsOutsideHours` for the warning before new hours are saved, `isValidAwakeHours`, and `parseClock` / `formatClock` for the server's time text. Default 8 am to 10 pm. |
| `display/nightHours.ts` | The tablet's current night hours. Stays until the night hours ticket moves the tablet to `awakeHours`. |
| `playback/currentlyPlaying.ts` | What the screen shows at a given time: her usual screen, the night screen, or a video, audio (over her usual screen or one photo) or timed photo that is playing. An item that starts outside the awake hours never plays. |
| `widgets/layoutPresets.ts` | The layouts (Simple, Balanced, Grid, Busier): each spot's place on the grid, `firstEmptySpot`, and `isSpotInLayout`. The server stores only a widget's `spot` number and how many spots a layout has. |
| `widgets/screenWidgets.ts` | Widget types and names, `shownWidgets` / `hiddenWidgets` for a layout, and the "doesn't fit" note. Widgets are never deleted when a smaller layout is chosen. |
| `widgets/weather.ts` | The weather tile's word ("Chilly", "Sunny"...). Takes the weather widget's data as it is. The words and temperature lines are provisional. |
| `widgets/timeOfDay.ts` | Morning, afternoon or evening, the day name and the date label. |
| `widgets/grid.ts` | The tablet's current 4 x 3 grid. Goes when the tablet moves from the old grid to layout spots. |

## What belongs here

- Plain TypeScript: functions, constants and types.
- Rules take the time and data as arguments. Nothing here reads the clock, the network or storage.

## What doesn't

- No imports of React, React Native, Expo or any DOM API. Anything that depends on the React Native theme, fonts or screen size stays in the app.
- No dependencies. `shared/` has no `package.json` or `node_modules`.

`npm run typecheck:shared` (from the repo root, using the carepartner's TypeScript, so run `npm install` in `carepartner/` first) compiles this folder with no DOM lib and no framework types, so such an import fails.

## Using it

Import from the barrel, never from a deep path:

```ts
import { currentlyPlaying, isAwake, weatherWord } from "@shared";
```

- **Carepartner:** the `@shared` path is set in `tsconfig.json`, Metro watches this folder and looks for packages in the app's `node_modules` (`metro.config.js`), and Jest maps it and looks there too (`roots`, `moduleNameMapper` and `modulePaths` in `package.json`). Files here sit outside the app, so without that lookup they could not find Babel's helpers.
- **Caregiver:** the `@shared` alias is in `vite.config.ts` (which also lets the dev server read this folder) and `paths` and `include` in `tsconfig.app.json`. The caregiver has no test runner, so these tests run from the carepartner.
- **Docker:** `server/Dockerfile` copies this folder next to `caregiver/`, so the caregiver build resolves `../shared`.

Add a module under a folder named for its feature, put its tests in a `__tests__/` folder next to it, and re-export it from `index.ts`. Write code that compiles under the strictest settings of both apps (`import type`, no extensions in imports, no unused locals).

## Testing

```bash
cd carepartner && npm test           # runs the tests in shared/ too
npm run typecheck:shared             # from the repo root
```
