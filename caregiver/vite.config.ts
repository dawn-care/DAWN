import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// `shared/` sits outside this app, so the dev server has to be allowed to read it.
const sharedDir = fileURLToPath(new URL('../shared', import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@shared': fileURLToPath(new URL('../shared/index.ts', import.meta.url)),
    },
  },
  server: {
    fs: {
      allow: ['.', sharedDir],
    },
  },
})
