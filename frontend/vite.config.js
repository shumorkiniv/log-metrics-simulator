import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        host: '0.0.0.0',
        port: 5173,
        proxy: {
            '/api': {
                target: process.env.VITE_API_URL || 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
            },
            // Прокси для health и metrics эндпоинтов бэкенда в дев-режиме
            '/health': {
                target: process.env.VITE_API_URL || 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
            },
            '/metrics': {
                target: process.env.VITE_API_URL || 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
            },
        }
    },
    preview: {
        host: '0.0.0.0',
        port: 3000
    },
    build: {
        outDir: 'dist',
        sourcemap: false,
        minify: 'terser',
        terserOptions: {
            compress: {
                drop_console: true,
                drop_debugger: true,
            },
        },
        rollupOptions: {
            output: {
                manualChunks: {
                    vendor: ['react', 'react-dom'],
                    ui: ['antd'],
                    charts: ['recharts'],
                    utils: ['axios', 'dayjs']
                }
            }
        }
    }
})