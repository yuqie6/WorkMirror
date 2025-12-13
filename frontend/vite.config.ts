import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://127.0.0.1:7777',
        changeOrigin: true,
      },
      '/health': {
        target: process.env.VITE_API_TARGET || 'http://127.0.0.1:7777',
        changeOrigin: true,
      },
    }
  }
})
