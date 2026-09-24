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
        // istanbul rather than v8: v8 reported GlobeView.tsx at 100% with no
        // test importing it, then istanbul read it at 0% (measured 2026-09-24).
        // The floors are the measured figures, not targets (measured again on
        // 2026-09-24, after the cloud layer's tests): the drawing itself
        // is exercised by eye, as TESTING.md says.
        // A floor above what is measured only teaches lowering it.
        coverage: {
            provider: 'istanbul',
            include: ['src/**'],
            // main.tsx and App.tsx are the page's composition root, as main.go is
            // the Go side's: they wire the parts together and are checked by eye.
            exclude: ['src/**/*.test.{ts,tsx}', 'src/test-setup.ts', 'src/main.tsx', 'src/App.tsx'],
            reporter: ['text-summary'],
            thresholds: {statements: 82.59, branches: 75.78, functions: 80.22, lines: 84.69},
        },
    },
})
