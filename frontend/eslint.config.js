// What the page is checked for; why each rule is here. Ported from Bridge Talk's
// frontend/eslint.config.js.
//
// The Go side has gofmt, vet and staticcheck standing over it, plus a suite that
// refuses to pass below its own bar. The page had the type checker and nothing
// else: enough to know that a name exists, never enough to know that a hook was
// told the truth about what it depends on.

import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'

export default tseslint.config(
    {
        ignores: ['dist/**', 'wailsjs/**', 'node_modules/**', 'coverage/**'],
    },
    js.configs.recommended,
    ...tseslint.configs.recommended,
    {
        files: ['**/*.{ts,tsx}'],
        plugins: {'react-hooks': reactHooks},
        rules: {
            // An effect that reads a value it never declared runs once against
            // whatever that value happened to be, then never again. The rule
            // leaves refs alone by design, which GlobeView leans on: its layout
            // reads the events through a ref so a zoom never rebuilds the effect.
            'react-hooks/exhaustive-deps': 'error',
            'react-hooks/rules-of-hooks': 'error',

            // An unused name is either a leftover or a mistake about what a
            // function was given. Arguments deliberately ignored say so with a
            // leading underscore.
            '@typescript-eslint/no-unused-vars': [
                'error',
                {argsIgnorePattern: '^_', varsIgnorePattern: '^_'},
            ],

            // any is how a typed codebase stops being one, quietly and in one
            // file at a time.
            '@typescript-eslint/no-explicit-any': 'error',

            // == against null is the one coercion worth keeping, because it
            // catches undefined as well and every line that wants it means both.
            eqeqeq: ['error', 'always', {null: 'ignore'}],
        },
    },
)
