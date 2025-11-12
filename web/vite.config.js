import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],

  // Base path for the build (you can adjust this based on your project needs)
  base: '/',

  // Define development server configuration
  server: {
    // The port Vite dev server will run on
    port: 5173, // Default for Vite

    // Set up a proxy for API calls to your Go backend during development
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // Go backend (API) port
        changeOrigin: true, // Ensures that the host header is rewritten
        secure: false, // Allows connection even if there is no SSL (useful in dev)
        rewrite: (path) => path.replace(/^\/api/, ''), // Optional: remove /api from request path
      },
    },
  },

  // Optional: You can configure build options if needed for your project
  build: {
    outDir: './dist', // The output directory for the built app
    assetsDir: 'assets', // Directory for assets (images, etc.)
    sourcemap: true, // Enable source maps for debugging (optional)
  },

  // Optional: Configure how Vite handles CSS, JS, and other assets
  css: {
    modules: {
      scopeBehaviour: 'local', // Use local scoping for CSS modules
    },
  },

  // Optional: Configure environment variables if you need them
  define: {
    'process.env': {},
  },
})
