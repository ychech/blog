import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    // 开发时代理 API 请求到后端 Go 服务
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    // 生成的静态资源使用相对路径，方便与后端一起部署
    assetsDir: 'assets'
  },
  // 引入 .env 文件中的 VITE_ 前缀变量，便于配置不同环境的后端地址
  envPrefix: 'VITE_'
})
