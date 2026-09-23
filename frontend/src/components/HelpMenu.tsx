// The Help button and its menu (FR-HLP, Appendix D.2 help-info). The button is a
// rail stop; its menu is a popup in the keeb model: Down, Enter or Space drops it
// open onto its first item, Up and Down walk the items wrapping, Enter or Space
// chooses, Escape closes back to the button and Tab or the horizontal arrows
// leave it for the ring, which steps on from the Help button (ring.indexOnRing).
import {useEffect, useId, useRef, useState} from 'react'
import {HELP_ENTRIES, type HelpKind} from './HelpDialogs'
import {RailButton} from './RailButton'

interface Props {
    icon: string
    onChoose: (kind: HelpKind) => void
}

export function HelpMenu({icon, onChoose}: Props) {
    const [open, setOpen] = useState(false)
    const wrap = useRef<HTMLDivElement>(null)
    const button = useRef<HTMLButtonElement>(null)
    const menu = useRef<HTMLDivElement>(null)
    const menuId = useId()

    const items = () => Array.from(menu.current?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])

    useEffect(() => { if (open) items()[0]?.focus() }, [open])

    // A press anywhere else closes it; the globe refuses click focus, so a blur
    // alone would not see a press there.
    useEffect(() => {
        if (!open) return
        const onDown = (e: MouseEvent) => { if (!wrap.current?.contains(e.target as Node)) setOpen(false) }
        document.addEventListener('mousedown', onDown)
        return () => document.removeEventListener('mousedown', onDown)
    }, [open])

    // Focus goes back to the button before the menu unmounts, so the dialog a
    // choice opens hands focus back to it on close (NFR-KBD-006).
    const choose = (kind: HelpKind) => {
        button.current?.focus()
        setOpen(false)
        onChoose(kind)
    }

    const onButtonKey = (e: React.KeyboardEvent) => {
        if (e.key === 'ArrowDown') { e.preventDefault(); setOpen(true) }
        else if (e.key === 'ArrowUp') { e.preventDefault(); setOpen(false) }
    }

    const onMenuKey = (e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
            // One press closes one level: the menu, not a panel behind it.
            e.preventDefault()
            e.stopPropagation()
            button.current?.focus()
            setOpen(false)
        } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
            e.preventDefault()
            const list = items()
            const at = list.indexOf(document.activeElement as HTMLElement)
            list[(at + (e.key === 'ArrowDown' ? 1 : -1) + list.length) % list.length]?.focus()
        }
    }

    // Focus leaving the menu and its button (Tab, an arrow, a click) closes it.
    const onBlur = (e: React.FocusEvent) => {
        if (!wrap.current?.contains(e.relatedTarget as Node | null)) setOpen(false)
    }

    return <div ref={wrap} className="rail-menu" onBlur={onBlur}>
        <RailButton ref={button} label="Help" icon={icon} onClick={() => setOpen(o => !o)} onKeyDown={onButtonKey}
            aria-haspopup="menu" aria-expanded={open} aria-controls={menuId}/>
        {open && <div ref={menu} id={menuId} className="menu" role="menu" aria-label="Help" onKeyDown={onMenuKey}>
            {HELP_ENTRIES.map(entry => <button key={entry.kind} role="menuitem" tabIndex={-1}
                onClick={() => choose(entry.kind)}>{entry.label}</button>)}
        </div>}
    </div>
}
