import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import eslintPlugin from 'vite-plugin-eslint'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    eslintPlugin({
      lintOnStart: false,
      // cache: true,
      include: ['./src/**/*.ts', './src/**/*.tsx'],
      exclude: ['**/node_modules/**', '**/dist/**'],
    }),
  ]
})
