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
    band: number
    sourceUrl: string
    sourceText: string
    ended: boolean
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

export interface SettingsDTO {
    autoRotate: boolean
    magnitude: string
    speed: string
    windowKey: string
    hiddenCategories: string[]
    hiddenProviders: string[]
    cloudsShown: boolean
    dayNightShown: boolean
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
}
