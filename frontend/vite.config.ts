import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: [
      // The v3 binding generator emits `import ... from "/wails/runtime.js"`.
      // At runtime Wails serves the runtime under that path, but for the
      // bundled build we redirect to the npm package so Rollup can resolve
      // the type/symbol surface.
      { find: /^\/wails\/runtime\.js$/, replacement: '@wailsio/runtime' },
      { find: '@bindings', replacement: resolve(__dirname, './src/bindings/github.com/minjejeon/convert4share/internal') },
      { find: '@', replacement: resolve(__dirname, './src') },
    ],
  },
})
