import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import eslintPlugin from 'vite-plugin-eslint'
import { vanillaExtractPlugin } from '@vanilla-extract/vite-plugin'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    vanillaExtractPlugin(),
    eslintPlugin({
      lintOnStart: false,
      cache: true,
      include: ['./src/**/*.ts', './src/**/*.tsx'],
      exclude: ['**/node_modules/**', '**/dist/**'],
    }),
  ],
})
