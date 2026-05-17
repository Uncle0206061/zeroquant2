import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [VantResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [VantResolver()],
      dts: 'src/components.d.ts',
    }),

  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    target: 'es2020',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('element-plus')) return 'element-plus'
          if (id.includes('vant')) return 'vant'
          if (id.includes('echarts')) return 'echarts'
          if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) return 'vendor'
        },
      },
    },
    chunkSizeWarningLimit: 600,
  },
  server: {
    port: 5173,
    host: process.env.NODE_ENV === 'development', // 仅开发环境开启局域网访问
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/data': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  optimizeDeps: {
    include: ['vant', 'echarts'],
  },
})
