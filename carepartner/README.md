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
- `src/` holds everything a route is made of, grouped by feature (`api/`, `pairing/`, `theme/`). Each feature's tests are in its `__tests__/` folder.
- `.maestro/` holds end-to-end flows.

## Backend address

The app talks to the Go backend at `EXPO_PUBLIC_API_URL`, which defaults to `http://localhost:8080/api`. To point it elsewhere, copy `.env.example` to `.env.local` and edit it. On an Android emulator the dev machine is `10.0.2.2`, not `localhost`.

## Tests

```bash
npm test        # unit and component tests (Jest + React Native Testing Library)
npm run e2e     # end-to-end flows (Maestro) against the app installed on an emulator or tablet
```

Unit tests live in a `__tests__/` folder inside each feature folder (`src/pairing/carepartnerId.ts` is covered by `src/pairing/__tests__/carepartnerId.test.ts`). Keep test files out of `app/`, where they would become routes.

The end-to-end flows in `.maestro/` need [Maestro](https://maestro.dev) installed and a build of the app on a running emulator or connected tablet:

```bash
npx expo run:android --variant release   # needs JDK 17 or 21
npm run e2e
```

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
