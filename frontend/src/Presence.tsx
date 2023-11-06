import './Presence.css'
import { usePresence } from './hooks/query'

function Body() {
  const pr = usePresence()
  if (pr.type !== 'ok') {
    return pr.alt
  }

  const { clan, name, ok, gameIconURI } = pr.v
  if (!ok) {
    return null
  }

  return (
    <>
      {gameIconURI ? <img className="presence-icon" src={gameIconURI} /> : null}
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
