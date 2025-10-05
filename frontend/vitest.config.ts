import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { svelteTesting } from '@testing-library/svelte/vite'

export default defineConfig({
  plugins: [svelteTesting(), svelte()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest.setup.ts']
  },
  optimizeDeps: {
    exclude: [
      'codemirror',
      '@codemirror/lang-javascript',
      '@codemirror/lang-go',
      '@codemirror/lang-php'
    ]
  }
})
