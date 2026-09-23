import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// The test block is ported from ED Voyage Companion's frontend/vite.config.ts.
export default defineConfig({
    plugins: [react()],
    // The auto-scroll machine has one home, beside the setup page it also serves
    // (installer/frontend/dist/auto-scroll.js), so the dev server may read there.
    server: {fs: {allow: ['.', '../installer/frontend/dist']}},
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
