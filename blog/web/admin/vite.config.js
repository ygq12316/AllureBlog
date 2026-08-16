import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    // naive-ui 按需引入：模板里的 n-* 组件编译期自动解析，替代全量 app.use(naive)
    Components({ resolvers: [NaiveUiResolver()] }),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
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
