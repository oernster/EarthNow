// The wire, stated a second time: these mirror internal/application/dto/dto.go
// field for field. A structural test compares the two (NFR-MNT-003).

export interface EventDTO {
    id: string
    provider: string
    category: string
    title: string
    description: string
    lat: number
    lng: number
    at: string
    dayOnly: boolean
    reported: string
    retrievedAt: string
    retrieved: string
    measurement: string
    // The Depth row, worded on the Go side (FR-SEL-010 to 012); empty when none.
    depth: string
    // A storm's positions inside the window as [lat, lng], oldest first, ending
    // at the marker (FR-TRL-001); empty for any other event.
    trail: [number, number][]
    band: number
    sourceUrl: string
    sourceText: string
    ended: boolean
    // An event in progress from a report (FR-PRV-015): reported names its
    // report week and no exact time is shown (FR-SEL-004).
    ongoing: boolean
}

export interface ProviderDTO {
    name: string
    loading: boolean
    // refreshing: a fetch is running with events already held (FR-STS-007).
    refreshing: boolean
    stale: boolean
    retrieved: string
    problem: string
    nextAttempt: string
    // A standing fact about what the provider holds, not a failure: its latest
    // report is too old to show (FR-PRV-016). Empty when none.
    notice: string
}

export interface ViewDTO {
    windowKey: string
    countLine: string
    events: EventDTO[]
    counts: Record<string, number>
    providers: ProviderDTO[]
    notice: string
}

export interface ChoiceDTO {
    key: string
    label: string
}

export interface SpeedDTO {
    key: string
    label: string
    secondsPerRevolution: number
}

// One replay speed (FR-RPL-025) with how long one pass of the span takes.
export interface ReplaySpeedDTO {
    key: string
    label: string
    passSeconds: number
}

export interface SettingsDTO {
    autoRotate: boolean
    magnitude: string
    speed: string
    windowKey: string
    hiddenCategories: string[]
    hiddenProviders: string[]
    cloudsShown: boolean
    dayNightShown: boolean
    trailsShown: boolean
    burntShown: boolean
    replaySpeed: string
}

// The burnt-area layer's state (FR-BA-008, FR-BA-009, FR-BA-017). The image is
// asked for apart; key changes whenever the image would.
export interface BurntAreasDTO {
    shown: boolean
    line: string
    key: string
    notice: string
    // GWIS's popover entry, filled only while shown (FR-BA-012).
    provider: ProviderDTO
}

// The cloud layer's state (FR-CLD-009, FR-CLD-013). The image is asked for
// apart; validTime changes when the held image is replaced.
export interface CloudsDTO {
    shown: boolean
    line: string
    validTime: string
    notice: string
    // The cloud service's popover entry, filled only while shown. It is kept
    // out of ViewDTO.providers, which are the event sources the key filters.
    provider: ProviderDTO
}

// Where the sun stands overhead now (FR-DAY-001), with FR-DAY-002's twilight
// limit in degrees and FR-DAY-009's share of cloud opacity kept at night.
export interface SunDTO {
    lat: number
    lng: number
    twilightDegrees: number
    cloudNightFloor: number
}

// Everything a replay shows at one instant (FR-RPL-009 to 019). The images
// are asked for apart, by cloudTime and burntKey, only when those change.
export interface ReplayFrameDTO {
    view: ViewDTO
    at: string
    line: string
    sun: SunDTO
    cloudTime: string
    cloudsLine: string
    cloudsProvider: ProviderDTO
    burntKey: string
}

// Where the globe opens (FR-GLB-015): facing lat, lng when found; otherwise
// as it always has (FR-GLB-016).
export interface StartViewDTO {
    found: boolean
    lat: number
    lng: number
}

export interface AboutDTO {
    name: string
    version: string
    copyright: string
    licence: string
    attributions: string[]
}

export interface SettingChoicesDTO {
    magnitudes: ChoiceDTO[]
    speeds: SpeedDTO[]
    replaySpeeds: ReplaySpeedDTO[]
}
