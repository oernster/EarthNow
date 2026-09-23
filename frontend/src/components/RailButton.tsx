// One rail button (FR-RAIL-001, FR-RAIL-003), in the house nav-band style ported
// from PigeonPost's TitleBar: artwork in a square box, named by an immediate
// tooltip that opens to its right and by the same accessible name. Every button
// on the rail is this one, so none can drift in size.
import {forwardRef, type ButtonHTMLAttributes} from 'react'

type Props = {
    label: string
    icon: string
    // attention marks a button with something waiting to be read.
    attention?: boolean
} & Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'>

export const RailButton = forwardRef<HTMLButtonElement, Props>(function RailButton(
    {label, icon, attention = false, className, ...rest}, ref) {
    const classes = ['rail-btn', attention ? 'attention' : '', className ?? ''].filter(Boolean).join(' ')
    return <button ref={ref} data-stop className={classes}
        data-tip={label} aria-label={label} {...rest}>
        <img src={icon} alt="" draggable={false}/>
    </button>
})
