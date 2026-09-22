import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// 前端开发服务器配置。/api 请求代理到 Go 后端，避免开发期跨域问题。
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
