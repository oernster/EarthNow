// The donate action (FR-DON-003, FR-DON-006). The page asks the Go side to open
// the donation page; the address never reaches the page. Free stays free: no
// feature depends on a donation (FR-DON-009).
import {api} from './api'

/** donate opens the donation page; a refusal is said, never swallowed. */
export function donate(onProblem: (reason: string) => void) {
    void api.donate(reason => onProblem(`The donation page could not be opened: ${reason}`))
}
