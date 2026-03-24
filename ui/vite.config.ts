import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { TanStackRouterVite } from '@tanstack/router-vite-plugin'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    TanStackRouterVite({
      routesDirectory: './src/routes',
      generatedRouteTree: './src/routeTree.gen.ts',
    }),
  ],
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('node_modules')) {
            if (id.includes('@mantine') || id.includes('@emotion')) return 'vendor-mantine';
            if (id.includes('@tanstack')) return 'vendor-tanstack';
            if (id.includes('@xyflow')) return 'vendor-flow';
            if (id.includes('lucide-react')) return 'vendor-icons';
            if (id.includes('node_modules/react/') || id.includes('node_modules/react-dom/') || id.includes('node_modules/scheduler/')) return 'vendor-react';
            return 'vendor';
          }
        },
      },
    },
  },
})
