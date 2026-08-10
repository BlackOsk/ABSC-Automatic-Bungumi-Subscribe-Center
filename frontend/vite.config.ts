import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // 将前端 /api/v1/... 代理转发至后端 http://localhost:8899/api/v1/...
      '/api': {
        target: 'http://localhost:8899',
        changeOrigin: true,
      },
    },
  },
})