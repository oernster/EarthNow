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
    band: number
    sourceUrl: string
    sourceText: string
}

export interface ProviderDTO {
    name: string
    loading: boolean
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

export interface WindowDTO {
    key: string
    label: string
}
