import {render, screen} from '@testing-library/react'
import {afterEach, describe, expect, it, vi} from 'vitest'
import {GlobeView, NO_WEBGL2} from './components/GlobeView'

const noop = () => undefined

describe('FR-GLB-009 without WebGL2', () => {
    afterEach(() => { vi.restoreAllMocks() })

    it('states WebGL2 as the missing requirement in place of the globe', () => {
        vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)

        render(<GlobeView events={[]} selectedId={null} autoRotate={false} secondsPerRevolution={1} cloudImage="" dayNightShown={false} sun={null} start={{found: false, lat: 0, lng: 0}}
            onSelect={noop} onProblem={noop}/>)

        expect(screen.getByRole('alert').textContent).toBe(NO_WEBGL2)
        expect(NO_WEBGL2).toContain('WebGL2')
        expect(document.querySelector('canvas')).toBeNull()
    })
})
