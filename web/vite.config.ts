import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true, // proxy WebSocket connections (collab hub)
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    chunkSizeWarningLimit: 700,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/highlight.js')) return 'highlight'
          if (id.includes('node_modules/marked')) return 'markdown-runtime'
          if (id.includes('node_modules/yjs') || id.includes('node_modules/y-websocket') || id.includes('node_modules/y-protocols') || id.includes('node_modules/lib0') || id.includes('node_modules/y-codemirror')) return 'yjs'
          if (id.includes('node_modules/@codemirror') || id.includes('node_modules/@lezer')) return 'codemirror'
          if (id.includes('node_modules/vue') || id.includes('node_modules/vue-router') || id.includes('node_modules/pinia')) {
            return 'vue-core'
          }
        },
      },
    },
  },
  plugins: [tailwindcss(), vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
