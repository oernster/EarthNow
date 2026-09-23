import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// The test block is ported from ED Voyage Companion's frontend/vite.config.ts.
export default defineConfig({
    plugins: [react()],
    test: {
        environment: 'jsdom',
        globals: true,
        setupFiles: ['./src/test-setup.ts'],
        coverage: {
            provider: 'v8',
            include: ['src/**'],
            reporter: ['text'],
        },
    },
})
