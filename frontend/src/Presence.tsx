import './Presence.css'
import { useAppEvent } from './api'
import { usePresence } from './hooks/query'

function Body() {
  const pr = usePresence()
  useAppEvent('sv.PresenceChange', () => {
    pr.refetch()
  })

  if (pr.type !== 'ok') {
    return pr.alt
  }

  const { clan, name, ok, iconURI } = pr.v
  if (!ok) {
    return null
  }

  return (
    <>
      {iconURI ? <img className="presence-icon" src={iconURI} /> : null}
      <span className="presence-clan">{clan}</span>
      <span className="presence-name">{name}</span>
    </>
  )
}

export default function Presence() {
  return (
    <div className="presence-ctr">
      <Body />
    </div>
  )
}
