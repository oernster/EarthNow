// The members of a cluster the camera cannot separate (FR-MRK-011, FR-MRK-012):
// events at one place (else too close to part even at the closest zoom), listed
// so each can be opened. The shared Dialog gives it the ring, Escape and Close.
import {clusterTitle, eventTitle} from '../categories'
import type {EventDTO} from '../types'
import {Dialog} from './Dialog'

interface Props {
    members: EventDTO[]
    onChoose: (e: EventDTO) => void
    onClose: () => void
}

export function ClusterList({members, onChoose, onClose}: Props) {
    return <Dialog title={clusterTitle(members)} onClose={onClose}>
        <ul className="cluster-list">
            {members.map(e => <li key={e.id}>
                <button data-stop className="cluster-member" onClick={() => onChoose(e)}>
                    {eventTitle(e)} <span className="muted">{e.reported}</span>
                </button>
            </li>)}
        </ul>
    </Dialog>
}
