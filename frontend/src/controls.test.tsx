// The rotation button and the settings dialog's rotation switch (FR-GLB-010,
// NFR-UX-004, FR-SET-001).
import {fireEvent, render, screen} from '@testing-library/react'
import {describe, expect, it, vi} from 'vitest'
import {Rail} from './components/Rail'
import {SettingsDialog} from './components/SettingsDialog'
import {icons} from './icons'
import type {SettingChoicesDTO, SettingsDTO} from './types'

const noop = () => undefined

function rotationButton(autoRotate: boolean) {
    render(<Rail autoRotate={autoRotate} attention={false} onToggleRotate={noop} onResetView={noop} onZoom={noop}
        onRefresh={noop} onStatus={noop} onSettings={noop} onHelp={noop} onDonate={noop}/>)
    return document.querySelector<HTMLElement>('.rail-btn')!
}

describe('FR-GLB-010 and NFR-UX-004 the rotation button shows what it switches to', () => {
    it.each([
        [true, 'Stop rotating', icons.rotateStop],
        [false, 'Start rotating', icons.rotate],
    ])('while rotation is %s it offers %s', (autoRotate, label, icon) => {
        const button = rotationButton(autoRotate)
        expect(button.getAttribute('aria-label')).toBe(label)
        expect(button.getAttribute('data-tip')).toBe(label)
        expect(button.querySelector('img')!.getAttribute('src')).toBe(icon)
    })

    it('draws the two states with different artwork', () => {
        expect(icons.rotateStop).not.toBe(icons.rotate)
    })
})

describe('FR-SET-001 the settings dialog', () => {
    it('offers auto-rotate on or off and applies a change at once', () => {
        const onChange = vi.fn()
        const settings = {autoRotate: true, speed: 'normal', magnitude: '2.5'} as SettingsDTO
        const choices = {speeds: [], magnitudes: []} as unknown as SettingChoicesDTO
        render(<SettingsDialog settings={settings} choices={choices} onChange={onChange} onClose={noop}/>)
        const toggle = screen.getByRole('checkbox', {name: 'Rotate the globe when idle'}) as HTMLInputElement
        expect(toggle.checked).toBe(true)
        fireEvent.click(toggle)
        expect(onChange).toHaveBeenCalledWith({autoRotate: false})
    })
})
