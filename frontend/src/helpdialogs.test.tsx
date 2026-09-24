// The help dialogs against the real notices file and their reading bodies
// (FR-HLP-003, FR-HLP-005).
import {render, screen} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import notices from '../../THIRD_PARTY_NOTICES?raw'
import {HELP_ENTRIES, HelpDialog} from './components/HelpDialogs'

const noop = () => undefined

describe('the help dialogs', () => {
    const bound = {
        About: () => Promise.resolve({name: 'Product', version: '1.2.3', copyright: '©', licence: 'a licence', attributions: []}),
        Licence: () => Promise.resolve('the licence text'),
        Notices: () => Promise.resolve(notices),
    }
    beforeEach(() => { (window as unknown as {go: unknown}).go = {main: {App: bound}} })
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    // Every entry of the file the application ships, whole, licence texts and all.
    it('FR-HLP-003 shows every entry of THIRD_PARTY_NOTICES', async () => {
        render(<HelpDialog kind="notices" onClose={noop} onProblem={noop}/>)
        await screen.findByText((_, el) => el?.tagName === 'PRE')
        expect(document.querySelector('pre')!.textContent).toBe(notices)
        expect(notices).toContain('Natural Earth')
    })

    // Each dialog puts its content in the self-reading body the auto-scroll
    // cycle drives; the cycle itself is autoScroll.test.ts's.
    it.each(HELP_ENTRIES.map(e => e.kind))('FR-HLP-005 the %s dialog reads itself', kind => {
        const {unmount} = render(<HelpDialog kind={kind} onClose={noop} onProblem={noop}/>)
        expect(screen.getByRole('dialog').querySelector('.reading-body[data-reading]')).not.toBeNull()
        unmount()
    })
})
