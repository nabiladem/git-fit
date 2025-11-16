import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],

  base: '/',

  // Define development server configuration
  server: {
    port: 5173,

    // Set up a proxy for API calls to your Go backend during development
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },

  build: {
    outDir: './dist',
    assetsDir: 'assets',
    sourcemap: true,
  },

  css: {
    modules: {
      scopeBehaviour: 'local',
    },
  },

  define: {
    'process.env': {},
  },
})
