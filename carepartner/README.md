# Welcome to your Expo app 👋

This is an [Expo](https://expo.dev) project created with [`create-expo-app`](https://www.npmjs.com/package/create-expo-app).

## Get started

1. Install dependencies

   ```bash
   npm install
   ```

2. Start the app

   ```bash
   npx expo start
   ```

In the output, you'll find options to open the app in a

- [development build](https://docs.expo.dev/develop/development-builds/introduction/)
- [Android emulator](https://docs.expo.dev/workflow/android-studio-emulator/)
- [iOS simulator](https://docs.expo.dev/workflow/ios-simulator/)
- [Expo Go](https://expo.dev/go), a limited sandbox for trying out app development with Expo

You can start developing by editing the files inside the **app** directory. This project uses [file-based routing](https://docs.expo.dev/router/introduction).

## Project layout

- `app/` holds routes only. Every file in it becomes a screen.
- `src/` holds everything a route is made of, grouped by feature (`api/`, `clock/`, `pairing/`, `playback/`, `theme/`). Each feature's tests are in its `__tests__/` folder.
- `.maestro/` holds end-to-end flows.

## Backend address

With no setup, the app talks to a backend running on your computer at port 8080:

- **Web:** `http://localhost:8080/api`
- **Android emulator:** `http://10.0.2.2:8080/api`, which is how the emulator reaches the computer it runs on. No `adb reverse` is needed.

To use any other backend, copy `.env.example` to `.env.local` and set `EXPO_PUBLIC_API_URL`. A real tablet always needs this: use your computer's address on the local network (for example `http://192.168.1.20:8080/api`) or the deployed server.

Android blocks plain `http://` connections in release builds. The build allows them automatically when the server address is `http://` (including the local defaults above) and keeps them blocked for an `https://` server, so production builds stay secure. After changing `EXPO_PUBLIC_API_URL`, rebuild the app.

## Testing at a different time of day

To see how the app behaves at another time, set `EXPO_PUBLIC_FAKE_TIME` in `.env.local` (for example `EXPO_PUBLIC_FAKE_TIME=2026-10-03T21:30:00`). The app's clock starts at that time and runs forward normally. Leave it unset for the real time. Code reads the time through `src/clock`, never `new Date()` directly, so this works everywhere.

## Tests

```bash
npm test        # unit and component tests (Jest + React Native Testing Library)
npm run e2e     # end-to-end flows (Maestro) against the app installed on an emulator or tablet
```

Unit tests live in a `__tests__/` folder inside each feature folder (`src/pairing/carepartnerId.ts` is covered by `src/pairing/__tests__/carepartnerId.test.ts`). Keep test files out of `app/`, where they would become routes.

The end-to-end flows in `.maestro/` need [Maestro](https://maestro.dev) installed and a build of the app on a running emulator or connected tablet:

Build first with the emulator closed, then start the emulator. Running both at once needs a lot of memory.

```bash
npm run build:emulator   # release build for an x86_64 emulator; needs JDK 17 or 21 (set JAVA_HOME)
# start the emulator, then:
npm run e2e:install
npm run e2e
```

`build:emulator` builds for the emulator's processor only. A real tablet needs `arm64-v8a`; use `npx expo run:android --variant release` with the tablet connected.

The emulator starts at Android's home screen and doesn't open the app by itself. Open it from the app drawer (swipe up, tap **carepartner**) or with `adb shell am start -n com.anonymous.carepartner/.MainActivity`. The app stays installed across emulator restarts; reinstall only after code changes.

### Known issue: emulator sound

On Linux, the emulator's sound can stop working after the first video or song: later items play (their position advances and they move on when they end) but are silent, and the emulator's log reports a stalled audio device. Restarting the emulator with `-no-snapshot-load` brings sound back for the first item only. This hasn't been seen in the browser. Check sound in the web build or on a real Android device, and use the emulator for everything else.

## Get a fresh project

When you're ready, run:

```bash
npm run reset-project
```

This command will move the starter code to the **app-example** directory and create a blank **app** directory where you can start developing.

## Learn more

To learn more about developing your project with Expo, look at the following resources:

- [Expo documentation](https://docs.expo.dev/): Learn fundamentals, or go into advanced topics with our [guides](https://docs.expo.dev/guides).
- [Learn Expo tutorial](https://docs.expo.dev/tutorial/introduction/): Follow a step-by-step tutorial where you'll create a project that runs on Android, iOS, and the web.

## Join the community

Join our community of developers creating universal apps.

- [Expo on GitHub](https://github.com/expo/expo): View our open source platform and contribute.
- [Discord community](https://chat.expo.dev): Chat with Expo users and ask questions.
