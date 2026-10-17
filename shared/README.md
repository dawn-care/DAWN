# shared

Rules used by both the caregiver app and the carepartner tablet, so the caregiver's preview always matches what the tablet shows.

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
import { isNight, timeOfDay } from "@shared";
```

- **Carepartner:** the `@shared` path is set in `tsconfig.json`, Metro watches this folder (`metro.config.js`), and Jest maps it (`package.json`).
- **Caregiver:** not wired up yet (deferred to a follow-up ticket, see #354). Once it is, it should also run the tests in this folder.

Add a module under a folder named for its feature, put its tests in a `__tests__/` folder next to it, and re-export it from `index.ts`. Write code that compiles under the strictest settings of both apps (`import type`, no extensions in imports, no unused locals).

## Testing

```bash
cd carepartner && npm test           # runs the tests in shared/ too
npm run typecheck:shared             # from the repo root
```
