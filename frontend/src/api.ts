// The one door to the Go side. Every call takes a refusal handler as its last
// argument and answers null rather than rejecting, so a call without one does
// not compile and nothing is left for a console nobody opens (NFR-REL-004).
import type {AboutDTO, ChoiceDTO, SettingChoicesDTO, SettingsDTO, ViewDTO} from './types'

type Refused = (reason: string) => void

interface Bound {
    View(windowKey: string, hiddenCategories: string[], hiddenProviders: string[]): Promise<ViewDTO>
    Windows(): Promise<ChoiceDTO[]>
    Settings(): Promise<SettingsDTO>
    SettingChoices(): Promise<SettingChoicesDTO>
    SaveSettings(chosen: SettingsDTO): Promise<SettingsDTO>
    Place(lat: number, lng: number): Promise<string>
    OpenSource(link: string): Promise<void>
    RefreshNow(): Promise<string>
    About(): Promise<AboutDTO>
    Licence(): Promise<string>
    Notices(): Promise<string>
    Donate(): Promise<void>
}

interface Runtime {
    EventsOn(name: string, callback: (...data: unknown[]) => void): () => void
}

function bound(): Bound | null {
    return (window as unknown as {go?: {main?: {App?: Bound}}}).go?.main?.App ?? null
}

async function call<T>(work: (b: Bound) => Promise<T>, onRefused: Refused): Promise<T | null> {
    const b = bound()
    if (!b) {
        onRefused('The EarthNow backend is not connected.')
        return null
    }
    try {
        return await work(b)
    } catch (e) {
        onRefused(e instanceof Error ? e.message : String(e))
        return null
    }
}

export const api = {
    view: (windowKey: string, hiddenCategories: string[], hiddenProviders: string[], onRefused: Refused) =>
        call(b => b.View(windowKey, hiddenCategories, hiddenProviders), onRefused),
    windows: (onRefused: Refused) => call(b => b.Windows(), onRefused),
    settings: (onRefused: Refused) => call(b => b.Settings(), onRefused),
    settingChoices: (onRefused: Refused) => call(b => b.SettingChoices(), onRefused),
    // saveSettings answers the settings as now held, which may differ from those sent.
    saveSettings: (chosen: SettingsDTO, onRefused: Refused) => call(b => b.SaveSettings(chosen), onRefused),
    place: (lat: number, lng: number, onRefused: Refused) => call(b => b.Place(lat, lng), onRefused),
    openSource: (link: string, onRefused: Refused) => call(b => b.OpenSource(link), onRefused),
    // refreshNow answers "" when a refresh started, else when one becomes available.
    refreshNow: (onRefused: Refused) => call(b => b.RefreshNow(), onRefused),
    about: (onRefused: Refused) => call(b => b.About(), onRefused),
    licence: (onRefused: Refused) => call(b => b.Licence(), onRefused),
    notices: (onRefused: Refused) => call(b => b.Notices(), onRefused),
    // donate hands the donation page to the system browser. The page never holds
    // the address; its one home is internal/product (FR-DON-004).
    donate: (onRefused: Refused) => call(b => b.Donate(), onRefused),
}

// on subscribes to a backend event; the answer unsubscribes. Outside the app
// window there is no runtime, so nothing is subscribed.
export function on(name: string, callback: (...data: unknown[]) => void): () => void {
    const runtime = (window as unknown as {runtime?: Runtime}).runtime
    return runtime ? runtime.EventsOn(name, callback) : () => undefined
}
