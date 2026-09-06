import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import Components from 'unplugin-vue-components/vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    // ui 基件按需自动引入（src/components/ui），其余组件显式 import
    Components({ dirs: ['src/components/ui'] }),
  ],
  server: {
    proxy: {
      // ws:true — 评论 WS(/api/notes/:id/ws) 与全局弹幕 WS(/api/ws) 经 Vite 转发升级
      '/api': { target: 'http://localhost:8080', ws: true },
      '/uploads': 'http://localhost:8080',
      // 笔墨精灵 agent — 生产由 Caddy 转发，开发模式在此代理（含 WebSocket）
      '/chat': { target: 'http://localhost:8000', ws: true },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
