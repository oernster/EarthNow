// NFR-KBD-007 (the house noborderfocus rule): a region is never ringed and
// never takes focus from a click. Two guards, each proved by planting the
// defect back: a scan of the stylesheet and a check of the click handler.
import {readFileSync} from 'node:fs'
import {join} from 'node:path'
import {describe, expect, it} from 'vitest'
import {fireEvent, render} from '@testing-library/react'
import {noClickFocus} from './ring'

// Read from disk: under Vitest a '?raw' CSS import is an empty string
// (measured), which made the first version of this scan check nothing. Vitest
// runs from frontend/ (npm test, test.ps1); under jsdom import.meta.url is not
// a file URL, so the path is taken from there.
const css = readFileSync(join(process.cwd(), 'src', 'style.css'), 'utf8')

// REGIONS are everything that holds or shows rather than being pressed: the
// panes, the bars, the globe, the dialog frame. Controls inside them (rail
// buttons, key rows, segments) ring as controls do.
const REGIONS = [
    '*', 'html', 'body', 'div', 'main', 'aside', 'nav', 'section', 'fieldset',
    '.app', '.rail', '.rail-actions', '.stage', '.globe', '.top-bar', '.time-window',
    '.status', '.detail', '.side', '.key', '.key-sources', '.scrim', '.dialog', '.tip',
]

// RING_PROPERTY is a declaration that draws something round the element.
// The lookahead also refuses whitespace, so the \s* before it cannot back off
// and let "outline: none" through as though it drew something.
const RING_PROPERTY = /(^|;|\s)(border(-color)?|outline|box-shadow)\s*:\s*(?!\s|none|transparent|0\b)/

interface Rule { selectors: string[]; body: string }

function rules(sheet: string): Rule[] {
    const bare = sheet.replace(/\/\*[\s\S]*?\*\//g, '')
    const out: Rule[] = []
    for (const match of bare.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
        out.push({selectors: match[1].split(',').map(s => s.trim()), body: match[2]})
    }
    return out
}

/** subject answers the element a selector's last compound names, pseudo parts stripped. */
function subject(selector: string): string {
    const last = selector.split(/[\s>+~]+/).filter(Boolean).pop() ?? ''
    const base = last.split(':')[0]
    const firstClass = base.match(/^[a-z*]*\.?[\w-]*/i)?.[0] ?? base
    return firstClass === '' ? base : firstClass
}

export function regionRings(sheet: string): string[] {
    const found: string[] = []
    for (const {selectors, body} of rules(sheet)) {
        if (!RING_PROPERTY.test(body)) continue
        for (const s of selectors) {
            if (/:(focus|hover)/.test(s) && REGIONS.includes(subject(s))) found.push(s)
        }
    }
    return found
}

describe('NFR-KBD-007 no region is ringed', () => {
    it('finds no focus or hover ring on a region in the stylesheet', () => {
        expect(css).toContain('.globe')
        expect(regionRings(css)).toEqual([])
    })

    it('would catch one: the rule removed from the globe is refused', () => {
        const planted = '.globe:focus-visible::after { border: 2px solid var(--ring); }'
        expect(regionRings(css + planted)).toEqual(['.globe:focus-visible::after'])
    })

    it('does not mistake removing a ring for drawing one', () => {
        expect(regionRings('.globe:focus { outline: none; }')).toEqual([])
    })

    it('leaves a control ring alone', () => {
        expect(regionRings('button:enabled:focus-visible { border-color: green; }')).toEqual([])
    })
})

// Every component source, read as text, so the check covers the real markup.
const sources = import.meta.glob('./components/*.tsx', {query: '?raw', import: 'default', eager: true}) as Record<string, string>

describe('NFR-KBD-007 no region takes focus from a click', () => {
    it('refuses the default action of mousedown, which is what moves focus', () => {
        const {container} = render(<div tabIndex={-1} onMouseDown={noClickFocus}/>)
        const pressed = fireEvent.mouseDown(container.firstElementChild as HTMLElement)
        expect(pressed).toBe(false)
    })

    // A div given tabIndex so the ring can focus it is a region by construction;
    // a click would focus it too unless it carries the handler.
    it('guards every focusable region in the components', () => {
        const unguarded: string[] = []
        for (const [file, text] of Object.entries(sources)) {
            for (const tag of text.matchAll(/<div\b[^>]*tabIndex=\{-1\}[^>]*>/g)) {
                if (!tag[0].includes('onMouseDown={noClickFocus}')) unguarded.push(file)
            }
        }
        expect(Object.keys(sources).length).toBeGreaterThan(0)
        expect(unguarded).toEqual([])
    })
})
